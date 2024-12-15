package models

import (
	"context"
	"time"
)

type JobCompany struct {
	ID   int
	Name string
	Logo string
}

type JobLocation struct {
	ID      int
	City    string
	State   string
	Country string
}

type JobPost struct {
	ID          int
	PostedById  int
	Location    *JobLocation
	Company     *JobCompany
	Description string
	JobTitle    string
	JobType     string
	Salary      string
	Remote      string
	PostedDate  time.Time
	IsActive    bool
	IsSaved     bool
}

func (m *DBModel) AddJobPost(jp JobPost) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	//Add JobLocation
	stmt := "insert into job_location (city, state, country) values (?, ?, ?)"
	_, err := m.DB.ExecContext(ctx, stmt,
		&jp.Location.City,
		&jp.Location.State,
		&jp.Location.Country,
	)
	if err != nil {
		return err
	}

	//Mysql / mariadb does not support "returning id". Use last_insert_id.
	//The ID that was generated is maintained on the server on a per-connection basis.
	query := "select LAST_INSERT_ID()"
	row := m.DB.QueryRowContext(ctx, query)
	err = row.Scan(&jp.Location.ID)
	if err != nil {
		return err
	}

	//Add JobCompany
	stmt = "insert into job_company (name, logo) values (?, ?)"
	_, err = m.DB.ExecContext(ctx, stmt,
		&jp.Company.Name,
		&jp.Company.Logo,
	)
	if err != nil {
		return err
	}

	query = "select LAST_INSERT_ID()"
	row = m.DB.QueryRowContext(ctx, query)
	err = row.Scan(&jp.Company.ID)
	if err != nil {
		return err
	}

	// Add the JobPost
	stmt = `insert into job_post_activity (posted_by_id, description_of_job, job_title, job_type, posted_date, 
                               remote, salary, job_company_id, job_location_id)
			values (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = m.DB.ExecContext(ctx, stmt,
		jp.PostedById,
		jp.Description,
		jp.JobTitle,
		jp.JobType,
		time.Now(),
		jp.Remote,
		jp.Salary,
		&jp.Company.ID,
		&jp.Location.ID,
	)
	if err != nil {
		return err
	}

	return nil
}
