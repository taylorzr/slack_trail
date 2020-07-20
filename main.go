package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/urfave/cli"
)

/*
Concept:

Fetch all users from slack and compare the slack user list to our user list stored in the database

Any user who is in the slack list, but not our list is a new user, so announce those new users, and
store them in the database

Any user who is in the slack list as deleted, but in our list is not deleted, is a user who has been
recently deleted. Announce them as a deleted user, and update their deleted state in the database.

*/

var (
	slackChannelID string
	slackClient    *slack.Client
	verbose        bool
	aws            bool
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	slackChannelID = os.Getenv("SLACK_CHANNEL_ID")
	slackClient = slack.New(os.Getenv("SLACK_TOKEN"))
}

func main() {
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
				// FIXME: Problem here is aws doesn't call trail with argument
				// argument is stored in env COMMAND
				if aws {
					// FIXME: Report errors?
					lambda.Start(withSentry(runUsersIteration))
					return nil
				} else {
					return runUsersIteration()
				}
			},
		},
		{
			Name:  "emojis",
			Usage: "check for emoji changes",
			Action: func(c *cli.Context) error {
				if aws {
					// FIXME: Report errors?
					lambda.Start(withSentry(runEmojisIteration))
					return nil
				} else {
					return runEmojisIteration()
				}
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
				if aws {
					// FIXME: Report errors?
					lambda.Start(withSentry(runEmployeesIteration))
					return nil
				} else {
					return runEmployeesIteration()
				}
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

	if _, exists := os.LookupEnv("LAMBDA"); exists {
		aws = true
		if _, exists := os.LookupEnv("COMMAND"); !exists {
			lambda.Start(withSentry(func() error {
				return fmt.Errorf("You must specify an env COMMAND")
			}))
		}
	}
	args := os.Args
	if aws {
		args = append(args, os.Getenv("COMMAND"))
	}
	fmt.Println(args)
	err := app.Run(args)

	if err != nil {
		log.Fatal(err)
	}
}

// Oregon Trail Diseases:
var diseases = []string{
	"Dysentery",
	"Typhoid Fever",
	"Cholera",
	"Diphtheria",
	"Measles",
	"Thirst Traps",
}

func randomDisease() string {
	rand.Seed(time.Now().UTC().UnixNano())
	disease := diseases[rand.Intn(len(diseases)+1)]
	return disease
}

func message(text string, emoji string, attachments ...slack.Attachment) error {
	_, _, err := slackClient.PostMessage(
		slackChannelID,
		slack.MsgOptionUsername("trail"),
		slack.MsgOptionIconEmoji(emoji),
		slack.MsgOptionText(text, false),
		slack.MsgOptionAttachments(attachments...),
	)

	return errors.Wrap(err, "sending a slack message")
}

func withSentry(f func() error) func() error {
	function := f
	return func() error {

		fmt.Println("Starting iteration...")
		var err error

		raven.CapturePanic(func() {
			err = function()

			if err != nil {
				fmt.Printf("Something broke :(\n%s\n", err.Error())
				raven.CaptureErrorAndWait(err, nil)
			}
		}, nil)

		fmt.Println("Finished iteration.")

		return err
	}
}
