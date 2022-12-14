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
				return
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
						Body: &ast.BlockStmt{
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
								&ast.ReturnStmt{
									Expr:     nil,
									Position: 55,
								},
							},
							Position: 35,
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
				pipe("echo «hello ${name}»")
				let input = read("\\w+")
				if x ~~ ".+" {
				  // ...
				} else if x !~ "hi" {
				  // ...
				} else {
				  // ...
				}
				curl() | jq() | head()
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
						Body: &ast.BlockStmt{
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
											Name:     "pipe",
											Position: IgnorePos,
										},
										Arg: &ast.ParenExpr{
											Exprs: []ast.Expr{
												&ast.String{
													Root:      parser.MustParseStr(`echo «hello ${name}»`, false, true),
													StringLit: `"echo «hello ${name}»"`,
													Position:  IgnorePos,
												},
											},
											Position: IgnorePos,
										},
										PipedArg: &ast.ParenExpr{
											Exprs:    nil,
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
													Root:      parser.MustParseStr(`\w+`, false, true),
													StringLit: `"\\w+"`,
													Position:  IgnorePos,
												},
											},
											Position: IgnorePos,
										},
										PipedArg: &ast.ParenExpr{
											Exprs:    nil,
											Position: IgnorePos,
										},
										Position: IgnorePos,
									},
									Position: IgnorePos,
								},
								&ast.IfStmt{
									Cond: &ast.BinaryExpr{
										X: &ast.Ident{Name: "x", Position: IgnorePos},
										Y: &ast.String{
											Root:      parser.MustParseStr(`.+`, false, true),
											StringLit: `".+"`,
											Position:  IgnorePos,
										},
										Op:       token.REG,
										Position: IgnorePos,
									},
									Body: &ast.BlockStmt{
										Statements: nil,
										Position:   IgnorePos,
									},
									Else: &ast.IfStmt{
										Cond: &ast.BinaryExpr{
											X: &ast.Ident{Name: "x", Position: IgnorePos},
											Y: &ast.String{
												Root:      parser.MustParseStr(`hi`, false, true),
												StringLit: `"hi"`,
												Position:  IgnorePos,
											},
											Op:       token.NREG,
											Position: IgnorePos,
										},
										Body: &ast.BlockStmt{
											Statements: nil,
											Position:   IgnorePos,
										},
										Else: &ast.BlockStmt{
											Statements: nil,
											Position:   IgnorePos,
										},
										Position: IgnorePos,
									},
									Position: IgnorePos,
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.Ident{Name: "head", Position: 228},
										Arg: &ast.ParenExpr{Position: 232},
										PipedArg: &ast.ParenExpr{Exprs: []ast.Expr{
											&ast.CallExpr{
												Fun: &ast.Ident{Name: "jq", Position: 221},
												Arg: &ast.ParenExpr{Position: 223},
												PipedArg: &ast.ParenExpr{Exprs: []ast.Expr{
													&ast.CallExpr{
														Fun:      &ast.Ident{Name: "curl", Position: 212},
														Arg:      &ast.ParenExpr{Position: 216},
														PipedArg: &ast.ParenExpr{},
														Position: 212,
													},
												}},
												Position: 221,
											},
										}},
										Position: 228,
									},
									Position: 212,
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
