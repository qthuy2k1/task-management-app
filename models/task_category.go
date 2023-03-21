package models

import (
	"fmt"
	"net/http"
)

type TaskCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TaskCategoryList struct {
	TaskCategories []TaskCategory `json:"task_categories"`
}

func (t *TaskCategory) Bind(r *http.Request) error {
	if t.Name == "" {
		return fmt.Errorf("name is a required field")
	}
	return nil
}

func (*TaskCategoryList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (*TaskCategory) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
