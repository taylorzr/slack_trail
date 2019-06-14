package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

type (
	NameChange struct {
		User    User
		NewName string
	}

	StatusChange struct {
		User      User
		NewStatus string
	}

	DiffResult struct {
		Babies        []slack.User
		Corpses       []User
		Zombies       []User
		NameChanges   []NameChange
		StatusChanges []StatusChange
	}
)

func (d *DiffResult) AddBaby(user slack.User) {
	d.Babies = append(d.Babies, user)
}

func (d *DiffResult) AddCorpse(user User) {
	d.Corpses = append(d.Corpses, user)
}

func (d *DiffResult) AddZombie(user User) {
	d.Zombies = append(d.Zombies, user)
}

func (d *DiffResult) AddNameChange(user User, newName string) {
	d.NameChanges = append(d.NameChanges, NameChange{User: user, NewName: newName})
}

func (d *DiffResult) AddStatusChange(user User, newStatus string) {
	d.StatusChanges = append(d.StatusChanges, StatusChange{User: user, NewStatus: newStatus})
}

func diff(users []User, slackUsers []slack.User) DiffResult {
	diff := DiffResult{}

	lookup := make(map[string]User)

	for _, user := range users {
		lookup[user.ID] = user
	}

	for _, slackUser := range slackUsers {
		if user, ok := lookup[slackUser.ID]; ok {
			displayName := slackUser.Profile.DisplayName
			if displayName != user.DisplayName {
				diff.AddNameChange(user, displayName)
			}

			status := fmt.Sprintf("%s %s", slackUser.Profile.StatusEmoji, slackUser.Profile.StatusText)
			if status != user.Status {
				diff.AddStatusChange(user, status)
			}

			if slackUser.Deleted != user.Deleted {
				if slackUser.Deleted {
					diff.AddCorpse(user)
				} else {
					diff.AddZombie(user)
				}
			}
		} else {
			diff.AddBaby(slackUser)
		}
	}

	return diff
}
