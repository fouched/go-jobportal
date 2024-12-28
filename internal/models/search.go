package models

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type SearchCriteria struct {
	Title         string
	Location      string
	Today         bool
	Days7         bool
	Days30        bool
	RemoteOnly    bool
	OfficeOnly    bool
	PartialRemote bool
	PartTime      bool
	FullTime      bool
	Freelance     bool
}

func (m *DBModel) SearchJobPosts(r *http.Request) (SearchCriteria, []*JobPost, error) {
	//TODO this really needs pagination
	var searchCriteria SearchCriteria
	var jobPosts []*JobPost

	err := r.ParseForm()
	if err != nil {
		return searchCriteria, jobPosts, err
	}

	searchCriteria.Title = r.Form.Get("title")
	searchCriteria.Location = r.Form.Get("location")
	searchCriteria.Today = strings.EqualFold(r.Form.Get("today"), "on")
	searchCriteria.Days7 = strings.EqualFold(r.Form.Get("days7"), "on")
	searchCriteria.Days30 = strings.EqualFold(r.Form.Get("days30"), "on")
	searchCriteria.RemoteOnly = strings.EqualFold(r.Form.Get("remoteOnly"), "on")
	searchCriteria.OfficeOnly = strings.EqualFold(r.Form.Get("officeOnly"), "on")
	searchCriteria.PartialRemote = strings.EqualFold(r.Form.Get("partialRemote"), "on")
	searchCriteria.PartTime = strings.EqualFold(r.Form.Get("partTime"), "on")
	searchCriteria.FullTime = strings.EqualFold(r.Form.Get("fullTime"), "on")
	searchCriteria.Freelance = strings.EqualFold(r.Form.Get("freelance"), "on")

	var searchDate time.Time
	isDateSearch := true
	isRemote := true
	isJobType := true

	if searchCriteria.Days30 {
		searchDate = time.Now().Add(time.Hour * 24 * 30 * -1)
	} else if searchCriteria.Days7 {
		searchDate = time.Now().Add(time.Hour * 24 * 7 * -1)
	} else if searchCriteria.Today {
		searchDate = time.Now()
	} else {
		isDateSearch = false
	}

	if !searchCriteria.PartTime && !searchCriteria.FullTime && !searchCriteria.Freelance {
		isRemote = false
	}

	if !searchCriteria.OfficeOnly && !searchCriteria.RemoteOnly && !searchCriteria.PartialRemote {
		isJobType = false
	}

	if !isDateSearch && !isRemote && !isJobType &&
		strings.EqualFold(searchCriteria.Title, "") && strings.EqualFold(searchCriteria.Location, "") {

		jobPosts, err = searchAll(m)
		if err != nil {
			return searchCriteria, jobPosts, err
		}

	} else {
		//add job types
		var jobTypes []string
		if searchCriteria.FullTime {
			jobTypes = append(jobTypes, "'Full-Time'")
		}
		if searchCriteria.PartTime {
			jobTypes = append(jobTypes, "'Part-Time'")
		}
		if searchCriteria.Freelance {
			jobTypes = append(jobTypes, "'Freelance'")
		}

		var jobType = ""
		for _, j := range jobTypes {
			jobType += j + ","
		}
		jobType = strings.TrimSuffix(jobType, ",")
		//if none specified add all
		if strings.EqualFold(jobType, "") {
			jobType = "'Full-Time', 'Part-Time', 'Freelance'"
		}

		//add remotes
		var remotes []string
		if searchCriteria.OfficeOnly {
			remotes = append(remotes, "'Office-Only'")
		}
		if searchCriteria.PartialRemote {
			remotes = append(remotes, "'Partial-Remote'")
		}
		if searchCriteria.RemoteOnly {
			remotes = append(remotes, "'Remote-Only'")
		}
		var remote = ""
		for _, r := range remotes {
			remote += r + ","
		}
		remote = strings.TrimSuffix(remote, ",")
		//if none specified add all
		if strings.EqualFold(remote, "") {
			remote = "'Office-Only', 'Partial-Remote', 'Remote-Only'"
		}

		if isDateSearch {
			jobPosts, err = searchWithDate(m, searchDate, "%"+searchCriteria.Title+"%", "%"+searchCriteria.Location+"%", jobType, remote)
			if err != nil {
				return searchCriteria, jobPosts, err
			}
		} else {
			jobPosts, err = searchWithoutDate(m, "%"+searchCriteria.Title+"%", "%"+searchCriteria.Location+"%", jobType, remote)
			if err != nil {
				return searchCriteria, jobPosts, err
			}
		}
	}

	return searchCriteria, jobPosts, nil
}

func searchAll(m *DBModel) ([]*JobPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var jobPosts []*JobPost

	query := `
			select j.job_post_id, j.job_title, j.job_type, j.remote, j.salary,
			       l.id as locationId, l.city, l.state, l.country,
			       c.id as companyId, c.name
			from job_post_activity j
				inner join job_location l on j.job_post_id = l.job_post_activity_id
				inner join job_company c  on j.job_post_id = c.job_post_activity_id			
		`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return jobPosts, err
	}
	defer rows.Close()

	for rows.Next() {
		var jp JobPost

		err = rows.Scan(
			&jp.ID,
			&jp.JobTitle,
			&jp.JobType,
			&jp.Remote,
			&jp.Salary,
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

func searchWithDate(m *DBModel, searchDate time.Time, title, location, jobType, remote string) ([]*JobPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var jobPosts []*JobPost

	query := `
			select j.job_post_id, j.job_title, j.job_type, j.remote, j.salary,
			       l.id as locationId, l.city, l.state, l.country,
			       c.id as companyId, c.name
			from job_post_activity j
				inner join job_location l on j.job_post_id = l.job_post_activity_id
				inner join job_company c  on j.job_post_id = c.job_post_activity_id
			where j.job_title like ?
			and posted_date >= ?
			and (
				l.city like ?
			    or l.country like ?
			    or l.state like ? 
			)
		`
	// substitution does not work with multiple in clauses so do it manually
	query += "and j.job_type in (" + jobType + ") " +
		"and j.remote in (" + remote + ")"

	rows, err := m.DB.QueryContext(ctx, query, title, searchDate, location, location, location)
	if err != nil {
		return jobPosts, err
	}

	defer rows.Close()

	for rows.Next() {
		var jp JobPost

		err = rows.Scan(
			&jp.ID,
			&jp.JobTitle,
			&jp.JobType,
			&jp.Remote,
			&jp.Salary,
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

func searchWithoutDate(m *DBModel, title, location, jobType, remote string) ([]*JobPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var jobPosts []*JobPost

	query := `
			select j.job_post_id, j.job_title, j.job_type, j.remote, j.salary,
			       l.id as locationId, l.city, l.state, l.country,
			       c.id as companyId, c.name
			from job_post_activity j
				inner join job_location l on j.job_post_id = l.job_post_activity_id
				inner join job_company c  on j.job_post_id = c.job_post_activity_id
			where j.job_title like ?
			and (
				l.city like ?
			    or l.country like ?
			    or l.state like ? 
			)
		`
	// substitution does not work with multiple in clauses so do it manually
	query += "and j.job_type in (" + jobType + ") " +
		"and j.remote in (" + remote + ")"

	rows, err := m.DB.QueryContext(ctx, query, title, location, location, location)
	if err != nil {
		return jobPosts, err
	}

	defer rows.Close()

	for rows.Next() {
		var jp JobPost

		err = rows.Scan(
			&jp.ID,
			&jp.JobTitle,
			&jp.JobType,
			&jp.Remote,
			&jp.Salary,
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
