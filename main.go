package main

import (
	"fmt"
	"os"

	"github.com/joshuapmorgan/forgerock-aws-sts/cmd"
	"github.com/joshuapmorgan/forgerock-aws-sts/version"
	"github.com/mitchellh/cli"
)

var Ui cli.Ui

func init() {
	Ui = &cli.BasicUi{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
}

func main() {
	meta := cmd.Meta{
		Ui: Ui,
	}

	ver := fmt.Sprintf("%s %s (%s)", version.Version, version.GitCommit, version.BuildDate)

	c := cli.NewCLI("forgerock-aws-sts", ver)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"configure": func() (cli.Command, error) {
			return &cmd.ConfigureCommand{
				Meta: meta,
			}, nil
		},

		"login": func() (cli.Command, error) {
			return &cmd.LoginCommand{
				Meta: meta,
			}, nil
		},
	}

	c.Run()
}
