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
)

const IgnorePos = -1

func TestParser(tt *testing.T) {
	var testCases = []struct {
		src  string
		want *ast.Root
	}{
		{
			src: `
			function main(s string, i int) {
				let x = 3
			}
			`,
			want: &ast.Root{
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name:     "main",
							Position: 13,
						},
						Signature: &ast.FuncSignature{
							Args: []ast.FuncSignatureArg{
								{
									Name: "s",
									Type: "string",
								},
								{
									Name: "i",
									Type: "int",
								},
							},
							RetTypes: nil,
							Position: 17,
						},
						Statements: []ast.Stmt{
							&ast.LetDecl{
								Name: &ast.Ident{
									Name:     "x",
									Position: 45,
								},
								Rhs: &ast.Integer{
									Value:    3,
									Position: 49,
								},
								Position: 41,
							},
						},
						Position: 4,
					},
				},
			},
		},
		{
			src: `
			function main() string {
				let x = 3
				external("echo «hello ${name}»")
				let input = read("\\w+")
				return input
			}
			`,
			want: &ast.Root{
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name:     "main",
							Position: 13,
						},
						Signature: &ast.FuncSignature{
							Args:     nil,
							RetTypes: []string{"string"},
							Position: IgnorePos,
						},
						Statements: []ast.Stmt{
							&ast.LetDecl{
								Name: &ast.Ident{
									Name:     "x",
									Position: IgnorePos,
								},
								Rhs: &ast.Integer{
									Value:    3,
									Position: IgnorePos,
								},
								Position: IgnorePos,
							},
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.Ident{
										Name:     "external",
										Position: IgnorePos,
									},
									Arg: &ast.ParenExpr{
										Exprs: []ast.Expr{
											&ast.String{
												Root:     parser.MustParseStr(`echo «hello ${name}»`, false, true),
												Position: IgnorePos,
											},
										},
										Position: IgnorePos,
									},
									Position: IgnorePos,
								},
								Position: IgnorePos,
							},
							&ast.LetDecl{
								Name: &ast.Ident{
									Name:     "input",
									Position: IgnorePos,
								},
								Rhs: &ast.CallExpr{
									Fun: &ast.Ident{
										Name:     "read",
										Position: IgnorePos,
									},
									Arg: &ast.ParenExpr{
										Exprs: []ast.Expr{
											&ast.String{
												Root:     parser.MustParseStr(`\w+`, false, true),
												Position: IgnorePos,
											},
										},
										Position: IgnorePos,
									},
									Position: IgnorePos,
								},
								Position: IgnorePos,
							},
							&ast.ReturnStmt{
								Expr: &ast.Ident{
									Name:     "input",
									Position: IgnorePos,
								},
								Position: IgnorePos,
							},
						},
						Position: IgnorePos,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var p = parser.NewParser()
		p.SetDebug(true)
		var src = tc.src
		var got, err = p.Parse(strings.NewReader(src))
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
