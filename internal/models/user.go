package models

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID               int
	Email            string
	Password         string
	IsActive         bool
	RegistrationDate time.Time
	UserType         *UserType
}

type UserType struct {
	ID           int
	UserTypeName string
	Users        []*User
}

type JobSeekerProfile struct {
	UserAccountID     int
	FirstName         string
	LastName          string
	City              string
	State             string
	Country           string
	WorkAuthorization string
	EmploymentType    string
	Resume            string
	ProfilePhoto      string
	Skills            *[]Skills
}

type RecruiterProfile struct {
	UserAccountID int
	FirstName     string
	LastName      string
	City          string
	State         string
	Country       string
	Company       string
	ProfilePhoto  string
}

type Skills struct {
	ID                int
	Name              string
	ExperienceLevel   string
	YearsOfExperience string
	JobSeekerProfile  JobSeekerProfile
}

func (m *DBModel) GetAllUserTypes() ([]*UserType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select user_type_id, user_type_name from users_type order by user_type_name`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userTypes []*UserType
	for rows.Next() {
		var u UserType
		err := rows.Scan(
			&u.ID,
			&u.UserTypeName,
		)
		if err != nil {
			return nil, err
		}
		userTypes = append(userTypes, &u)
	}

	return userTypes, nil
}

func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var u User
	var ut UserType
	query := `
		select 
		    u.user_id, u.email, u.is_active, u.registration_date, u.user_type_id,
		    ut.user_type_name
		from 
		    users u
			left join users_type ut on u.user_type_id = ut.user_type_id
		where 
		    u.email = ?`

	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.IsActive,
		&u.RegistrationDate,
		&ut.ID,
		&ut.UserTypeName,
	)
	if err != nil {
		return u, err
	}

	u.UserType = &ut
	return u, nil
}

func (m *DBModel) AddUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `insert into users (email, is_active, password, registration_date, user_type_id)
		values(?, ?, ?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, stmt,
		u.Email,
		u.IsActive,
		hash,
		time.Now(),
		u.UserType.ID,
	)

	if err != nil {
		return err
	}

	// Mysql / mariadb does not support "returning id"
	// Use last_insert_id - The ID that was generated is maintained
	// in the server on a per-connection basis.
	query := "select LAST_INSERT_ID()"
	row := m.DB.QueryRowContext(ctx, query)
	err = row.Scan(&u.ID)
	if err != nil {
		return err
	}

	if u.UserType.ID == 1 {
		stmt = "insert into recruiter_profile (user_account_id) values (?)"
	} else {
		stmt = "insert into job_seeker_profile (user_account_id) values (?)"
	}
	_, err = m.DB.ExecContext(ctx, stmt,
		u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) Authenticate(email, password string) (int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var id int
	var userTypeId int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select user_id, user_type_id, password from users where email = ?", email)
	err := row.Scan(&id, &userTypeId, &hashedPassword)
	if err != nil {
		return id, userTypeId, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, 0, errors.New("incorrect credentials")
	} else if err != nil {
		return 0, 0, err
	}

	return id, userTypeId, nil
}

func (m *DBModel) GetRecruiterProfile(userID int) (RecruiterProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select user_account_id, company, city, state, country, first_name, last_name, coalesce(profile_photo, '') as profile_photo 
		from recruiter_profile 
		where user_account_id = ?
	`

	var rp RecruiterProfile
	row := m.DB.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&rp.UserAccountID,
		&rp.Company,
		&rp.City,
		&rp.State,
		&rp.Country,
		&rp.FirstName,
		&rp.LastName,
		&rp.ProfilePhoto,
	)
	if err != nil {
		return rp, err
	}

	return rp, nil
}

func (m *DBModel) GetJobSeekerProfile(userID int) (JobSeekerProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select user_account_id, city, state, country, employment_type, first_name, last_name, profile_photo, resume, work_authorization 
		from job_seeker_profile 
		where user_account_id = ?
	`

	var jp JobSeekerProfile
	row := m.DB.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&jp.UserAccountID,
		&jp.City,
		&jp.State,
		&jp.Country,
		&jp.EmploymentType,
		&jp.FirstName,
		&jp.LastName,
		&jp.ProfilePhoto,
		&jp.Resume,
		&jp.WorkAuthorization,
	)
	if err != nil {
		return jp, err
	}

	return jp, nil
}

func (m *DBModel) UpdateRecruiterProfile(p RecruiterProfile) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// update profile info
	stmt := `update recruiter_profile set company = ?, city = ?, state = ?, country = ?, first_name = ?, last_name = ?
		where user_account_id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, p.Company, p.City, p.State, p.Country, p.FirstName, p.LastName, p.UserAccountID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) UpdateRecruiterProfilePhoto(p RecruiterProfile) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "update recruiter_profile set profile_photo = ? where user_account_id = ?"
	_, err := m.DB.ExecContext(ctx, stmt, p.ProfilePhoto, p.UserAccountID)
	if err != nil {
		return err
	}

	return nil
}
