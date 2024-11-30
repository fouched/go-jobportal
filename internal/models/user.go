package models

import (
	"context"
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
