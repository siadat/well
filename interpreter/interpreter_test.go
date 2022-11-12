package interpreter_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
	"github.com/siadat/well/interpreter"
	"github.com/siadat/well/syntax/scanner"
)

var testCases = []struct {
	src        string
	wantObj    interpreter.Object
	wantStdout string
}{
	{
		src: `
		external echo(stdin reader, s string) => "echo ${s:%q}"

	    function main() {
	        print_stream(echo(MainStdin, "hello"))
	    }
	    `,
		wantObj:    nil,
		wantStdout: "hello\n",
	},
	{
		src: `
		external echo(stdin reader, s string) => "echo ${s:%q}"
		external ping(stdin reader, ip string) => "ping -c1 ${ip:%q}"
		external nl(stdin reader) => "nl"

		function greet(s1 string, s2 string) {
			println(s1, "and", s2)
			println(f2(0, 0))
			if "hello" ~~ "ll" {
			  return true
			}
			print_stream(ping(MainStdin, "4.2.2.4"))
			print_stream(ping(MainStdin, "127.0.0.1"))
		}
		function f2(s1 string, s2 string) (string) {
			return "s1=${s1} and s2=${s2}"
		}

		function main() {
			let s1 = "hi"
			let bye = "bye"
			let out = MainStdin | echo("hello1") | nl()
			print_stream(out)
			let res = greet(s1, bye)
			println(res)
		}
		`,
		wantObj:    nil,
		wantStdout: "     1\thello1\nhi and bye\ns1=0 and s2=0\ntrue\n",
	},
}

func TestParser(tt *testing.T) {
	for ti, tc := range testCases {
		var src = tc.src
		src = scanner.FormatSrc(src, true)
		fmt.Println("testing ->")
		fmt.Println(src)

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		interp := interpreter.NewInterpreter(&stdout, &stderr)
		interp.SetDebug(true)
		var env = interpreter.NewEnvironment()
		if err := env.Set("MainStdin", &interpreter.PipeStream{ReadCloser: os.Stdin}); err != nil {
			tt.Fatal(err)
		}
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
