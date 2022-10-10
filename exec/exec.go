package exec

import (
	"fmt"
	"strings"

	"github.com/siadat/well/parser"
	"github.com/siadat/well/scanner"
)

func MappingFuncFromMap(m map[string]interface{}) func(string, string) string {
	return func(name, flags string) string {
		var v, ok = m[name]
		if !ok {
			panic(fmt.Sprintf("missing value for variable %q", name))
		}
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
}

func EncodeRoot(root *parser.Root, mapping func(string, string) string) []string {
	var args []string
	for _, item := range root.Items {
		var arg, ok = Encode(item, 0, mapping)
		if ok {
			args = append(args, arg)
		}
	}
	return args
}

func Encode(node parser.CmdNode, depth int, mapping func(string, string) string) (string, bool) {
	switch item := node.(type) {
	case parser.ContainerNode:
		var args []string
		for _, item := range item.Items {
			var arg, ok = Encode(item, depth+1, mapping)
			if ok {
				args = append(args, arg)
			}
		}
		if depth == 0 {
			return strings.Join(args, ""), true
		} else {
			var joined = strings.Join(args, "")
			switch item.Type {
			case scanner.LDOUBLE_GUILLEMET, scanner.DOUBLE_QUOTE:
				return fmt.Sprintf("%q", joined), true
			case scanner.LSINGLE_GUILLEMET:
				panic("TODO")
			case scanner.SINGLE_QUOTE:
				panic("TODO")
			default:
				panic(fmt.Sprintf("unsupported container %s", item.Type))
			}
		}
	case parser.Whs:
		if depth == 0 {
			return "", false
		} else {
			return item.Lit, true
		}
	case parser.Var:
		return mapping(item.Name, item.Opts), true
	case parser.Wrd:
		return item.Lit, true
	default:
		panic(fmt.Sprintf("unsupported encoding for node type %T", item))
	}
}
