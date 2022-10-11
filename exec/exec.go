package exec

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/siadat/well/parser"
	"github.com/siadat/well/scanner"
)

func formatter(v interface{}, flags string) string {
	switch flags {
	case "":
		return fmt.Sprintf("%s", v)
	case "%s":
		return fmt.Sprintf("%s", v)
	case "%f":
		return fmt.Sprintf("%f", v)
	case "%q":
		return fmt.Sprintf("%q", v)
	case "%Q":
		return fmt.Sprintf("%Q", v) // TODO: allow doing single-quote vs double-quote
	default:
		panic(fmt.Sprintf("unsupported variable flags %q", flags))
	}
}

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
	var s, _ = encode(node, false, true, mapping)
	return s
}

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
		var arg, isWhitespace = encode(item, true, false, mapping)
		// fmt.Printf("%s %q %v\n", "arg", arg, isWhitespace)
		if isWhitespace {
			args = append(args, argBuf.String())
			argBuf.Reset()
		} else {
			argBuf.WriteString(arg)
		}
	}

	// last arg
	if argBuf.String() != "" {
		args = append(args, argBuf.String())
	}

	return args
}

func encode(node parser.CmdNode, ignoreWhitespace bool, escapeOuter bool, mapping func(string) interface{}) (string, bool) {
	switch item := node.(type) {
	case *parser.Root:
		var args []string
		for _, item := range item.Items {
			var arg, isWhitespace = encode(item, ignoreWhitespace, escapeOuter, mapping)
			if !(ignoreWhitespace && isWhitespace) {
				args = append(args, arg)
			}
		}
		return strings.Join(args, ""), false
	case parser.ContainerNode:
		var args []string
		for _, item := range item.Items {
			var arg, _ = encode(item, false, true, mapping)
			args = append(args, arg)
		}
		if !escapeOuter {
			return strings.Join(args, ""), false
		} else {
			var joined = strings.Join(args, "")
			switch item.Type {
			case scanner.LDOUBLE_GUILLEMET, scanner.DOUBLE_QUOTE:
				return fmt.Sprintf("%q", joined), false
			case scanner.LSINGLE_GUILLEMET:
				return SingleQuoteEscaper(joined), false
			case scanner.SINGLE_QUOTE:
				return SingleQuoteEscaper(joined), false
			default:
				panic(fmt.Sprintf("unsupported container %s", item.Type))
			}
		}
	case parser.Whs:
		return item.Lit, true
		// if ignoreWhitespace {
		// 	return "", true
		// } else {
		// 	return item.Lit, true
		// }

		// if ignoreWhitespace {
		// 	return "", false
		// } else {
		// 	return item.Lit, true
		// }
	case parser.Var:
		return formatter(mapping(item.Name), item.Opts), false
	case parser.Wrd:
		return item.Lit, false
	default:
		panic(fmt.Sprintf("unsupported encoding for node type %T", item))
	}
}
