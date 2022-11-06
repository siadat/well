package interpreter_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	"github.com/siadat/well/interpreter"
)

var testCases = []struct {
	src        string
	wantObj    interpreter.Object
	wantStdout string
}{
	{
		src: `
	    function main() {
	        external("echo 'hello'")
	    }
	    `,
		wantObj:    nil,
		wantStdout: "hello\n",
	},
	{
		src: `
	    function f(s1 string, s2 string) {
			println(s1, "and", s2)
		}

	    function main() {
	        external(
			  "echo 'hello'",
			  "nl",
			)
			f("hi", "bye")
	    }
	    `,
		wantObj:    nil,
		wantStdout: "     1\thello\nhi and bye\n",
	},
}

func TestParser(tt *testing.T) {
	for ti, tc := range testCases {
		var src = tc.src
		src = formatSrc(src, true)

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		interp := interpreter.NewInterpreter(&stdout, &stderr)
		interp.SetDebug(true)
		env := interpreter.NewEnvironment()
		env.SetDebug(true)
		gotResult, err := interp.Eval(strings.NewReader(tc.src), env)
		if err != nil {
			tt.Fatalf("eval failed (test case %d)\nsrc:\n%s\nerr:\n%s", ti, src, err)
		}

		if diff := cmp.Diff(tc.wantObj, gotResult); diff != "" {
			fmt.Printf("got: %# v\n", pretty.Formatter(gotResult))
			tt.Fatalf("mismatching results (test case %d)\nsrc:\n%s\ndiff guide:\n  - want\n  + got\ndiff:\n%s", ti, src, diff)
		}

		if diff := cmp.Diff(tc.wantStdout, stdout.String()); diff != "" {
			tt.Fatalf("mismatching results (test case %d) \nsrc:\n%s\ndiff guide:\n  - want\n  + got\ndiff:\n%s", ti, src, diff)
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
