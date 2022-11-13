package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/parser"
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/syntax/token"
)

func TestParseExpr(tt *testing.T) {
	var testCases = []struct {
		src  string
		want ast.Expr
	}{
		{
			src: `1 + 2 * 3 * 4 + 5`,
			want: &ast.BinaryExpr{
				X: &ast.Integer{Value: 1},
				Y: &ast.BinaryExpr{
					X: &ast.BinaryExpr{
						X: &ast.Integer{Value: 2, Position: IgnorePos},
						Y: &ast.BinaryExpr{
							X:        &ast.Integer{Value: 3, Position: IgnorePos},
							Y:        &ast.Integer{Value: 4, Position: IgnorePos},
							Op:       token.MUL,
							Position: IgnorePos,
						},
						Op:       token.MUL,
						Position: IgnorePos,
					},
					Y:        &ast.Integer{Value: 5, Position: IgnorePos},
					Op:       token.ADD,
					Position: IgnorePos,
				},
				Op:       token.ADD,
				Position: IgnorePos,
			},
		},
		{
			src: `1 * 2 + 3 + 4 * 5`,
			want: &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:        &ast.Integer{Value: 1},
					Y:        &ast.Integer{Value: 2, Position: IgnorePos},
					Op:       token.MUL,
					Position: IgnorePos,
				},
				Y: &ast.BinaryExpr{
					X: &ast.Integer{Value: 3, Position: IgnorePos},
					Y: &ast.BinaryExpr{
						X:        &ast.Integer{Value: 4, Position: IgnorePos},
						Y:        &ast.Integer{Value: 5, Position: IgnorePos},
						Op:       token.MUL,
						Position: IgnorePos,
					},
					Op:       token.ADD,
					Position: IgnorePos,
				},
				Op:       token.ADD,
				Position: IgnorePos,
			},
		},
	}

	for _, tc := range testCases {
		var p = parser.NewParser()
		p.SetDebug(true)
		var src = tc.src
		var got, err = p.ParseExpr(strings.NewReader(src))
		src = scanner.FormatSrc(src, true)
		if err != nil {
			tt.Fatalf("test case failed\nsrc:\n%s\nerr:\n%s", src, err)
		}

		var cmpOpt = cmp.FilterValues(func(p1, p2 scanner.Pos) bool { return p1 == IgnorePos || p2 == IgnorePos || p1 == p2 }, cmp.Ignore())

		if diff := cmp.Diff(tc.want, got, cmpOpt); diff != "" {
			fmt.Printf("got: %# v\n", pretty.Formatter(got))
			tt.Fatalf("mismatching results\nsrc:\n%s\ndiff guide:\n  - want\n  + got\ndiff:\n%s", src, diff)
		}
	}
}
