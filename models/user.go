package models

import (
	"fmt"
	"net/http"
)

type User struct {
	ID       int    `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	TaskID   int    `json:"task_id"`
}

type UserList struct {
	Users []User `json:"users"`
}

func (u *User) Bind(r *http.Request) error {
	if u.Name == "" {
		return fmt.Errorf("name is a required field")
	}
	return nil
}

func (*UserList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
