package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v1"
)

var verbose bool

func runCLI() {
	app := cli.NewApp()

	app.Name = "trail"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "more cowbell",
		},
	}

	app.Before = func(c *cli.Context) error {
		verbose = c.Bool("verbose")
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "initialize application",
			Action: func(c *cli.Context) error {
				// err := initializeUsers()
				// if err != nil {
				// 	return errors.Wrap(err, "initializing users")
				// }

				// err = initializeEmojis()
				// return errors.Wrap(err, "initializing emojis")

				err := initializeEmployees()
				return errors.Wrap(err, "initializing employees")
			},
		},
		{
			Name:  "users",
			Usage: "check for users changes",
			Action: func(c *cli.Context) error {
				return runUsersIteration()
			},
		},
		{
			Name:  "emojis",
			Usage: "check for emoji changes",
			Action: func(c *cli.Context) error {
				return runEmojisIteration()
			},
		},
		{
			Name:  "mononym",
			Usage: "check for mononym changes",
			Action: func(c *cli.Context) error {
				return runMononymIteration()
			},
		},
		{
			Name:  "employees",
			Usage: "check for employees changes",
			Action: func(c *cli.Context) error {
				return runEmployeesIteration()
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
