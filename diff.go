package main

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
		Babies        []User
		Corpses       []User
		Zombies       []User
		NameChanges   []NameChange
		StatusChanges []StatusChange
	}
)

func (d *DiffResult) AddBaby(user User) {
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

func diff(users []User, slacker []User) DiffResult {
	diff := DiffResult{}

	lookup := make(map[string]User)

	for _, user := range users {
		lookup[user.ID] = user
	}

	for _, slacker := range slacker {
		if user, ok := lookup[slacker.ID]; ok {
			displayName := slacker.DisplayName
			if displayName != user.DisplayName {
				diff.AddNameChange(user, displayName)
			}

			if slacker.Status != user.Status {
				diff.AddStatusChange(user, slacker.Status)
			}

			if slacker.Deleted != user.Deleted {
				if slacker.Deleted {
					diff.AddCorpse(user)
				} else {
					diff.AddZombie(user)
				}
			}
		} else {
			diff.AddBaby(slacker)
		}
	}

	return diff
}
