package main

import (
	"fmt"
	"io/ioutil"
	"os"

	execsh "github.com/siadat/well/exec"
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
	var s = execsh.EncodeToString(string(byts), envMapper)
	fmt.Print(s)
}
