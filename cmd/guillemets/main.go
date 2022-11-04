package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/siadat/well/syntax/strs/expander"
	"github.com/siadat/well/syntax/strs/parser"
	"github.com/urfave/cli/v2"
)

func envMapper(name string) interface{} {
	var value, ok = os.LookupEnv(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "Missing value for variable %q. Did you export it?\n", name)
		os.Exit(1)
		return nil
	}
	return value
}

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
				Name:  "render",
				Usage: "render a given string",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "newline",
						Usage: "add newline after output",
					},
					&cli.StringFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Usage:   "input command or string",
					},
				},
				Action: func(cmdCtx *cli.Context) error {
					var input string
					if cmdCtx.String("input") == "-" {
						var byts, err = io.ReadAll(os.Stdin)
						if err != nil {
							panic(err)
						}
						input = string(byts)
					} else {
						input = cmdCtx.String("input")
					}
					if input == "" {
						return fmt.Errorf("nothing to execute")
					}
					if cmdCtx.Bool("debug") {
						fmt.Fprintf(os.Stderr, "[debug] input: %q\n", input)
					}
					var s, err = expander.ParseAndEncodeToString(input, envMapper, cmdCtx.Bool("debug"))
					if err != nil {
						return err
					}
					fmt.Print(s)
					if cmdCtx.Bool("newline") {
						fmt.Println()
					}
					return nil
				},
			},
			{
				Name:  "exec",
				Usage: "execute a given command",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "newline",
						Usage: "add newline after output",
					},
					&cli.StringFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Usage:   "input command or string",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "enable verbose mode",
					},
				},
				Action: func(cmdCtx *cli.Context) error {
					var input string
					if cmdCtx.String("input") == "-" {
						var byts, err = io.ReadAll(os.Stdin)
						if err != nil {
							panic(err)
						}
						input = string(byts)
					} else {
						input = cmdCtx.String("input")
					}
					if input == "" {
						return fmt.Errorf("nothing to execute")
					}
					if cmdCtx.Bool("debug") {
						fmt.Fprintf(os.Stderr, "[debug] input: %#v\n", input)
					}

					var p = parser.NewParser()
					if cmdCtx.Bool("debug") {
						p.SetDebug(true)
					}
					var root, err = p.Parse(strings.NewReader(input))
					if err != nil {
						return fmt.Errorf("failed to parse command: %v", err)
					}

					var words, encodeErr = expander.EncodeToCmdArgs(root, envMapper)
					if encodeErr != nil {
						return fmt.Errorf("failed to create args: %s", encodeErr)
					}
					var pwd, pwdErr = os.Getwd()
					if pwdErr != nil {
						return fmt.Errorf("failed to get current working directory: %s", pwdErr)
					}

					if len(words) == 0 {
						return fmt.Errorf("nothing to execute")
					}

					if cmdCtx.Bool("verbose") {
						var rendered, renderErr = expander.ParseAndEncodeToString(input, envMapper, cmdCtx.Bool("debug"))
						if renderErr != nil {
							return renderErr
						}
						fmt.Fprintf(os.Stderr, "+%s\n", rendered)
					}

					var cmd = exec.CommandContext(context.TODO(), words[0], words[1:]...)
					cmd.Dir = pwd
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Stdin = os.Stdin
					cmd.Env = os.Environ()
					if err := cmd.Start(); err != nil {
						return fmt.Errorf("failed to start command: %s", err)
					}
					if err := cmd.Wait(); err != nil {
						return fmt.Errorf("failed to wait for command: %s", err)
					}
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "command failed: %s\n", err)
	}

}
