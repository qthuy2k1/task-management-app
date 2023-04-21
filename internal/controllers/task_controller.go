package controllers

import (
	"context"
	"encoding/csv"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	models "github.com/qthuy2k1/task-management-app/internal/models/gen"
	"github.com/qthuy2k1/task-management-app/internal/repositories"
	"github.com/volatiletech/null/v8"
)

type TaskController struct {
	TaskRepository *repositories.TaskRepository
}

func NewTaskController(taskRepository *repositories.TaskRepository) *TaskController {
	return &TaskController{TaskRepository: taskRepository}
}

func (c *TaskController) GetAllTasks(ctx context.Context, filterValues map[string]interface{}) (models.TaskSlice, error) {
	tasks, err := c.TaskRepository.GetAllTasks(ctx, filterValues)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}

func (c *TaskController) AddTask(task *models.Task, ctx context.Context) error {
	err := c.TaskRepository.AddTask(task, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskController) GetTaskByID(taskID int, ctx context.Context) (*models.Task, error) {
	task, err := c.TaskRepository.GetTaskByID(taskID, ctx)
	if err != nil {
		return task, err
	}
	return task, nil
}

func (c *TaskController) DeleteTask(taskID int, ctx context.Context) error {
	err := c.TaskRepository.DeleteTask(taskID, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskController) UpdateTask(taskID int, taskData models.Task, ctx context.Context, isManager bool) (*models.Task, error) {
	task, err := c.TaskRepository.GetTaskByID(taskID, ctx)
	if err != nil {
		return task, err
	}

	task.Name = taskData.Name
	task.Description = taskData.Description
	task.StartDate = taskData.StartDate
	task.EndDate = taskData.EndDate
	task.Status = taskData.Status
	if task.Status.String == "Lock" && !isManager {
		return task, errors.New("you are not the manager, cannot lock this task")
	}
	task.AuthorID = taskData.AuthorID
	task.UpdatedAt = time.Now()
	task.TaskCategoryID = taskData.TaskCategoryID

	taskUpdated, err := c.TaskRepository.UpdateTask(task, ctx)
	if err != nil {
		return taskUpdated, err
	}
	return taskUpdated, nil
}

func (c *TaskController) LockTask(taskID int, ctx context.Context) error {
	err := c.TaskRepository.LockTask(taskID, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *TaskController) UnLockTask(taskID int, ctx context.Context) error {
	err := c.TaskRepository.UnLockTask(taskID, ctx)
	if err != nil {
		return err
	}
	return nil
}

// Import tasks data from a CSV file
func (c *TaskController) ImportTaskDataFromCSV(path string) ([]models.Task, error) {
	// Create a slice to store the task data
	taskList := []models.Task{}

	// Open the CSV file
	file, err := os.Open(path)
	if err != nil {
		return taskList, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	// Trim leading spaces around fields
	reader.TrimLeadingSpace = true

	// Read all the records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return taskList, err
	}

	// Layout for time.Time field in struct
	const layoutDate = "2006-01-02T15:04:05Z"

	// Loop through the records and create a new Task struct for each record
	for i, record := range records {
		// Skip the first index, which is the header name
		if i == 0 {
			continue
		}
		data := strings.Split(record[0], ",")
		// remove unexpected characters
		re := regexp.MustCompile("[^a-zA-Z0-9 ]+")

		// regex for Name
		name := re.ReplaceAllString(data[0], "")

		// regex for description
		description := re.ReplaceAllString(data[1], "")

		// convert start date from string to time.Time
		startDate, err := time.Parse(layoutDate, data[2])
		if err != nil {
			return taskList, err
		}

		// convert end date from string to time.Time
		endDate, err := time.Parse(layoutDate, data[3])
		if err != nil {
			return taskList, err
		}

		// regex for task status
		status := re.ReplaceAllString(data[4], "")

		// convert authorID from string to int
		authorID, err := strconv.Atoi(data[5])
		if err != nil {
			return taskList, err
		}

		// convert createdAt from string to time.Time
		createdAt, err := time.Parse(layoutDate, data[6])
		if err != nil {
			return taskList, err
		}

		// convert updatedAt from string to time.Time
		updatedAt, err := time.Parse(layoutDate, data[7])
		if err != nil {
			return taskList, err
		}

		// convert taskCategoryID from string to int
		taskCategoryID, err := strconv.Atoi(data[8])
		if err != nil {
			return taskList, err
		}

		task := models.Task{
			Name:           name,
			Description:    description,
			StartDate:      startDate,
			EndDate:        endDate,
			Status:         null.StringFrom(status),
			AuthorID:       authorID,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			TaskCategoryID: taskCategoryID,
		}
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (c *TaskController) GetTaskCategoryOfTask(taskID int, ctx context.Context) (*models.TaskCategory, error) {
	taskCategory, err := c.TaskRepository.GetTaskCategoryOfTask(taskID, ctx)
	if err != nil {
		return taskCategory, err
	}
	return taskCategory, err
}

func (c *TaskController) GetTasksByName(name string, ctx context.Context) (models.TaskSlice, error) {
	tasks, err := c.TaskRepository.GetTasksByName(name, ctx)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}

func (c *TaskController) GetTaskCount(ctx context.Context) (int64, error) {
	count, err := c.TaskRepository.GetTaskCount(ctx)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (c *TaskController) CountFilteredStatusTask(status string, ctx context.Context) (int64, error) {
	count, err := c.TaskRepository.CountFilteredStatusTask(status, ctx)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// func (c *TaskController) FilterTasks(filterValues map[string]string, ctx context.Context) (models.TaskSlice, error) {
// 	tasks, err := c.TaskRepository.FilterTasks(filterValues, ctx)
// 	if err != nil {
// 		return tasks, err
// 	}
// 	return tasks, nil
// }
