package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/siadat/well/fumt"
	"github.com/siadat/well/interpreter"
	"github.com/siadat/well/types"
	"github.com/urfave/cli/v2"
)

func main() {
	var app = &cli.App{
		Name: "well",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "debug",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "run",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to Well file to be executed",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "enable verbose mode",
					},
				},
				Action: func(cmdCtx *cli.Context) error {
					var byts, readErr = os.ReadFile(cmdCtx.String("file"))
					if readErr != nil {
						return readErr
					}

					var checker = types.NewChecker()
					checker.SetDebug(cmdCtx.Bool("debug"))
					var _, checkErr = checker.Check(bytes.NewReader(byts))
					if checkErr != nil {
						fmt.Fprintf(os.Stderr, "type checker failed\n")
						return checkErr
					}

					var fileDependencies = checker.UnresolvedDependencies()
					if cmdCtx.Bool("debug") {
						fmt.Fprintf(os.Stderr, "%d file dependencies\n", len(fileDependencies))
					}
					if l := len(fileDependencies); l > 0 {
						var lines = append([]string{""}, fileDependencies...)
						return fmt.Errorf("The following external commands and files are undeclared:%s", strings.Join(lines, "\n   "))
					}

					interp := interpreter.NewInterpreter(os.Stdout, os.Stderr)
					interp.SetVerbose(cmdCtx.Bool("verbose"))
					interp.SetDebug(cmdCtx.Bool("debug"))
					env := interpreter.NewEnvironment()
					env.SetDebug(cmdCtx.Bool("debug"))

					// TODO: allow passing CLI args as function arguments. Either:
					//  [->] well run -f ./testdata/test5.well --expression 'my_function("s value", "x value")'
					//  [  ] well run -f ./testdata/test5.well --function my_function --args '-s "s value" -x "x value"'
					//  [  ] well run -f ./testdata/test5.well --function my_function -s "s value" -x "x value"
					//  [  ] well run -f ./testdata/test5.well my_function s='"s value"' x='"x value"'
					//  [->] well run -f ./testdata/test5.well my_function s="s value" x="x value"
					//  [->] well run -f ./testdata/test5.well my_function -s "s value" -x "x value"

					var _, evalErr = interp.Eval(bytes.NewReader(byts), env)
					if evalErr != nil {
						return evalErr
					}

					return nil
				},
			},
			{
				Name: "fmt",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to Well file to be executed",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "enable verbose mode",
					},
				},
				Action: func(cmdCtx *cli.Context) error {
					var byts, readErr = os.ReadFile(cmdCtx.String("file"))
					if readErr != nil {
						return readErr
					}

					var formater = fumt.NewFormater()
					formater.SetDebug(cmdCtx.Bool("debug"))
					return formater.Format(bytes.NewReader(byts), os.Stdout)
				},
			},
			// TODO: a subcommand to generate Bash script to handle the risk of "What if we change our mind?" before getting approval.
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
