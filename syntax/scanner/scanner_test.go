package scanner_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/syntax/scanner"
	"github.com/siadat/well/syntax/token"
)

const IgnorePos = -1

func TestIdentifier(tt *testing.T) {
	var testCases = []struct {
		skipWhitespace bool
		src            string
		want           []scanner.Token
	}{
		{
			src: `abc  "some string" xyz`,
			want: []scanner.Token{
				{token.IDENTIFIER, `abc`, 0},
				{token.WHITESPACE, `  `, 3},
				{token.STRING, `"some string"`, 5},
				{token.WHITESPACE, ` `, 18},
				{token.IDENTIFIER, `xyz`, 19},
			},
		},
		{
			skipWhitespace: true,
			src: `
			abc1, abc2,
			[ ( { - 123 + 1.23 / 0 } ) ]
			"echo «hi»"
			`,
			want: []scanner.Token{
				{token.NEWLINE, "\n", IgnorePos},
				{token.IDENTIFIER, `abc1`, IgnorePos},
				{token.COMMA, `,`, IgnorePos},
				{token.IDENTIFIER, `abc2`, IgnorePos},
				{token.COMMA, `,`, IgnorePos},

				{token.NEWLINE, "\n", IgnorePos},
				{token.LBRACK, `[`, IgnorePos},
				{token.LPAREN, `(`, IgnorePos},
				{token.LBRACE, `{`, IgnorePos},
				{token.SUB, `-`, IgnorePos},
				{token.INTEGER, `123`, IgnorePos},
				{token.ADD, `+`, IgnorePos},
				{token.FLOAT, `1.23`, IgnorePos},
				{token.QUO, `/`, IgnorePos},
				{token.INTEGER, `0`, IgnorePos},
				{token.RBRACE, `}`, IgnorePos},
				{token.RPAREN, `)`, IgnorePos},
				{token.RBRACK, `]`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},

				{token.STRING, `"echo «hi»"`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},
			},
		},
		{
			skipWhitespace: true,
			src: `
			func main(
			  verbose    bool,
			  first_name string,
			) {
			  let out, err = external("echo hello «${first_name}!»")
			}
			`,
			want: []scanner.Token{
				{token.NEWLINE, "\n", IgnorePos},
				{token.IDENTIFIER, `func`, IgnorePos},
				{token.IDENTIFIER, `main`, IgnorePos},
				{token.LPAREN, `(`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},

				{token.IDENTIFIER, `verbose`, IgnorePos},
				{token.IDENTIFIER, `bool`, IgnorePos},
				{token.COMMA, `,`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},

				{token.IDENTIFIER, `first_name`, IgnorePos},
				{token.IDENTIFIER, `string`, IgnorePos},
				{token.COMMA, `,`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},

				{token.RPAREN, `)`, IgnorePos},
				{token.LBRACE, `{`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},

				{token.IDENTIFIER, `let`, IgnorePos},
				{token.IDENTIFIER, `out`, IgnorePos},
				{token.COMMA, `,`, IgnorePos},
				{token.IDENTIFIER, `err`, IgnorePos},
				{token.ASSIGN, `=`, IgnorePos},
				{token.IDENTIFIER, `external`, IgnorePos},
				{token.LPAREN, `(`, IgnorePos},
				{token.STRING, `"echo hello «${first_name}!»"`, IgnorePos},
				{token.RPAREN, `)`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},

				{token.RBRACE, `}`, IgnorePos},
				{token.NEWLINE, "\n", IgnorePos},
			},
		},
	}

	for _, tc := range testCases {
		var src = tc.src
		var s = scanner.NewScanner(strings.NewReader(src))
		s.SetSkipWhitespace(tc.skipWhitespace)
		var got []scanner.Token
		var err error
		for {
			var t, gotErr = s.NextToken()
			if t.Typ == token.EOF {
				break
			}
			if gotErr != nil {
				err = gotErr
				break
			}
			got = append(got, t)
		}

		var cmpOpt = cmp.FilterValues(func(p1, p2 scanner.Pos) bool { return p1 == IgnorePos || p2 == IgnorePos || p1 == p2 }, cmp.Ignore())

		if diff := cmp.Diff((error)(nil), err); diff != "" {
			tt.Fatalf("case error failed to match src=%q\n-want\n+got\ndiff:\n%s", src, diff)
		}
		if diff := cmp.Diff(tc.want, got, cmpOpt); diff != "" {
			tt.Fatalf("case failed src=%q\n-want\n+got\ndiff:\n%s", src, diff)
		}
	}
}
