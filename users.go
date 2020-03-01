package main

// User repesents the stored user
type User struct {
	Email        string // primary key
	FirstName    string
	LastName     string
	Password     string
	CreatedDate  string
	ModifiedDate string
}

// Users is a slice of User
type Users []User

var users = []User{
	User{Email: "john@ecn.com", FirstName: "John", LastName: "Tester", Password: "password"},
	User{Email: "jane@ecn.com", FirstName: "Jane", LastName: "Doe", Password: "password"},
}
