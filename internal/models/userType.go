package models

type UserType struct {
	ID           int
	UserTypeName string
	Users        []*User
}
