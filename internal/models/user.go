package models

import "time"

type User struct {
	ID               int
	Email            string
	Password         string
	IsActive         bool
	RegistrationDate time.Time
	UserType         *UserType
}
