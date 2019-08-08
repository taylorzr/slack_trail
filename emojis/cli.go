package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v1"
)

func runCLI() {
	app := cli.NewApp()

	app.Name = "emoji"
	app.Version = "0.1"

	app.Commands = []cli.Command{
		{
			Name:  "iterate",
			Usage: "check for emoji changes",
			Action: func(c *cli.Context) error {
				return runIterationWithSentry()
			},
		},
		{
			Name:  "init",
			Usage: "initialize application",
			Action: func(c *cli.Context) error {
				err := initializeApplication()
				return errors.Wrap(err, "initializing app")
			},
		},
		{
			Name:  "test",
			Usage: "manual testing",
			Action: func(c *cli.Context) error {
				fmt.Println("noop")
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "slack",
					Usage: "post test message to slack",
					Action: func(c *cli.Context) error {
						err := message("Testing, testing, 123...", ":rip:")
						return errors.Wrap(err, "sending slack message")
					},
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
