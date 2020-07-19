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

	app.Name = "emoji"
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
				err := initializeUsers()
				if err != nil {
					return errors.Wrap(err, "initializing users")
				}

				err = initializeEmojis()
				return errors.Wrap(err, "initializing emojis")
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
		{
			Name:  "org",
			Usage: "list employees according to ultipro",
			Action: func(c *cli.Context) error {
				browser, err := Login()

				if err != nil {
					return err
				}

				// Adam's ID
				root, err := GetDirectReports(browser, "BY4GHG02C0K0")

				if err != nil {
					return err
				}

				fmt.Println("Finding all employees...")
				peeps := GetAllReports(browser, root, []*Person{}, []int{})

				lookup := map[string]*Person{}

				for _, p := range peeps {
					lookup[p.ID] = p
				}

				for _, p := range peeps {
					supervisor := lookup[p.SupervisorID]

					boss := "unknown"

					if supervisor != nil {
						boss = supervisor.Name
					}
					fmt.Println(p.Name, boss)
				}

				fmt.Println("Total:", len(peeps))

				return nil
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
