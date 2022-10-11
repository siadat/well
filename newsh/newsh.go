package newsh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// This package contains function, types, and
// variables that act as builtin or stdlib for the
// new language (if I decide to create a new
// language). At the moment, it works like a DSL
// inside Go.

const SigpipeErrorMessage = "signal: broken pipe"

type Void struct{}

var Nothing = Void{}
var CmdCounter uint64 = 0

func Cd(dir File, f func()) {
	var backup = global_dir
	global_dir = dir.Path
	f()
	global_dir = backup
}

func Exit(str string) {
	fmt.Println(str)
	os.Exit(1)
}

func FileExists(f File) bool {
	if _, err := os.Stat(filepath.Join(global_dir, f.Path)); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		Exit(fmt.Sprintf("failed to check file stat: %v", err))
		return false
	}
}

func PrintInfo(str string) {
	fmt.Println(str)
}

type ValMap map[string]interface{}

func expand_str(str string, mapping func(string) string) string {
	var re = regexp.MustCompile(`\${\w+}|@+{\w+}`)
	return re.ReplaceAllStringFunc(str, func(s string) string {
		switch s[:2] {
		case "${":
			var name = s[2 : len(s)-1]
			return mapping(name)
		case "@{":
			var name = s[2 : len(s)-1]
			var value = mapping(name)
			return fmt.Sprintf("%q", value)
		case "@@":
			var name = s[3 : len(s)-1]
			var value = mapping(name)
			value = strings.ReplaceAll(value, `\`, `\\`)
			value = strings.ReplaceAll(value, `'`, `\'`)
			return `'` + value + `'`
		default:
			Exit(fmt.Sprintf("internal error: unknown match prefix in %q", s))
			return ""
		}
	})
}

func Interpolate(str string, env ValMap) string {
	return expand_str(str, func(s string) string {
		value, ok := env[s]
		if !ok {
			Exit(fmt.Sprintf("missing key %q in string", s))
		}
		switch v := value.(type) {
		case string:
			return v
		case int:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%f", v)
		case fmt.Stringer:
			return v.String()
		default:
			Exit(fmt.Sprintf("unsupported key type %T for %q", v, v))
			return ""
		}
	})
}

var global_dir string = "."

type Options struct {
	TrimSpaces bool
}

type CmdInfo struct {
	Time    string
	Message string
	Pwd     string
	Dir     string
	Cmd     string
	Pipe    []string
	Err     string
	Stdout  string
	Stderr  string
}

type CmdNode struct {
	Cmd  string
	Args []string
}

type Pipe []string

func truncate_string(str string) string {
	var max_bytes = 500
	var max_lines = 5

	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	var lines1 = 0
	var lines2 = 0

	var i1 = 0
	var i2 = 0

	for i1 = 0; i1 < len(str); i1++ {
		if str[i1] == '\n' {
			lines1 += 1
		}
		buf1.WriteByte(str[i1])
		if buf1.Len() > max_bytes {
			break
		}
		if lines1 > max_lines {
			break
		}
	}

	for i2 = len(str) - 1; i2 > i1; i2-- {
		if str[i2] == '\n' {
			lines2 += 1
		}
		buf2.WriteByte(str[i2])
		if buf2.Len() > max_bytes {
			break
		}
		if lines2 > max_lines {
			break
		}
	}

	var middle string
	if i2-i1 > 1 {
		middle = fmt.Sprintf("...(%d more bytes)...", i2-i1-1)
	}
	return buf1.String() + middle + Reverse(buf2.String())
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ExternalPiped(strs Pipe, opts ...Options) string {
	var ctx = context.TODO()
	var opt Options
	if len(opts) > 1 {
		panic(fmt.Sprintf("syntax error: expected at most 1 option, got %d", len(opts)))
	} else if len(opts) == 1 {
		opt = opts[0]
	}

	var cmds = make([]*exec.Cmd, len(strs))
	var stdin_closers = make([]io.WriteCloser, len(strs))
	var stdouts = make([]bytes.Buffer, len(strs))
	var stderrs = make([]bytes.Buffer, len(strs))
	var first_words []string
	var all_words = make([][]string, len(strs))
	for i := range cmds {
		var str = strs[i]
		var words = strings.SplitN(str, " ", -1)
		if len(words) < 1 {
			Exit(fmt.Sprintf("expected at least 1 word in command; got in %d", len(words)))
			return ""
		}
		all_words[i] = words
		first_words = append(first_words, words[0])
		ctx = context.WithValue(ctx, "words", words)

		stdin_closers[i] = os.Stdin
		cmds[i] = exec.CommandContext(ctx, words[0], words[1:]...)
		cmds[i].Dir = global_dir
		cmds[i].Stdout = &stdouts[i]
		cmds[i].Stderr = &stderrs[i]
		if i > 0 {
			var wc, pipe_err2 = cmds[i].StdinPipe()
			if pipe_err2 != nil {
				Exit(fmt.Sprintf("failed to create stdin pipe: %v", pipe_err2))
				return ""
			}
			stdin_closers[i] = wc
			cmds[i-1].Stdout = io.MultiWriter(wc, &stdouts[i-1])
		}
	}

	var cmdinfo_chan = make(chan CmdInfo)

	go func() {
		var wg sync.WaitGroup
		for i, cmd := range cmds {
			wg.Add(1)
			go func(i int, cmd *exec.Cmd) {
				// fmt.Println("started", cmd)
				defer func() {
					// fmt.Println("ended", cmd)
					wg.Done()
					if i != len(cmds)-1 {
						// TODO: what if it is already closed?
						defer stdin_closers[i+1].Close()
					}
				}()

				var cmd_err = cmd.Run()

				// log
				if true {
					var stdout = stdouts[i]
					var stderr = stderrs[i]

					var re_newline = regexp.MustCompile(`\r?\n`)
					// var re_tab = regexp.MustCompile(`\t`)

					var newout = stdout.String()
					newout = re_newline.ReplaceAllString(stdout.String(), "\n")
					// newout = re_tab.ReplaceAllString(stdout.String(), "    ")
					newout = strings.TrimSpace(newout)
					newout = truncate_string(newout)

					var newerr = stdout.String()
					newerr = re_newline.ReplaceAllString(stderr.String(), "\n")
					// newerr = re_tab.ReplaceAllString(stderr.String(), "    ")
					newerr = strings.TrimSpace(newerr)
					newerr = truncate_string(newerr)

					var pwd, pwd_err = os.Getwd()
					if pwd_err != nil {
						fmt.Fprintf(os.Stderr, "\nfailed to get pwd: %v\n", pwd_err)
						os.Exit(1)
					}
					var marked_pipe = make([]string, len(first_words))
					for k := range first_words {
						if k == i {
							marked_pipe[k] = fmt.Sprintf("%s (current)", first_words[k])
						} else {
							marked_pipe[k] = fmt.Sprintf("%s", first_words[k])
						}
					}
					var info_item = CmdInfo{
						Time:   time.Now().Format("2006-01-02 15:04:05.999 -07:00"),
						Cmd:    strs[i],
						Pwd:    pwd,
						Dir:    cmd.Dir,
						Pipe:   marked_pipe,
						Stdout: newout,
						Stderr: newerr,
					}
					if cmd_err != nil {
						info_item.Err = cmd_err.Error()
					}
					cmdinfo_chan <- info_item
				}

				if cmd_err != nil {
					// TODO: nicer errors, eg yaml
					if strings.Contains(cmd_err.Error(), SigpipeErrorMessage) {
						// Fine, no-op :)
						// broken pipe just means that the stdin of the process we
						// were piping to is closed. That's fine, because that
						// process might have finished its job. E.g. in `yes | head`
						// head exits faster than yes.
						fmt.Println("SIGPIPE received by", cmd)
					} else {
						Exit(fmt.Sprintf("command %q failed: %v", cmd, cmd_err))
					}
				}
			}(i, cmd)
		}
		wg.Wait()
		close(cmdinfo_chan)
	}()

	for info_item := range cmdinfo_chan {
		var enc = yaml.NewEncoder(os.Stderr)
		CmdCounter += 1
		var yaml_err = enc.Encode(map[uint64]CmdInfo{CmdCounter: info_item})
		if yaml_err != nil {
			fmt.Fprintf(os.Stderr, "\nyaml encoding failed with %v\n", yaml_err)
			os.Exit(1)
		}
	}

	// return the output of the last command
	var last_stdout = stdouts[len(stdouts)-1]
	if opt.TrimSpaces {
		return strings.TrimSpace(last_stdout.String())
	} else {
		return last_stdout.String()
	}
}

func External(str string, opts ...Options) string {
	return ExternalPiped(Pipe{str}, opts...)
}

type File struct {
	Path string
}

func (f File) String() string {
	return f.Path
}
