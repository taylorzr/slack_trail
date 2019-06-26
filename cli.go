package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v1"
)

func runCLI() {
	app := cli.NewApp()

	app.Name = "trail"
	app.Version = "0.1"

	app.Commands = []cli.Command{
		{
			Name:  "iterate",
			Usage: "run the trail",
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
					Name:  "images",
					Usage: "images",
					Action: func(c *cli.Context) error {
						if c.NArg() == 0 {
							return errors.New("You must provide me with at least one google image search argument")
						}

						result, err := findImages(strings.Join(c.Args(), " "))

						if err != nil {
							return errors.Wrap(err, "finding images")
						}

						attachments := []slack.Attachment{}

						for _, item := range result.Items[0:5] {
							attachments = append(attachments, slack.Attachment{
								Title:    "",
								ImageURL: item.Image.ThumbnailLink,
							})
						}

						err = message("Are any of these the new baby amountee?!?", ":frame_with_picture:", attachments...)

						return errors.Wrap(err, "sending slack message")
					},
				},
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
