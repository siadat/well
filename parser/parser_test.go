package parser_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/parser"
)

func TestParser(tt *testing.T) {
	var testCases = []struct {
		src  string
		want parser.CmdNode
	}{
		{
			src: `ls  -lash --directory -C ./something`,
			want: parser.ContainerNode{
				Items: []parser.CmdNode{
					parser.Arg{`ls`},
					parser.Whs{`  `},
					parser.Arg{`-lash`},
					parser.Whs{` `},
					parser.Arg{`--directory`},
					parser.Whs{` `},
					parser.Arg{`-C`},
					parser.Whs{` `},
					parser.Arg{`./something`},
				},
			},
		},
		{
			src: `echo "Hello ${name}!"`,
			want: parser.ContainerNode{
				Items: []parser.CmdNode{
					parser.Arg{`echo`},
					parser.Whs{` `},
					parser.ContainerNode{
						Items: []parser.CmdNode{
							parser.Arg{`Hello`},
							parser.Whs{` `},
							parser.Var{`name`, ``},
							parser.Arg{`!`},
						},
					},
				},
			},
		},
		{
			src: `one ${var} three`,
			want: parser.ContainerNode{
				Items: []parser.CmdNode{
					parser.Arg{`one`},
					parser.Var{`var`, ``},
					parser.Arg{`three`},
				},
			},
		},
		{
			src: `one ${var:%q} three`,
			want: parser.ContainerNode{
				Items: []parser.CmdNode{
					parser.Arg{`one`},
					parser.Var{`var`, `%q`},
					parser.Arg{`three`},
				},
			},
		},
		{
			src: `one «${var:%q} «this is \«three\»» four» end`,
			want: parser.ContainerNode{
				Items: []parser.CmdNode{
					parser.Arg{`one`},
					parser.Whs{` `},
					parser.ContainerNode{
						Items: []parser.CmdNode{
							parser.Var{`var`, `%q`},
							parser.Whs{` `},
							parser.ContainerNode{
								Items: []parser.CmdNode{
									parser.Arg{`this is «three»`},
								},
							},
							parser.Whs{` `},
							parser.Arg{`four`},
						},
					},
					parser.Whs{` `},
					parser.Arg{`end`},
				},
			},
		},
	}

	for _, tc := range testCases {
		var p = parser.NewParser()
		var src = tc.src
		var got, err = p.Parse(strings.NewReader(src))
		if err != nil {
			tt.Fatalf("test case failed src=%q: %v", src, err)
		}
		if diff := cmp.Diff(tc.want, got); diff != "" {
			tt.Errorf("case failed src=%q (-want +got):\n%s", src, diff)
		}
	}
}
