package expander

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/siadat/well/syntax/strs/parser"
	"github.com/siadat/well/syntax/strs/scanner"
)

var WhitespaceRe = regexp.MustCompile(`\s`)

func MappingFuncFromMap(m map[string]interface{}) func(string) interface{} {
	return func(name string) interface{} {
		var v, ok = m[name]
		if !ok {
			return nil
		}
		return v
	}
}

type QuotingVariant int

const (
	// Just use backslash to escape ' and "
	Basic QuotingVariant = iota
	// BashAnsiCVariant returns $'it\'s great' instead of 'it\'s great'.
	// The dollar sign is required for bash, otherwise it won't work.
	// https://stackoverflow.com/questions/6697753/difference-between-single-and-double-quotes-in-bash/42082956#42082956
	// TODO: This mode must also replace ` with \`, but so far I decided not to support it.
	// TODO: Remove this variant. I don't want to support Bash syntax. Anything you do with Bash should be possible with Well.
	BashAnsiCVariant
)

var quotingVariant = Basic // BashAnsiCVariant

// TODO: this is a toy implemntation, please rewrite, see also: strconv.Quote
func EscapeQuote(s string, typ scanner.CmdTokenType) (string, error) {
	switch typ {
	case scanner.LDOUBLE_GUILLEMET, scanner.DOUBLE_QUOTE:
		return fmt.Sprintf("%q", s), nil
	case scanner.LSINGLE_GUILLEMET, scanner.SINGLE_QUOTE:
		return EscapeSinglequote(s), nil
	default:
		return "", fmt.Errorf("unsupported quote %s", typ)
	}
}

func EscapeSinglequote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	switch quotingVariant {
	case BashAnsiCVariant:
		return fmt.Sprintf("$'%s'", s)
	case Basic:
		return fmt.Sprintf("'%s'", s)
	default:
		panic(fmt.Sprintf("unsupported quotingVariant %d", quotingVariant))
	}
}

func ParseAndEncodeToString(src string, mapping func(string) interface{}, debug bool) (string, error) {
	var p = parser.NewParser()
	p.SetDebug(debug)
	var node, parseErr = p.Parse(strings.NewReader(src))
	if parseErr != nil {
		return "", parseErr
	}

	var s, convertErr = convertToExecNode(node, true, mapping)
	if convertErr != nil {
		return "", convertErr
	}
	return s.Value(), nil
}

func EncodeToString(root *parser.Root, mapping func(string) interface{}) (string, error) {
	var s, err = convertToExecNode(root, true, mapping)
	if err != nil {
		return "", err
	}
	return s.Value(), nil
}

type Variable struct {
	Name string
	Type string
}

func GetVariables(src string) ([]Variable, error) {
	var p = parser.NewParser()
	var root, parseErr = p.Parse(strings.NewReader(src))
	if parseErr != nil {
		return nil, parseErr
	}

	var variables []Variable
	var f = func(name, opts string) {
		var typ string
		switch opts {
		case "":
			typ = "string"
		case "%s":
			typ = "string"
		case "%q":
			typ = "string"
		case "%Q":
			typ = "string"
		case "%d":
			typ = "int"
		}
		variables = append(variables, Variable{name, typ})
	}
	var err = findVars(root, f)
	if err != nil {
		return nil, err
	}
	return variables, nil
}

// TODO: refactor arg, varg, args
// TODO: remove Root, and replace all uses with ContainerNode

func EncodeToCmdArgs(root *parser.Root, mapping func(string) interface{}) ([]string, error) {
	var args []string
	// Arguments are splited by whitespace, ie any 2 parsed nodes that have no
	// space between them should be joined as 1 argument.
	// currArgBuf concatinates every parsed node in the root that is not split with
	// whitespaces. This is necessary, because for example the following input:
	//    ." hello world "
	// is parsed as (.) and (" hello world ") and we need to join the two, because
	// there's no space between them, and return (." hello world ") as 1 arg.
	var currArgBuf bytes.Buffer

	var growArg = func(fragment string) {
		currArgBuf.WriteString(fragment)
	}

	var commitArg = func() {
		if currArgBuf.String() == "" {
			// nothing to commit
			return
		}
		args = append(args, currArgBuf.String())
		currArgBuf.Reset()
	}

	for _, item := range root.Items {
		var arg, err = convertToExecNode(item, false, mapping)
		if err != nil {
			return nil, err
		}
		switch arg.(type) {
		case ExecVar:
			var words = WhitespaceRe.Split(arg.Value(), -1)
			for i, w := range words {
				growArg(w)
				if i < len(words)-1 {
					commitArg()
				}
			}
		case ExecWhs:
			commitArg()
		default:
			growArg(arg.Value())
		}
	}

	// last arg
	if currArgBuf.String() != "" {
		commitArg()
	}

	return args, nil
}

func findVars(node parser.CmdNode, onFound func(string, string)) error {
	switch item := node.(type) {
	case *parser.Root:
		for _, item := range item.Items {
			var err = findVars(item, onFound)
			if err != nil {
				return err
			}
		}
		return nil
	case parser.ContainerNode:
		for _, item := range item.Items {
			var err = findVars(item, onFound)
			if err != nil {
				return err
			}
		}
		return nil
	case parser.Whs:
		return nil
	case parser.Wrd:
		return nil
	case parser.Var:
		onFound(item.Name, item.Opts)
		return nil
	default:
		panic(fmt.Sprintf("unsupported encoding for node type %T", item))
	}
}

func convertToExecNode(node parser.CmdNode, escapeOuter bool, mapping func(string) interface{}) (ExecNode, error) {
	// fmt.Printf("[---] convertToExecNode: %#v (escapeOuter=%v)\n", node, escapeOuter)
	switch item := node.(type) {
	case *parser.Root:
		var args []string
		for _, item := range item.Items {
			var arg, err = convertToExecNode(item, escapeOuter, mapping)
			if err != nil {
				return nil, err
			}
			args = append(args, arg.Value())
		}
		return ExecWrd{strings.Join(args, "")}, nil
	case parser.ContainerNode:
		var args []string
		for _, item := range item.Items {
			var arg, err = convertToExecNode(item, true, mapping)
			if err != nil {
				return nil, err
			}
			args = append(args, arg.Value())
		}

		var s = strings.Join(args, "")
		if !escapeOuter {
			return ExecWrd{Lit: s}, nil
		} else {
			var s, err = EscapeQuote(s, item.Type)
			if err != nil {
				panic(err)
			}
			return ExecWrd{Lit: s}, nil
		}
	case parser.Whs:
		return ExecWhs{Lit: item.Lit}, nil
	case parser.Wrd:
		return ExecWrd{Lit: item.Lit}, nil
	case parser.Var:
		var val = mapping(item.Name)
		if val == nil {
			return nil, fmt.Errorf("variable %s is %v", item.Name, val)
		}
		return varFormatter(val, item.Opts, escapeOuter)
	default:
		panic(fmt.Sprintf("unsupported encoding for node type %T", item))
	}
}

func varFormatter(v interface{}, flags string, escapeOuter bool) (ExecNode, error) {
	// fmt.Printf("[===] varFormatter:%#v flags=%q\n", v, flags)
	switch flags {
	case "":
		return ExecVar{Lit: fmt.Sprintf("%s", v)}, nil
	case "%s":
		return ExecVar{Lit: fmt.Sprintf("%s", v)}, nil
	case "%f":
		return ExecVar{Lit: fmt.Sprintf("%f", v)}, nil
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
		return ExecVar{Lit: fmt.Sprintf("%s", v)}, nil
	default:
		panic(fmt.Sprintf("unsupported variable flags %q", flags))
	}
}
