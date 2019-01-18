package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/nlopes/slack"
	cli "gopkg.in/urfave/cli.v1"
)

func runCLI() {
	app := cli.NewApp()

	app.Name = "trail"
	app.Version = "0.1"

	app.Commands = []cli.Command{
		{
			Name:  "run",
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
				return err
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
							return err
						}

						attachments := []slack.Attachment{}

						for _, item := range result.Items[0:5] {
							attachments = append(attachments, slack.Attachment{
								Title:    "",
								ImageURL: item.Image.ThumbnailLink,
							})
						}

						_, _, err = slackClient.PostMessage(
							slackChannelID,
							slack.MsgOptionText("Are any of these the new baby amountee?!?", false),
							slack.MsgOptionIconEmoji(":frame_with_picture:"),
							slack.MsgOptionAttachments(attachments...),
						)

						return err
					},
				},
				{
					Name:  "slack",
					Usage: "post test message to slack",
					Action: func(c *cli.Context) error {
						_, _, err := slackClient.PostMessage(
							slackChannelID,
							slack.MsgOptionUser("trail"),
							slack.MsgOptionText("Testing, testing, 123...", false),
							slack.MsgOptionUsername("trail"),
							slack.MsgOptionIconEmoji(":rip:"),
						)
						return err
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
