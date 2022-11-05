package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/siadat/well/interpreter"
	"github.com/siadat/well/types"
	"github.com/urfave/cli/v2"
)

func main() {
	var app = &cli.App{
		Name:  "guillemets",
		Usage: "Utility cli tool for working with guillemets",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug mode",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "execute a Well file",
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

					checker := types.NewChecker()
					checker.SetDebug(cmdCtx.Bool("debug"))
					var _, checkErr = checker.Check(bytes.NewReader(byts))
					if checkErr != nil {
						return checkErr
					}

					interp := interpreter.NewInterpreter(os.Stdout, os.Stderr)
					interp.SetVerbose(cmdCtx.Bool("verbose"))
					interp.SetDebug(cmdCtx.Bool("debug"))
					env := interpreter.NewEnvironment()
					env.SetDebug(cmdCtx.Bool("debug"))
					var _, evalErr = interp.Eval(bytes.NewReader(byts), env)
					if evalErr != nil {
						return evalErr
					}

					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
