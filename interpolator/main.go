package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/siadat/well/expander"
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
	var byts, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	var s = expander.MustEncodeToString(string(byts), envMapper)
	fmt.Print(s)
}
