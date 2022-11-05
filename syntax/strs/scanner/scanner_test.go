package scanner_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/syntax/strs/scanner"
)

func TestScanner(tt *testing.T) {
	var testCases = []struct {
		src  string
		want []scanner.CmdToken
	}{
		{
			src: `ls  -lash --directory   -C ./something`,
			want: []scanner.CmdToken{
				{scanner.WORD, `ls`},
				{scanner.SPACE, `  `},
				{scanner.WORD, `-lash`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `--directory`},
				{scanner.SPACE, `   `},
				{scanner.WORD, `-C`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `./something`},
			},
		},
		{
			src: `echo "Hello ${name}!"`,
			want: []scanner.CmdToken{
				{scanner.WORD, `echo`},
				{scanner.SPACE, ` `},
				{scanner.DOUBLE_QUOTE, `"`},
				{scanner.WORD, `Hello`},
				{scanner.SPACE, ` `},
				{scanner.ARG, `name`},
				{scanner.WORD, `!`},
				{scanner.DOUBLE_QUOTE, `"`},
			},
		},
		{
			src: `one ${var} three`,
			want: []scanner.CmdToken{
				{scanner.WORD, `one`},
				{scanner.SPACE, ` `},
				{scanner.ARG, `var`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `three`},
			},
		},
		{
			src: `one ${var:%q} three`,
			want: []scanner.CmdToken{
				{scanner.WORD, `one`},
				{scanner.SPACE, ` `},
				{scanner.ARG, `var:%q`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `three`},
			},
		},
		{
			src: `one «${var_123:%q} «this is \«three\»» four» end`,
			want: []scanner.CmdToken{
				{scanner.WORD, `one`},
				{scanner.SPACE, ` `},
				{scanner.LDOUBLE_GUILLEMET, `«`},
				{scanner.ARG, `var_123:%q`},
				{scanner.SPACE, ` `},
				{scanner.LDOUBLE_GUILLEMET, `«`},
				{scanner.WORD, `this`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `is`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `«`}, // NOTE: this is a WORD, because it was escaped
				{scanner.WORD, `three`},
				{scanner.WORD, `»`}, // NOTE: this is a WORD, because it was escaped
				{scanner.RDOUBLE_GUILLEMET, `»`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `four`},
				{scanner.RDOUBLE_GUILLEMET, `»`},
				{scanner.SPACE, ` `},
				{scanner.WORD, `end`},
			},
		},
	}

	for _, tc := range testCases {
		var src = tc.src
		var s = scanner.NewScanner(strings.NewReader(src))
		var got []scanner.CmdToken
		var errs []error
		for {
			var t, err = s.NextToken()
			if t.Typ == scanner.EOF {
				break
			}
			if err != nil {
				errs = append(errs, err)
			}
			got = append(got, t)
		}

		if diff := cmp.Diff(([]error)(nil), errs); diff != "" {
			tt.Fatalf("case error failed to match src=%q (-want +got):\n%s", src, diff)
		}
		if diff := cmp.Diff(tc.want, got); diff != "" {
			tt.Fatalf("case failed src=%q (-want +got):\n%s", src, diff)
		}
	}
}
