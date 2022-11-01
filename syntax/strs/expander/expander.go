package expander

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/siadat/well/syntax/strs/parser"
	"github.com/siadat/well/syntax/strs/scanner"
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

func EncodeToString(src string, mapping func(string) interface{}, debug bool) (string, error) {
	var p = parser.NewParser()
	p.SetDebug(debug)
	var node, err = p.Parse(strings.NewReader(src))
	if err != nil {
		return "", err
	}

	var s = convertToExecNode(node, true, mapping)
	return s.Value(), nil
}

func EncodeToString2(root *parser.Root, mapping func(string) interface{}) (string, error) {
	var s = convertToExecNode(root, true, mapping)
	return s.Value(), nil
}

func MustEncodeToString(src string, mapping func(string) interface{}) string {
	var s, err = EncodeToString(src, mapping, false)
	if err != nil {
		panic(fmt.Sprintf("test case failed src=%q: %v", src, err))
	}

	return s
}

// TODO: refactor arg, varg, args
// TODO: remove Root, and replace all uses with ContainerNode

func EncodeToCmdArgs(root *parser.Root, mapping func(string) interface{}) []string {
	var ws = regexp.MustCompile(`\s`)
	var args []string
	// Arguments are splited by whitespace, ie any 2 parsed nodes that have no
	// space between them should be joined as 1 argument.
	// buf concatinates every parsed node in the root that is not split with
	// whitespaces. This is necessary, because for example the following input:
	//    ." hello world "
	// is parsed as (.) and (" hello world ") and we need to join the two, because
	// there's no space between them, and return (." hello world ") as 1 arg.
	var buf bytes.Buffer

	var fillarg = func(fragment string) {
		buf.WriteString(fragment)
	}
	var closearg = func() {
		args = append(args, buf.String())
		buf.Reset()
	}

	for _, item := range root.Items {
		var arg = convertToExecNode(item, false, mapping)
		switch arg.(type) {
		case ExecVar:
			var words = ws.Split(arg.Value(), -1)
			for i, w := range words {
				fillarg(w)
				if i < len(words)-1 {
					closearg()
				}
			}
		case ExecWhs:
			closearg()
		default:
			fillarg(arg.Value())
		}
	}

	// last arg
	if buf.String() != "" {
		closearg()
	}

	return args
}

func convertToExecNode(node parser.CmdNode, escapeOuter bool, mapping func(string) interface{}) ExecNode {
	// fmt.Printf("node:%#v\n", node)
	switch item := node.(type) {
	case *parser.Root:
		var args []string
		for _, item := range item.Items {
			var arg = convertToExecNode(item, escapeOuter, mapping)
			args = append(args, arg.Value())
		}
		return ExecWrd{strings.Join(args, "")}
	case parser.ContainerNode:
		var args []string
		for _, item := range item.Items {
			var arg = convertToExecNode(item, true, mapping)
			args = append(args, arg.Value())
		}

		if !escapeOuter {
			return ExecWrd{Lit: strings.Join(args, "")}
		} else {
			var joined = strings.Join(args, "")
			switch item.Type {
			case scanner.LDOUBLE_GUILLEMET, scanner.DOUBLE_QUOTE:
				return ExecWrd{Lit: fmt.Sprintf("%q", joined)}
			case scanner.LSINGLE_GUILLEMET:
				return ExecWrd{Lit: SingleQuoteEscaper(joined)}
			case scanner.SINGLE_QUOTE:
				return ExecWrd{Lit: SingleQuoteEscaper(joined)}
			default:
				panic(fmt.Sprintf("unsupported container %s", item.Type))
			}
		}
	case parser.Whs:
		return ExecWhs{Lit: item.Lit}
	case parser.Wrd:
		return ExecWrd{Lit: item.Lit}
	case parser.Var:
		return formatterNewNew(mapping(item.Name), item.Opts, escapeOuter)
	default:
		panic(fmt.Sprintf("unsupported encoding for node type %T", item))
	}
}

func formatterNewNew(v interface{}, flags string, escapeOuter bool) ExecNode {
	// fmt.Printf("formatterNewNew:%#v %q\n", v, flags)
	switch flags {
	case "":
		return ExecVar{Lit: fmt.Sprintf("%s", v)}
	case "%s":
		return ExecVar{Lit: fmt.Sprintf("%s", v)}
	case "%f":
		return ExecVar{Lit: fmt.Sprintf("%f", v)}
	case "%q":
		return convertToExecNode(
			parser.ContainerNode{
				Type: scanner.DOUBLE_QUOTE,
				Items: []parser.CmdNode{
					parser.Wrd{Lit: fmt.Sprintf("%s", v)},
				},
			}, escapeOuter, nil)
	case "%Q":
		return convertToExecNode(
			parser.ContainerNode{
				Type: scanner.SINGLE_QUOTE,
				Items: []parser.CmdNode{
					parser.Wrd{Lit: fmt.Sprintf("%s", v)},
				},
			}, escapeOuter, nil)
	case "%-":
		return ExecVar{Lit: fmt.Sprintf("%s", v)}
	default:
		panic(fmt.Sprintf("unsupported variable flags %q", flags))
	}
}

func formatterNew(v interface{}, flags string) []parser.CmdNode {
	switch flags {
	case "":
		return []parser.CmdNode{parser.Wrd{Lit: fmt.Sprintf("%s", v)}}
	case "%s":
		return []parser.CmdNode{parser.Wrd{Lit: fmt.Sprintf("%s", v)}}
	case "%f":
		return []parser.CmdNode{parser.Wrd{Lit: fmt.Sprintf("%f", v)}}
	case "%q":
		return []parser.CmdNode{parser.Wrd{Lit: fmt.Sprintf("%q", v)}}
	case "%Q":
		return []parser.CmdNode{parser.Wrd{Lit: fmt.Sprintf("%Q", v)}} // TODO: allow doing single-quote vs double-quote
	case "%-":
		var val = fmt.Sprintf("%s", v)
		var ws = regexp.MustCompile(`\s`)
		indexes := ws.FindAllStringIndex(val, -1)

		var args []parser.CmdNode

		var last = 0
		for i := range indexes {
			var wsstart = indexes[i][0]
			var wsend = indexes[i][1]

			if last < wsstart {
				args = append(args, parser.Wrd{Lit: val[last:wsstart]})
			}
			args = append(args, parser.Whs{Lit: val[wsstart:wsend]})
			last = wsend
		}

		if last < len(val) {
			args = append(args, parser.Wrd{Lit: val[last:]})
		}

		return args
	default:
		panic(fmt.Sprintf("unsupported variable flags %q", flags))
	}
}
