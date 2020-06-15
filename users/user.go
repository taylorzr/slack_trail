package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

func usersFromSlack() ([]User, error) {
	slackers, err := slackClient.GetUsers()

	if err != nil {
		return nil, errors.Wrap(err, "getting slack users")
	}

	users := []User{}

	for _, slacker := range slackers {
		users = append(users, fromSlacker(slacker))
	}

	return users, nil
}

func usersFromDatabase() ([]User, error) {
	users := []User{}

	err := db.Select(&users, "SELECT * FROM users")

	if err != nil {
		return nil, errors.Wrap(err, "selecting all users")
	}

	return users, nil
}

func usersById(ids []string) ([]User, error) {
	var users []User

	for i, id := range ids {
		ids[i] = fmt.Sprintf("'%s'", id)
	}

	query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)

	if err != nil {
		return nil, err
	}

	err = db.Select(&users, db.Rebind(query), args...)

	return users, errors.Wrap(err, "finding users by id")
}

type User struct {
	ID          string      `db:"id"`
	Deleted     bool        `db:"deleted"`
	RealName    string      `db:"real_name"`
	DisplayName string      `db:"display_name"`
	Name        string      `db:"name"`
	Avatar      string      `db:"avatar"`
	Status      string      `db:"status"`
	Title       string      `db:"title"`
	CreatedAt   time.Time   `db:"created_at"`
	DeletedAt   pq.NullTime `db:"deleted_at"`
	Admin       bool        `db:"admin"`
	Bot         bool        `db:"bot"`
}

func (user *User) Update() error {
	_, err := db.NamedExec(`
    UPDATE users SET
			name = :name,
			real_name = :real_name,
			display_name = :display_name,
			avatar = :avatar,
			deleted = :deleted,
			deleted_at = :deleted_at,
			status = :status,
			title = :title
		WHERE
		  id = :id
	`, user)

	return errors.Wrapf(err, "updating user %#v", user)
}

func (user *User) Bury() error {
	text := fmt.Sprintf("After %s, %s died of %s", user.Age(), user.SomeName(), randomDisease())

	err := message(text, ":rip:", slack.Attachment{
		ImageURL: user.Avatar,
		Title:    "",
	})

	if err != nil {
		return errors.Wrap(err, "sending rip message")
	}

	user.Deleted = true
	user.DeletedAt = pq.NullTime{Time: time.Now(), Valid: true}

	err = user.Update()

	return errors.Wrap(err, "updating user")
}

func (user *User) ChangeName(newName string) error {
	text := fmt.Sprintf("%s changed their handle from %s to %s", user.SomeName(), user.DisplayName, newName)

	err := message(text, ":name_badge:")

	if err != nil {
		return errors.Wrap(err, "sending name change message")
	}

	user.DisplayName = newName

	err = user.Update()

	return errors.Wrap(err, "updating user")
}

func ignorableStatus(from, to string) bool {
	ignorable := map[string]bool{
		":slack_call: On a call":               true,
		":spiral_calendar_pad: In a meeting":   true,
		":bus: Commuting":                      true,
		":palm_tree: Vacationing":              true,
		":house_with_garden: Working remotely": true,
	}

	return ignorable[from] || ignorable[to]
}

func (user *User) ChangeStatus(newStatus string) error {
	text := fmt.Sprintf("%s changed their status from %s to %s", user.SomeName(), user.Status, newStatus)

	if ignorableStatus(user.Status, newStatus) {
		log.Println("Status is spam, not sending slack message...")
	} else {
		err := message(text, ":thought_balloon:")

		if err != nil {
			return errors.Wrap(err, "sending status change message")
		}
	}

	user.Status = newStatus

	err := user.Update()

	return errors.Wrapf(err, "updating user %s", user.DisplayName)
}

func (user *User) ChangeTitle(newTitle string) error {
	text := fmt.Sprintf("%s changed their title from %s to %s", user.SomeName(), user.Title, newTitle)

	err := message(text, ":name_badge:")

	if err != nil {
		return errors.Wrap(err, "sending title change message")
	}

	user.Title = newTitle

	err = user.Update()

	return errors.Wrapf(err, "updating user %s", user.Title)
}

func (user *User) Necromance() error {
	text := fmt.Sprintf("%s is back from the dead!", user.SomeName())

	err := message(text, ":zombie:", slack.Attachment{
		ImageURL: user.Avatar,
		Title:    "",
	})

	if err != nil {
		return errors.Wrap(err, "sending zombie message")
	}

	_, err = db.NamedExec(`
		UPDATE users SET
		  deleted = false,
			deleted_at = null
		WHERE id = :id`, user)

	return errors.Wrapf(err, "updating user %s to not deleted", user.DisplayName)
}

func (user *User) SomeName() string {
	if user.RealName == "" {
		return user.Name
	} else {
		return user.RealName
	}
}

func (user *User) Age() time.Duration {
	t := time.Now()

	if user.DeletedAt.Valid {
		t = user.DeletedAt.Time
	}

	duration := t.Sub(user.CreatedAt)

	return duration
}

func getUser(name string) (*User, error) {
	user := User{}

	err := db.Get(&user, "SELECT * from users where name = $1", name)

	return &user, errors.Wrapf(err, "get user %s failed", name)
}

func createUser(user *User) (*User, error) {
	user.CreatedAt = time.Now()

	if user.Deleted {
		user.DeletedAt = pq.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := db.NamedExec(`
		INSERT INTO users
		(id, name, real_name, display_name, avatar, deleted, deleted_at, created_at, status, title)
		VALUES
		(:id, :name, :real_name, :display_name, :avatar, :deleted, :deleted_at, :created_at, :status, :title)
		`, user)

	return user, errors.Wrapf(err, "inserting user %#v", user)
}

func fromSlacker(slacker slack.User) User {
	return User{
		ID:          slacker.ID,
		Name:        slacker.Name,
		RealName:    slacker.RealName,
		DisplayName: slacker.Profile.DisplayName,
		Deleted:     slacker.Deleted,
		Avatar:      slacker.Profile.ImageOriginal,
		Status:      fmt.Sprintf("%s %s", slacker.Profile.StatusEmoji, slacker.Profile.StatusText),
		Title:       slacker.Profile.Title,
		Admin:       slacker.IsAdmin,
		Bot:         slacker.IsBot,
	}
}

func registerAndAnnounceBaby(baby User) error {
	user, err := createUser(&baby)

	if err != nil {
		return errors.Wrapf(err, "creating user %#v", baby)
	}

	text := ""
	if baby.Deleted {
		text = "I'm sorry for your loss, %s was stillborn"
	} else {
		text = "Congratulations, you have a beautiful new baby named %s"
	}

	err = message(fmt.Sprintf(text, user.SomeName()), ":baby:")

	return errors.Wrap(err, "sending message")
}
