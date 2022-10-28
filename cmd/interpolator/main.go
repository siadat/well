package main

import (
	"fmt"
	"io"
	"os"

	"github.com/siadat/well/syntax/strs/expander"
)

func envMapper(name string) interface{} {
	var value, ok = os.LookupEnv(name)
	if !ok {
		fmt.Printf("Missing value for variable %q. Did you export it?\n", name)
		os.Exit(1)
		return nil
	}
	return value
}

func main() {
	var byts, err = io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	var s = expander.MustEncodeToString(string(byts), envMapper)
	fmt.Print(s)
}
