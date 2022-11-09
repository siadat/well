package fumt_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	"github.com/siadat/well/fumt"
	"github.com/siadat/well/syntax/scanner"
)

const IgnorePos = -1

var testCases = []struct {
	src  string
	want string
}{
	{
		src: `// external(x)

		let x = 1
		function main() {
	        let x = "hello"
	          let y = 3.14
	        let z = 123
			  f2(1,   external_capture (  "date ${x:%q}" 
			)   ,,, )

// external(x)

	    }


		  function f2( x   int   ) {
	          let y = 3.14
			    return
	    }
	    `,
		want: `let x = 1
function main() {
	let x = "hello"
	let y = 3.14
	let z = 123
	f2(1, external_capture("date ${x:%q}"))
}

function f2(x int) {
	let y = 3.14
	return
}

`,
	},
}

func TestFumt(tt *testing.T) {
	// TODO: preserve comments
	for ti, tc := range testCases {
		var src = tc.src
		src = scanner.FormatSrc(src, true)

		var formater = fumt.NewFormater()
		formater.SetDebug(true)

		var buf bytes.Buffer
		var err = formater.Format(strings.NewReader(tc.src), &buf)
		if err != nil {
			tt.Fatalf("check failed (test case %d)\nsrc:\n%s\nerr:\n%s", ti, src, err)
		}
		var got = buf.String()

		if diff := cmp.Diff(tc.want, got); diff != "" {
			fmt.Printf("got: %# v\n", pretty.Formatter(got))
			tt.Fatalf("mismatching results (test case %d)\nsrc:\n%s\ngot:\n%s\ndiff guide:\n  - want\n  + got\ndiff:\n%s", ti, src, scanner.FormatSrc(got, true), diff)
		}
	}
}
