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
	HasApplied      bool
	HasSaved        bool
}

type JobApplication struct {
	ID          int
	ApplyDate   time.Time
	CoverLetter string
	JobPostID   int
	UserID      int
}

type JobSave struct {
	ID      int
	UserID  int
	JobPost JobPost
}

type JobSeekerSave struct {
	ID               int
	JobSeekerProfile JobSeekerProfile
	JobPost          JobPost
}

func (m *DBModel) SaveJobPost(jp JobPost) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	//delete the job post, let cascading take care of it and add it back in
	if jp.ID != 0 {
		stmt := "delete from job_post_activity where job_post_id = ?"
		_, err := m.DB.ExecContext(ctx, stmt, jp.ID)
		if err != nil {
			return err
		}
	}

	// Add the JobPost
	stmt := `insert into job_post_activity (posted_by_id, description_of_job, job_title, job_type, posted_date, 
                               remote, salary)
			values (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := m.DB.ExecContext(ctx, stmt,
		jp.PostedById,
		jp.Description,
		jp.JobTitle,
		jp.JobType,
		time.Now(),
		jp.Remote,
		jp.Salary,
	)
	if err != nil {
		return err
	}

	//Mysql / mariadb does not support "returning id". Use last_insert_id.
	//The ID that was generated is maintained on the server on a per-connection basis.
	query := "select LAST_INSERT_ID()"
	row := m.DB.QueryRowContext(ctx, query)
	err = row.Scan(&jp.ID)
	if err != nil {
		return err
	}

	//Add JobLocation
	stmt = "insert into job_location (job_post_activity_id, city, state, country) values (?, ?, ?, ?)"
	_, err = m.DB.ExecContext(ctx, stmt,
		jp.ID,
		jp.Location.City,
		jp.Location.State,
		jp.Location.Country,
	)
	if err != nil {
		return err
	}

	//Add JobCompany
	stmt = "insert into job_company (job_post_activity_id, name, logo) values (?, ?, ?)"
	_, err = m.DB.ExecContext(ctx, stmt,
		jp.ID,
		jp.Company.Name,
		jp.Company.Logo,
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
			inner join job_location l on j.job_post_id = l.job_post_activity_id
			inner join job_company c  on j.job_post_id = c.job_post_activity_id
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
		if err != nil {
			return jobPosts, err
		}

		jobPosts = append(jobPosts, &jp)
	}

	return jobPosts, nil
}

func (m *DBModel) GetJob(id int) (*JobPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var jp JobPost
	query := `
		select j.job_post_id, j.job_title, j.posted_date, j.salary, j.job_type, j.remote, j.description_of_job,
		       l.id as locationId, l.city, l.state, l.country, 
		       c.id as companyId, c.name
		from job_post_activity j
			inner join job_location l on j.job_post_id = l.job_post_activity_id
			inner join job_company c  on j.job_post_id = c.job_post_activity_id
			left join job_seeker_apply s on s.job = j.job_post_id
		where j.job_post_id = ?
	`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&jp.ID,
		&jp.JobTitle,
		&jp.PostedDate,
		&jp.Salary,
		&jp.JobType,
		&jp.Remote,
		&jp.Description,
		&jp.Location.ID,
		&jp.Location.City,
		&jp.Location.State,
		&jp.Location.Country,
		&jp.Company.ID,
		&jp.Company.Name,
	)
	if err != nil {
		return nil, err
	}

	return &jp, nil
}

func (m *DBModel) GetJobApplicationCountForUserId(jobPostID, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var count int

	query := "select count(id) from job_seeker_apply where job = ? and user_id = ?"
	err := m.DB.QueryRowContext(ctx, query, jobPostID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *DBModel) GetJobSaveCountForUserId(jobPostID, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var count int

	query := "select count(id) from job_seeker_save where job = ? and user_id = ?"
	err := m.DB.QueryRowContext(ctx, query, jobPostID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *DBModel) SaveJobApplication(jobPostId, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "insert into job_seeker_apply (apply_date, cover_letter, job, user_id) values (?, ?, ?, ?)"
	_, err := m.DB.ExecContext(ctx, stmt,
		time.Now(),
		"",
		jobPostId,
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) SaveJobInterest(jobPostId, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "insert into job_seeker_save (job, user_id) values (?, ?)"
	_, err := m.DB.ExecContext(ctx, stmt,
		jobPostId,
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetJobApplicationsByUserId(id int) ([]*JobApplication, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := "select id, apply_date, cover_letter, job, user_id from job_seeker_apply where user_id = ?"
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobApplications []*JobApplication
	for rows.Next() {
		var ja JobApplication
		err = rows.Scan(
			&ja.ID,
			&ja.ApplyDate,
			&ja.CoverLetter,
			&ja.JobPostID,
			&ja.UserID,
		)
		if err != nil {
			return jobApplications, err
		}

		jobApplications = append(jobApplications, &ja)
	}

	return jobApplications, nil
}

func (m *DBModel) GetJobApplicationsByJobPostId(id int) ([]*JobApplication, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := "select id, apply_date, cover_letter, job, user_id from job_seeker_apply where job = ?"
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobApplications []*JobApplication
	for rows.Next() {
		var ja JobApplication
		err = rows.Scan(
			&ja.ID,
			&ja.ApplyDate,
			&ja.CoverLetter,
			&ja.JobPostID,
			&ja.UserID,
		)
		if err != nil {
			return jobApplications, err
		}

		jobApplications = append(jobApplications, &ja)
	}

	return jobApplications, nil
}

func (m *DBModel) GetSavedJobsByUserId(id int) ([]*JobSave, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select s.id, s.user_id,
		       j.job_post_id, j.job_title, j.job_type, j.remote, j.salary,
		       l.id as locationId, l.city, l.state, l.country,
		       c.id as companyId, c.name
		from job_seeker_save s
		    inner join job_post_activity j on s.job = j.job_post_id
		    inner join job_location l on j.job_post_id = l.job_post_activity_id
		    inner join job_company c  on j.job_post_id = c.job_post_activity_id
		and s.user_id = ?  
	`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobSaves []*JobSave
	for rows.Next() {
		var js JobSave
		err = rows.Scan(
			&js.ID,
			&js.UserID,
			&js.JobPost.ID,
			&js.JobPost.JobTitle,
			&js.JobPost.JobType,
			&js.JobPost.Remote,
			&js.JobPost.Salary,
			&js.JobPost.Location.ID,
			&js.JobPost.Location.City,
			&js.JobPost.Location.State,
			&js.JobPost.Location.Country,
			&js.JobPost.Company.ID,
			&js.JobPost.Company.Name,
		)
		if err != nil {
			return jobSaves, err
		}

		jobSaves = append(jobSaves, &js)
	}

	return jobSaves, nil
}
