package models

type TaskCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TaskCategories struct {
	TaskCategories []TaskCategory `json:"task_categories"`
}
