package parser_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/syntax/strs/parser"
	"github.com/siadat/well/syntax/strs/scanner"
)

func TestParser(tt *testing.T) {
	var testCases = []struct {
		src  string
		want *parser.Root
	}{
		{
			src: `ls  -lash --directory -C ./something`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`ls`},
					parser.Whs{`  `},
					parser.Wrd{`-lash`},
					parser.Whs{` `},
					parser.Wrd{`--directory`},
					parser.Whs{` `},
					parser.Wrd{`-C`},
					parser.Whs{` `},
					parser.Wrd{`./something`},
				},
			},
		},
		{
			src: `echo ${nameA} ${nameB}`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`echo`},
					parser.Whs{` `},
					parser.Var{`nameA`, ``},
					parser.Whs{` `},
					parser.Var{`nameB`, ``},
				},
			},
		},
		{
			src: `echo "Hello ${name}!"`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`echo`},
					parser.Whs{` `},
					parser.ContainerNode{
						Type: scanner.DOUBLE_QUOTE,
						Items: []parser.CmdNode{
							parser.Wrd{`Hello`},
							parser.Whs{` `},
							parser.Var{`name`, ``},
							parser.Wrd{`!`},
						},
					},
				},
			},
		},
		{
			src: `jq «.${key:%q} | .»`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`jq`},
					parser.Whs{` `},
					parser.ContainerNode{
						Type: scanner.LDOUBLE_GUILLEMET,
						Items: []parser.CmdNode{
							parser.Wrd{`.`},
							parser.Var{`key`, `%q`},
							parser.Whs{` `},
							parser.Wrd{`|`},
							parser.Whs{` `},
							parser.Wrd{`.`},
						},
					},
				},
			},
		},
		{
			src: `one ${var} three`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`one`},
					parser.Whs{` `},
					parser.Var{`var`, ``},
					parser.Whs{` `},
					parser.Wrd{`three`},
				},
			},
		},
		{
			src: `one ${var:%q} three`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`one`},
					parser.Whs{` `},
					parser.Var{`var`, `%q`},
					parser.Whs{` `},
					parser.Wrd{`three`},
				},
			},
		},
		{
			src: `one «${var:%q} «this is \«three\»» \$5 dolloars» end`,
			want: &parser.Root{
				[]parser.CmdNode{
					parser.Wrd{`one`},
					parser.Whs{` `},
					parser.ContainerNode{
						Type: scanner.LDOUBLE_GUILLEMET,
						Items: []parser.CmdNode{
							parser.Var{`var`, `%q`},
							parser.Whs{` `},
							parser.ContainerNode{
								Type: scanner.LDOUBLE_GUILLEMET,
								Items: []parser.CmdNode{
									parser.Wrd{`this`},
									parser.Whs{` `},
									parser.Wrd{`is`},
									parser.Whs{` `},
									parser.Wrd{`«`},
									parser.Wrd{`three`},
									parser.Wrd{`»`},
								},
							},
							parser.Whs{` `},
							parser.Wrd{`$`},
							parser.Wrd{`5`},
							parser.Whs{` `},
							parser.Wrd{`dolloars`},
						},
					},
					parser.Whs{` `},
					parser.Wrd{`end`},
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
			tt.Fatalf("case failed src=%q (-want +got):\n%s", src, diff)
		}
	}
}
