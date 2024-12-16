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
	ID              int
	PostedById      int
	TotalCandidates int
	Location        JobLocation
	Company         JobCompany
	Description     string
	JobTitle        string
	JobType         string
	Salary          string
	Remote          string
	PostedDate      time.Time
	IsActive        bool
	IsSaved         bool
}

func (m *DBModel) AddJobPost(jp JobPost) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	//Add JobLocation
	stmt := "insert into job_location (city, state, country) values (?, ?, ?)"
	_, err := m.DB.ExecContext(ctx, stmt,
		jp.Location.City,
		jp.Location.State,
		jp.Location.Country,
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
		jp.Company.Name,
		jp.Company.Logo,
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
		jp.Company.ID,
		jp.Location.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetRecruiterJobPosts(id int) ([]*JobPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select COUNT(s.user_id) as totalCandidates, j.job_post_id, j.job_title, l.id as locationId,
			l.city, l.state, l.country, c.id as companyId, c.name
		from job_post_activity j
		inner join job_location l on j.job_location_id = l.id
		inner join job_company c  on j.job_company_id = c.id
		left join job_seeker_apply s on s.job = j.job_post_id
		where j.posted_by_id = ?
		group by j.job_post_id
	`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobPosts []*JobPost
	for rows.Next() {
		var jp JobPost

		err = rows.Scan(
			&jp.TotalCandidates,
			&jp.ID,
			&jp.JobTitle,
			&jp.Location.ID,
			&jp.Location.City,
			&jp.Location.State,
			&jp.Location.Country,
			&jp.Company.ID,
			&jp.Company.Name,
		)

		jobPosts = append(jobPosts, &jp)
	}

	return jobPosts, nil
}
