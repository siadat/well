package parser_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	"github.com/siadat/well/syntax/strs/parser"
)

func TestNode(tt *testing.T) {
	var testCases = []struct {
		str string
	}{
		{
			str: ` abc  efg x${abc}
${abc:%q}  
    \$ literal dollar
	\› literal guillemet
	‹hello «1» ›


multiple newlines`,
		},
	}

	for _, tc := range testCases {
		var p = parser.NewParser()
		var src = tc.str
		var got, err = p.Parse(strings.NewReader(src))
		if err != nil {
			tt.Fatalf("test case failed src=%q: %v", src, err)
		}
		if diff := cmp.Diff(`"`+src+`"`, got.String()); diff != "" {
			pretty.Println("got:", got)
			tt.Fatalf("case failed src=%q (-want +got):\n%s", src, diff)
		}
	}
}
