package models

import (
	"net/http"
)

type UserTaskDetail struct {
	UserID int `json:"user_id"`
	TaskID int `json:"task_id"`
}

type UserTaskDetailList struct {
	UserTaskDetails []UserTaskDetail `json:"user_task_details"`
}

func (u *UserTaskDetail) Bind(r *http.Request) error {
	return nil
}

func (*UserTaskDetailList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*UserTaskDetail) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
