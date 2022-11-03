package types_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/types"
)

const IgnorePos = -1

var testCases = []struct {
	src  string
	want map[ast.Expr]types.Type
}{
	{
		src: `
		function main() {
	        let x = "hello"
	        let y = 3.14
	        let z = 123
			# external(x)
	    }
	    `,
		want: map[ast.Expr]types.Type{
			ast.Ident{Name: "main", Position: 12}: types.WellType{"Function"},

			ast.Ident{Name: "x", Position: 34}: types.WellType{"String"},
			ast.Ident{Name: "y", Position: 59}: types.WellType{"Float"},
			ast.Ident{Name: "z", Position: 81}: types.WellType{"Integer"},

			ast.Integer{Value: 123, Position: 85}:                        types.WellType{"Integer"},
			ast.Float{Value: 3.14, Position: 63}:                         types.WellType{"Float"},
			ast.String{Root: parser.MustParseStr("hello"), Position: 38}: types.WellType{"String"},

			// ast.Ident{Name: "external", Position: 92}: types.WellType{"Function"},
			// ast.Ident{Name: "x", Position: 101}:       types.WellType{"String"},
		},
	},
}

func TestParser(tt *testing.T) {
	for ti, tc := range testCases {
		var src = tc.src
		src = formatSrc(src, true)

		checker := types.NewChecker()
		checker.SetDebug(true)
		var gotResult, err = checker.Check(strings.NewReader(tc.src))
		if err != nil {
			tt.Fatalf("check failed (test case %d)\nsrc:\n%s\nerr:\n%s", ti, src, err)
		}

		// I have to convert the maps to slice, because the behavior of
		// "github.com/google/go-cmp/cmp" is strange for maps. I cannot ignore
		// positions, and the comparison is not deep (pointers are
		// dereferences, instead the pointer itself is compared.
		var note = `Note: These are maps, converted to slices because of how cmp works. See comments.`
		var wantSlice = mapToSlice(tc.want)
		var gotSlice = mapToSlice(gotResult)

		var cmpOpt = cmp.FilterValues(func(p1, p2 scanner.Pos) bool { return p1 == IgnorePos || p2 == IgnorePos || p1 == p2 }, cmp.Ignore())

		if diff := cmp.Diff(wantSlice, gotSlice, cmpOpt); diff != "" {
			fmt.Printf("got: %# v\n", pretty.Formatter(gotResult))
			tt.Fatalf("mismatching results (test case %d)\nsrc:\n%s\nnote: %s\ndiff guide:\n  - want\n  + got\ndiff:\n%s", ti, src, note, diff)
		}
	}
}

func mapToSlice(m map[ast.Expr]types.Type) []KeyValue {
	var ret []KeyValue

	for k, v := range m {
		ret = append(ret, KeyValue{k, v})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Key.Pos() < ret[j].Key.Pos() })
	return ret
}

type KeyValue struct {
	Key   ast.Expr
	Value types.Type
}

func formatSrc(src string, showWhitespaces bool) string {
	var prefix = "   | "
	if showWhitespaces {
		// src = strings.ReplaceAll(src, " ", "₋")
		src = strings.ReplaceAll(src, "\t", "␣")
		src = strings.Join(strings.Split(src, "\n"), "⏎\n"+prefix)
		src = prefix + src + "·"
		return src
	}
	return src
}
