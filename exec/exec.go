package exec

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/siadat/well/parser"
	"github.com/siadat/well/scanner"
)

func MappingFuncFromMap(m map[string]interface{}) func(string) interface{} {
	return func(name string) interface{} {
		var v, ok = m[name]
		if !ok {
			panic(fmt.Sprintf("missing value for variable %q", name))
		}
		return v
	}
}

type SingleQuotingVariant int

const (
	// Just use backslash to escape inner single-quotes
	Basic SingleQuotingVariant = iota
	// BashAnsiCVariant returns $'it\'s great' instead of 'it\'s great'.
	// The dollar sign is required for bash, otherwise it won't work.
	// https://stackoverflow.com/questions/6697753/difference-between-single-and-double-quotes-in-bash/42082956#42082956
	BashAnsiCVariant
)

var singleQuotingVariant = BashAnsiCVariant

// TODO: this is a toy implemntation, please rewrite, see also: strconv.Quote
func SingleQuoteEscaper(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	switch singleQuotingVariant {
	case BashAnsiCVariant:
		return fmt.Sprintf("$'%s'", s)
	case Basic:
		return fmt.Sprintf("'%s'", s)
	default:
		panic(fmt.Sprintf("unsupported singleQuotingVariant %d", singleQuotingVariant))
	}
}

func EncodeToString(src string, mapping func(string) interface{}) string {
	var p = parser.NewParser()
	var node, err = p.Parse(strings.NewReader(src))
	if err != nil {
		panic(fmt.Sprintf("test case failed src=%q: %v", src, err))
	}
	var s = encode(node, false, true, mapping)
	return strings.Join(pairLits(s), "")
}

// TODO: refactor arg, varg, args, Pair

func EncodeToCmdArgs(root *parser.Root, mapping func(string) interface{}) []string {
	var args []string
	// Arguments are splited by whitespace, ie any 2 parsed nodes that have no
	// space between them should be joined as 1 argument.
	// argBuf concatinates every parsed node in the root that is not split with
	// whitespaces. This is necessary, because for example the following input:
	//    ." hello world "
	// is parsed as (.) and (" hello world ") and we need to join the two, because
	// there's no space between them, and return (." hello world ") as 1 arg.
	var argBuf bytes.Buffer
	for _, item := range root.Items {
		var arg = encode(item, true, false, mapping)
		for _, varg := range arg {
			fmt.Printf("%s\n", varg)
			if varg.IsWhitespace {
				args = append(args, argBuf.String())
				argBuf.Reset()
			} else {
				argBuf.WriteString(varg.Lit)
			}
		}
	}

	// last arg
	if argBuf.String() != "" {
		args = append(args, argBuf.String())
	}

	return args
}

type Pair struct {
	Lit          string
	IsWhitespace bool
}

func (p Pair) String() string {
	return fmt.Sprintf("Pair(%q, %v)", p.Lit, p.IsWhitespace)
}

func pairLits(pairs []Pair) []string {
	var list []string
	for _, v := range pairs {
		list = append(list, v.Lit)
	}
	return list
}

func encode(node parser.CmdNode, ignoreWhitespace bool, escapeOuter bool, mapping func(string) interface{}) []Pair {
	switch item := node.(type) {
	case *parser.Root:
		var args []Pair
		for _, item := range item.Items {
			var arg = encode(item, ignoreWhitespace, escapeOuter, mapping)
			for _, varg := range arg {
				if !(ignoreWhitespace && varg.IsWhitespace) {
					args = append(args, arg...)
				}
			}
		}
		// return strings.Join(args, ""), false
		return args
	case parser.ContainerNode:
		var args []Pair
		for _, item := range item.Items {
			var arg = encode(item, false, true, mapping)
			for _, varg := range arg {
				args = append(args, varg)
			}
		}
		if !escapeOuter {
			return []Pair{{Lit: strings.Join(pairLits(args), ""), IsWhitespace: false}}
		} else {
			var joined = strings.Join(pairLits(args), "")
			switch item.Type {
			case scanner.LDOUBLE_GUILLEMET, scanner.DOUBLE_QUOTE:
				return []Pair{{Lit: fmt.Sprintf("%q", joined), IsWhitespace: false}}
			case scanner.LSINGLE_GUILLEMET:
				return []Pair{{Lit: SingleQuoteEscaper(joined), IsWhitespace: false}}
			case scanner.SINGLE_QUOTE:
				return []Pair{{Lit: SingleQuoteEscaper(joined), IsWhitespace: false}}
			default:
				panic(fmt.Sprintf("unsupported container %s", item.Type))
			}
		}
	case parser.Whs:
		return []Pair{{Lit: item.Lit, IsWhitespace: true}}
	case parser.Var:
		return formatter(mapping(item.Name), item.Opts)
	case parser.Wrd:
		return []Pair{{Lit: item.Lit, IsWhitespace: false}}
	default:
		panic(fmt.Sprintf("unsupported encoding for node type %T", item))
	}
}

func formatter(v interface{}, flags string) []Pair {
	switch flags {
	case "":
		return []Pair{{Lit: fmt.Sprintf("%s", v), IsWhitespace: false}}
	case "%s":
		return []Pair{{Lit: fmt.Sprintf("%s", v), IsWhitespace: false}}
	case "%f":
		return []Pair{{Lit: fmt.Sprintf("%f", v), IsWhitespace: false}}
	case "%q":
		return []Pair{{Lit: fmt.Sprintf("%q", v), IsWhitespace: false}}
	case "%Q":
		return []Pair{{Lit: fmt.Sprintf("%Q", v), IsWhitespace: false}} // TODO: allow doing single-quote vs double-quote
	case "%-":
		var val = fmt.Sprintf("%s", v)
		var ws = regexp.MustCompile(`\s`)
		indexes := ws.FindAllStringIndex(val, -1)

		var args []Pair

		var last = 0
		for i := range indexes {
			var wsstart = indexes[i][0]
			var wsend = indexes[i][1]

			if last < wsstart {
				args = append(args, Pair{Lit: val[last:wsstart], IsWhitespace: false})
			}
			args = append(args, Pair{Lit: val[wsstart:wsend], IsWhitespace: true})
			last = wsend
		}

		if last < len(val) {
			args = append(args, Pair{Lit: val[last:], IsWhitespace: false})
		}

		return args
	default:
		panic(fmt.Sprintf("unsupported variable flags %q", flags))
	}
}
