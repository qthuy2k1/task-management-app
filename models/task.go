package models

import (
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Status         string    `json:"status"`
	UserID         int       `json:"user_id"`
	AuthorID       int       `json:"author_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskCategoryID int       `json:"task_category_id"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}

func (t *Task) Bind(r *http.Request) error {
	if t.Name == "" {
		return fmt.Errorf("name is a required field")
	}
	return nil
}

func (*TaskList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*Task) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
