package erroring

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/maruel/panicparse/v2/stack"
)

// PrintTrace uses panicparse to print a readable panic stack.
// See https://pkg.go.dev/github.com/maruel/panicparse/v2/stack
func PrintTrace() {
	// debug.PrintStack()
	var stream = bytes.NewReader(debug.Stack())

	var s, suffix, err = stack.ScanSnapshot(stream, os.Stdout, stack.DefaultOpts())
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Find out similar goroutine traces and group them into buckets.
	var buckets = s.Aggregate(stack.AnyValue).Buckets

	// Calculate alignment.
	var colLen = 0
	for _, bucket := range buckets {
		for _, line := range filterCalls(bucket.Signature.Stack.Calls) {
			if l := len(formatFilename(line)); l > colLen {
				colLen = l
			}
		}
	}

	for _, bucket := range buckets {
		// Print the goroutine header.
		var extra = ""
		if s := bucket.SleepString(); s != "" {
			extra += " [" + s + "]"
		}
		if bucket.Locked {
			extra += " [locked]"
		}

		if len(bucket.CreatedBy.Calls) != 0 {
			extra += fmt.Sprintf(" [Created by %s.%s @ %s:%d]",
				bucket.CreatedBy.Calls[0].Func.DirName,
				bucket.CreatedBy.Calls[0].Func.Name,
				bucket.CreatedBy.Calls[0].SrcName,
				bucket.CreatedBy.Calls[0].Line,
			)
		}
		fmt.Printf("%d: %s%s\n", len(bucket.IDs), bucket.State, extra)

		// Print the stack lines.
		for _, line := range filterCalls(bucket.Signature.Stack.Calls) {
			fmt.Println(formatCall(line, colLen))
		}
		if bucket.Stack.Elided {
			io.WriteString(os.Stdout, "    (...) (elided)\n")
		}
	}

	// If there was any remaining data in the pipe, dump it now.
	if len(suffix) != 0 {
		os.Stdout.Write(suffix)
	}
	if err == nil {
		io.Copy(os.Stdout, stream)
	}
}

func filterCalls(lines []stack.Call) []stack.Call {
	var ret []stack.Call
	var sawStdlibPanic = false
	for _, line := range lines {
		if !sawStdlibPanic {
			if line.Func.DirName == "" && line.SrcName == "panic.go" {
				sawStdlibPanic = true
				continue
			} else {
				continue
			}
		}

		if line.Func.IsPkgMain {
			ret = append(ret, line)
			continue
		}

		if strings.HasPrefix(line.ImportPath, "github.com/siadat/well") {
			ret = append(ret, line)
			continue
		}
	}
	return ret
}

func formatCall(line stack.Call, colLen int) string {
	return fmt.Sprintf(
		"    %-*s %s(...)",
		colLen,
		formatFilename(line),
		line.Func.Name,
		// &line.Args, // pretty.Sprintf("% #v", line.Args),
	)
}

func formatFilename(line stack.Call) string {
	// return fmt.Sprintf("%s/%s:%d", line.Func.ImportPath, line.SrcName, line.Line)
	return fmt.Sprintf("%s/%s:%d", line.Func.DirName, line.SrcName, line.Line)
}
