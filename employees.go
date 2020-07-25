package main

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Employee struct {
	ID           string      `db:"id"`
	Name         string      `db:"name"`
	SupervisorID string      `db:"supervisor_id"`
	ReportsCount int         `db:"reports_count"`
	CreatedAt    time.Time   `db:"created_at"`
	Deleted      bool        `db:"deleted"`
	DeletedAt    pq.NullTime `db:"deleted_at"`
}

func FindEmployee(id string) (*Employee, error) {
	var employee Employee
	err := db.Get(&employee, "SELECT * FROM employees WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &employee, errors.Wrap(err, "finding employee by id")
}

func (employee *Employee) Update() error {
	_, err := db.NamedExec(`
    UPDATE employees SET
			name = :name
			, supervisor_id = :supervisor_id
			, reports_count = :reports_count
			, deleted = :deleted
			, deleted_at = :deleted_at
		WHERE
		  id = :id
	`, employee)

	return errors.Wrapf(err, "updating employee %#v", employee)
}

func (employee Employee) ChangeSupervisor(newSupervisorID string) error {
	old, err := FindEmployee(employee.SupervisorID)
	if err != nil {
		return err
	}

	new, err := FindEmployee(newSupervisorID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("%s's supervisor changed from %s to %s", employee.Name, old.Name, new.Name)

	err = sendMessage(text, ":name_badge:")

	if err != nil {
		return errors.Wrap(err, "sending name change message")
	}

	employee.SupervisorID = newSupervisorID

	return employee.Update()
}

func createEmployee(employee *Employee) (*Employee, error) {
	employee.CreatedAt = time.Now()

	if employee.Deleted {
		employee.DeletedAt = pq.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := db.NamedExec(`
		INSERT INTO employees
		(id, name, reports_count, supervisor_id, created_at, deleted, deleted_at)
		VALUES
		(:id, :name, :reports_count, :supervisor_id, :created_at, :deleted, :deleted_at)
		`, employee)

	return employee, errors.Wrapf(err, "inserting employee %#v", employee)
}

func runEmployeesIteration() error {
	oldEmployees, err := employeesFromDatabase()
	if err != nil {
		return err
	}

	newEmployees, err := GetAllEmployees()
	if err != nil {
		return err
	}

	err = diffEmployees(oldEmployees, newEmployees)

	return errors.Wrap(err, "diffing employees")
}

// FIXME: Copied from users, modify for employees
func diffEmployees(old, new []*Employee) error {
	lookup := make(map[string]*Employee)

	for _, oldEmployee := range old {
		lookup[oldEmployee.ID] = oldEmployee
	}

	for _, newEmployee := range new {
		if oldEmployee, ok := lookup[newEmployee.ID]; ok {
			if newEmployee.SupervisorID != oldEmployee.SupervisorID {
				err := oldEmployee.ChangeSupervisor(newEmployee.SupervisorID)

				if err != nil {
					return errors.Wrap(err, "changing employees supervisor")
				}
			}
		} else {
			// TODO: New employee
			// err := registerAndAnnounceBaby(slackUser)

			// if err != nil {
			// 	return errors.Wrapf(err, "delivering a new baby user %s", slackUser.DisplayName)
			// }
		}
	}
	return nil
}

func employeesFromDatabase() ([]*Employee, error) {
	employees := []*Employee{}

	err := db.Select(&employees, "SELECT * FROM employees")

	if err != nil {
		return nil, errors.Wrap(err, "selecting all employees")
	}

	return employees, nil
}

func initializeEmployees() error {
	employees, err := GetAllEmployees()

	if err != nil {
		return err
	}

	for _, e := range employees {
		_, err = createEmployee(e)

		if err != nil {
			return err
		}
	}

	fmt.Printf("Created %d employees\n", len(employees))

	return nil
}
