package main

import (
	"fmt"
	"strings"

	"github.com/siadat/well/newsh"
)

func main() {
	ff(123, "hello1")
}

func ff(a int, b string) {
	files = []newsh.File{
		{"exec/exec.go"},
	}
	asdasd + 2
	fmt.Printf("a is %d and b is %q", a, b)
	// newsh.External(nil, "ps aux")
	newsh.External(nil, "ps aux")
	newsh.External(nil, "ps aux")
	return

	var srcfile = "exec/exec.go" // TODO: read from commandline
	var lines = newsh.External(
		newsh.ValMap{
			"srcfile": srcfile,
		},
		`git log --pretty='format:%C(auto)%H' --follow -- ${srcfile:%q}`, // TODO: this should be ${srcfile:%q}
	)
	for i, hash := range strings.Split(lines, "\n") {
		var ord = i + 1
		fmt.Println("commit", ord, hash)
		var contents = newsh.ExternalPiped(
			newsh.ValMap{
				"hash":    hash,
				"prefix":  fmt.Sprintf("%5d %s", ord, hash[:5]),
				"srcfile": srcfile,
			},
			newsh.Pipe{
				`git show ${hash}:${srcfile:%q}`,
				`nl --body-numbering=a`,
				`perl -pe 's#^#${prefix}#'`,
				// `less`, // TODO: attach tty (and/or stdin?)
			},
		)
		fmt.Println(contents)
	}
}
