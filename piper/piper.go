package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/siadat/well/syntax/strs/expander"
	"github.com/siadat/well/syntax/strs/parser"
)

const SigpipeErrorMessage = "signal: broken pipe"

type external struct {
	pipeline []string
}

func (ext *external) Read() string {
	var ctx = context.TODO()

	var cmds = make([]*exec.Cmd, len(ext.pipeline))
	var stdin_closers = make([]io.WriteCloser, len(ext.pipeline))
	var stdouts = make([]bytes.Buffer, len(ext.pipeline))
	var stderrs = make([]bytes.Buffer, len(ext.pipeline))

	var all_words = make([][]string, len(ext.pipeline))

	for i, cmdStr := range ext.pipeline {

		var p = parser.NewParser()
		var node, err = p.Parse(strings.NewReader(cmdStr))
		if err != nil {
			panic(fmt.Sprintf("parsing command failed str=%q: %v", cmdStr, err))
		}
		var words, encodeErr = expander.EncodeToCmdArgs(node, nil)
		if encodeErr != nil {
			panic(fmt.Sprintf("parsing command failed str=%q: %v", cmdStr, err))
		}
		// var words = strings.SplitN(cmdStr, " ", -1)

		if len(words) < 1 {
			panic(fmt.Sprintf("expected at least 1 word in command; got in %d", len(words)))
		}
		all_words[i] = words

		stdin_closers[i] = os.Stdin
		cmds[i] = exec.CommandContext(ctx, words[0], words[1:]...)
		cmds[i].Stdout = &stdouts[i]
		cmds[i].Stderr = &stderrs[i]
		if i == 0 {
			cmds[i].Stdin = os.Stdin
			// fmt.Printf("[debug] connecting stdin of %v to os.Stdin\n", cmds[i])
		}
		if i > 0 {
			var wc, pipe_err2 = cmds[i].StdinPipe()
			if pipe_err2 != nil {
				panic(pipe_err2)
			}
			stdin_closers[i] = wc
			cmds[i-1].Stdout = io.MultiWriter(wc, &stdouts[i-1])
			// fmt.Printf("[debug] connecting stdout of %v to stdin of %v\n", cmds[i-1], cmds[i])
		}
	}

	var wg sync.WaitGroup
	for i, cmd := range cmds {
		wg.Add(1)
		go func(i int, cmd *exec.Cmd) {
			defer func() {
				// fmt.Println("[debug] finished", cmd)
				wg.Done()
				if i != len(cmds)-1 {
					// TODO: what if it is already closed?
					defer stdin_closers[i+1].Close()
				}
			}()

			// fmt.Println("[debug] running", cmd)
			if err := cmd.Run(); err != nil {
				if err != nil {
					// TODO: nicer errors, eg yaml
					if strings.Contains(err.Error(), SigpipeErrorMessage) {
						// Fine, no-op :)
						// broken pipe just means that the stdin of the process we
						// were piping to is closed. That's fine, because that
						// process might have finished its job. E.g. in `yes | head`
						// head exits faster than yes.
						// fmt.Println("[debug] SIGPIPE received by", cmd)
					} else {
						panic(fmt.Sprintf("command %q failed: %v", cmd, err))
					}
				}
			}
		}(i, cmd)
	}
	wg.Wait()
	var last_stdout = stdouts[len(stdouts)-1]
	// fmt.Printf("[debug] returning stdout of %v\n", cmds[len(stdouts)-1])
	return last_stdout.String()
}

func (ext *external) External(cmd string) *external {
	return &external{
		pipeline: append(ext.pipeline, cmd),
	}
}

func External(cmd string) *external {
	return &external{
		pipeline: []string{cmd},
	}
}

func main() {
	var out = External("yes").
		External("nl").
		External("shuf").
		External("tail -f").
		Read()

	fmt.Print(out)
}
