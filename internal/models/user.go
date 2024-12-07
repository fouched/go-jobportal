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
