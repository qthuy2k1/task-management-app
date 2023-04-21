package controllers

import (
	"context"
	"encoding/csv"
	"os"
	"regexp"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
)

type TaskCategoryController struct {
	TaskCategoryRepository *repositories.TaskCategoryRepository
}

func NewTaskCategoryController(taskCategoryRepository *repositories.TaskCategoryRepository) *TaskCategoryController {
	return &TaskCategoryController{TaskCategoryRepository: taskCategoryRepository}
}

func (c *TaskCategoryController) GetAllTaskCategories(ctx context.Context) (models.TaskCategorySlice, error) {
	taskCategories, err := c.TaskCategoryRepository.GetAllTaskCategories(ctx)
	if err != nil {
		return taskCategories, err
	}
	return taskCategories, nil
}

func (c *TaskCategoryController) AddTaskCategory(taskCategory *models.TaskCategory, ctx context.Context) error {
	err := c.TaskCategoryRepository.AddTaskCategory(taskCategory, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskCategoryController) GetTaskCategoryByID(taskCategoryID int, ctx context.Context) (*models.TaskCategory, error) {
	taskCategory, err := c.TaskCategoryRepository.GetTaskCategoryByID(taskCategoryID, ctx)
	if err != nil {
		return taskCategory, err
	}
	return taskCategory, nil
}

func (c *TaskCategoryController) DeleteTaskCategory(taskCategoryID int, ctx context.Context) error {
	err := c.TaskCategoryRepository.DeleteTaskCategory(taskCategoryID, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskCategoryController) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, ctx context.Context) (*models.TaskCategory, error) {
	taskCategory, err := c.TaskCategoryRepository.GetTaskCategoryByID(taskCategoryID, ctx)
	if err != nil {
		return taskCategory, err
	}
	taskCategory.Name = taskCategoryData.Name

	taskCategoryUpdated, err := c.TaskCategoryRepository.UpdateTaskCategory(taskCategory, ctx)
	if err != nil {
		return taskCategoryUpdated, err
	}
	return taskCategoryUpdated, nil
}

// Import task categories data from a CSV file
func (re *TaskCategoryController) ImportTaskCategoryDataFromCSV(path string) ([]models.TaskCategory, error) {
	// Create a slice to store the taskCategory data
	taskCategoryList := []models.TaskCategory{}

	// Open the CSV file
	file, err := os.Open("./data/task-category.csv")
	if err != nil {
		return taskCategoryList, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true // Trim leading spaces around fields

	// Read all the records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return taskCategoryList, err
	}

	// Loop through the records and create a new TaskCategory struct for each record
	for i, record := range records {
		// Skip the first index, which is the header name
		if i == 0 {
			continue
		}

		// remove unexpected characters
		re := regexp.MustCompile("[^a-zA-Z0-9]+")
		name := re.ReplaceAllString(record[0], "")

		taskCategory := models.TaskCategory{
			Name: name,
		}
		taskCategoryList = append(taskCategoryList, taskCategory)
	}

	return taskCategoryList, nil
}

func (c *TaskCategoryController) GetTasksByCategory(taskCategoryID int, ctx context.Context) (models.TaskSlice, error) {
	tasks, err := c.TaskCategoryRepository.GetTasksByCategory(taskCategoryID, ctx)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}
