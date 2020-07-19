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

func diffEmployees() error {
	return nil
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
