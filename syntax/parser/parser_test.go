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
			function main() {
				let x = 3
			}
			`,
			want: &ast.Root{
				Decls: []ast.Decl{
					ast.FuncDecl{
						Name: ast.Ident{
							Name:     "main",
							Position: 13,
						},
						Signature: ast.FuncSignature{
							ArgNames: nil,
							ArgTypes: nil,
							RetTypes: nil,
							Position: 17,
						},
						Statements: []ast.Stmt{
							ast.LetDecl{
								Name: ast.Ident{
									Name:     "x",
									Position: 30,
								},
								Rhs: ast.Integer{
									Value:    3,
									Position: 34,
								},
								Position: 26,
							},
						},
						Position: 4,
					},
				},
			},
		},
		{
			src: `
			function main() {
				let x = 3
				external("echo «hello ${name}»")
			}
			`,
			want: &ast.Root{
				Decls: []ast.Decl{
					ast.FuncDecl{
						Name: ast.Ident{
							Name:     "main",
							Position: 13,
						},
						Signature: ast.FuncSignature{
							ArgNames: nil,
							ArgTypes: nil,
							RetTypes: nil,
							Position: 17,
						},
						Statements: []ast.Stmt{
							ast.LetDecl{
								Name: ast.Ident{
									Name:     "x",
									Position: 30,
								},
								Rhs: ast.Integer{
									Value:    3,
									Position: 34,
								},
								Position: 26,
							},
							ast.ExprStmt{
								X: ast.CallExpr{
									Fun: ast.Ident{
										Name:     "external",
										Position: 40,
									},
									Arg: ast.ParenExpr{
										Exprs: []ast.Expr{
											ast.String{
												Root:     parser.MustParseStr(`echo «hello ${name}»`),
												Position: 49,
											},
										},
										Position: 48,
									},
									Position: 40,
								},
								Position: 40,
							},
						},
						Position: 4,
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
		src = formatSrc(src, true)
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
