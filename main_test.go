package main

import (
	"testing"

	"github.com/nlopes/slack"
)

func TestNewborn(t *testing.T) {
	users := []User{}

	slackUsers := []slack.User{{
		ID:       "zt",
		Name:     "zach",
		RealName: "Zach Taylor",
		Deleted:  false,
	}}

	result := diff(users, slackUsers)

	if count := len(result.Babies); count != 1 {
		t.Errorf("Expected 1, got %d\n%#v\n", count, result)
	}
}

func TestCorpse(t *testing.T) {
	users := []User{{
		ID:       "zt",
		Name:     "zach",
		RealName: "Zach Taylor",
		Deleted:  false,
	}}

	slackUsers := []slack.User{{
		ID:       "zt",
		Name:     "zach",
		RealName: "Zach Taylor",
		Deleted:  true,
	}}

	result := diff(users, slackUsers)

	if count := len(result.Corpses); count != 1 {
		t.Errorf("Expected 1, got %d\n%#v\n", count, result)
	}
}

func TestZombie(t *testing.T) {
	users := []User{{
		ID:      "zt",
		Deleted: true,
	}}

	slackUsers := []slack.User{{
		ID:       "zt",
		Name:     "zach",
		RealName: "Zach Taylor",
		Deleted:  false,
	}}

	result := diff(users, slackUsers)

	if count := len(result.Zombies); count != 1 {
		t.Errorf("Expected 1, got %d\n%#v\n", count, result)
	}
}
