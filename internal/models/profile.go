package models

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
