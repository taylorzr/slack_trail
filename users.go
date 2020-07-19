package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
)

func usersFromMononym() ([]User, error) {
	group, err := slackClient.GetGroupInfo("GJUF0HLUC")
	if err != nil {
		return nil, errors.Wrap(err, "getting mononym users")
	}

	// NOTE: Lookup users in database because group info only returns ids and we need to look at a
	// users name. This means mononym is reliant on updates for other functions/lambdas that keep the
	// database up to date
	users, err := usersById(group.Members)

	return users, err
}

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

func (user *User) IsMononym() bool {
	usernameRegex := regexp.MustCompile(`^\w{3}\d{2}[evs]$`)
	return user.DisplayName != "" &&
		// handle doesn't include space/dot e.g. "zach taylor" or "zach.taylor"
		!strings.ContainsAny(user.DisplayName, " .") &&
		// handle is not default username e.g "zrt43e"
		!usernameRegex.Match([]byte(user.DisplayName)) &&
		// handle doesn't include full first and last names e.g "zachtaylor"
		!(strings.Contains(strings.ToLower(user.DisplayName), strings.ToLower(user.FirstName())) &&
			strings.Contains(strings.ToLower(user.DisplayName), strings.ToLower(user.LastName())))
}

// TODO: Store first/last from slack directly in db
func (user *User) FirstName() string {
	parts := strings.Split(user.RealName, " ")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}

func (user *User) LastName() string {
	parts := strings.Split(user.RealName, " ")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
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

func initializeUsers() error {
	users, err := usersFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching users from the database")
	}

	if len(users) != 0 {
		return errors.New("I expected the user table to be emtpy but it's not")
	}

	slackUsers, err := usersFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	for _, slackUser := range slackUsers {
		_, err := createUser(&slackUser)

		if err != nil {
			return errors.Wrapf(err, "creating user %#v", slackUser)
		}
	}

	return nil
}

func runUsersIteration() error {
	slackUsers, err := usersFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	knownUsers, err := usersFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching users from the database")
	}

	err = diffUsers(knownUsers, slackUsers)

	return errors.Wrap(err, "diffing users")
}

func runMononymIteration() error {
	users, err := usersFromMononym()
	usersLookup := map[string]bool{}

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	for _, user := range users {
		usersLookup[user.ID] = true
		if !user.IsMononym() {
			fmt.Printf("- %s\n", user.DisplayName)
		}
	}

	users, err = usersFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	new := []User{}
	for _, user := range users {
		if !usersLookup[user.ID] && !user.Deleted && user.IsMononym() {
			fmt.Printf("+ %s\n", user.DisplayName)
			new = append(new, user)
		}
	}

	return nil
}

func diffUsers(knownUsers, slackUsers []User) error {
	lookup := make(map[string]User)

	for _, knownUser := range knownUsers {
		lookup[knownUser.ID] = knownUser
	}

	for _, slackUser := range slackUsers {
		if user, ok := lookup[slackUser.ID]; ok {
			displayName := slackUser.DisplayName
			if displayName != user.DisplayName {
				err := user.ChangeName(slackUser.DisplayName)

				if err != nil {
					return errors.Wrap(err, "changing a users name")
				}
			}

			// NOTE: This is too spammy
			// if slackUser.Status != user.Status {
			// 	err := user.ChangeStatus(slackUser.Status)

			// 	if err != nil {
			// 		return errors.Wrap(err, "updating a users status")
			// 	}
			// }

			if slackUser.Title != user.Title {
				err := user.ChangeTitle(slackUser.Title)

				if err != nil {
					return errors.Wrap(err, "updating a users status")
				}
			}

			if slackUser.Deleted != user.Deleted {
				if slackUser.Deleted {
					err := user.Bury()

					if err != nil {
						return errors.Wrap(err, "burying a user")
					}
				} else {
					err := user.Necromance()

					if err != nil {
						return errors.Wrap(err, "raising a user from the dead")
					}
				}
			}
		} else {
			err := registerAndAnnounceBaby(slackUser)

			if err != nil {
				return errors.Wrapf(err, "delivering a new baby user %s", slackUser.DisplayName)
			}
		}
	}

	return nil
}
