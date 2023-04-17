package db

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	models "github.com/qthuy2k1/task-management-app/models/gen"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Gets all task categories from the database
func (db Database) GetAllTaskCategories(ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (models.TaskCategorySlice, error) {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return models.TaskCategorySlice{}, err
	}

	if !isManager {
		return models.TaskCategorySlice{}, errors.New("you are not the manager")
	}

	taskCategories, err := models.TaskCategories().All(ctx, db.Conn)
	if err != nil {
		return taskCategories, err
	}
	return taskCategories, nil
}

// Adds a new task category to the database
func (db Database) AddTaskCategory(taskCategory *models.TaskCategory, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	if taskCategory.Name == "" {
		return errors.New("bad request")
	}
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	err = taskCategory.Insert(ctx, db.Conn, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// Gets a task category from the database by ID
func (db Database) GetTaskCategoryByID(taskCategoryID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (*models.TaskCategory, error) {
	taskCategory := &models.TaskCategory{}
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return taskCategory, err
	}
	if !isManager {
		return taskCategory, errors.New("you are not the manager")
	}
	taskCategory, err = models.TaskCategories(Where("id = ?", taskCategoryID)).One(ctx, db.Conn)
	if err != nil {
		return taskCategory, err
	}
	return taskCategory, nil
}

// Deletes a task category from the database by ID
func (db Database) DeleteTaskCategory(taskCategoryID int, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) error {
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("you are not the manager")
	}
	rowsAff, err := models.TaskCategories(Where("id = ?", taskCategoryID)).DeleteAll(ctx, db.Conn)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoMatch
	}
	return err
}

// Updates a task category in the database by ID
func (db Database) UpdateTaskCategory(taskCategoryID int, taskCategoryData models.TaskCategory, ctx context.Context, r *http.Request, tokenAuth *jwtauth.JWTAuth, token jwt.Token) (*models.TaskCategory, error) {
	taskCategory := &models.TaskCategory{}
	isManager, err := db.IsManager(ctx, r, tokenAuth, token)
	if err != nil {
		return taskCategory, err
	}
	if !isManager {
		return taskCategory, errors.New("you are not the manager")
	}
	taskCategory, err = models.TaskCategories(Where("id = ?", taskCategoryID)).One(ctx, db.Conn)
	if err != nil {
		return taskCategory, err
	}
	taskCategory.Name = taskCategoryData.Name

	rowsAff, err := taskCategory.Update(ctx, db.Conn, boil.Infer())
	if err != nil {
		return taskCategory, err
	}
	if rowsAff == 0 {
		return taskCategory, ErrNoMatch
	}
	return taskCategory, nil
}

// // Import task categories data from a CSV file
// func (db Database) ImportTaskCategoryDataFromCSV(path string) (models.TaskCategoryList, error) {
// 	// Create a slice to store the taskCategory data
// 	taskCategoryList := models.TaskCategoryList{}

// 	// Open the CSV file
// 	file, err := os.Open("./data/task-category.csv")
// 	if err != nil {
// 		return taskCategoryList, err
// 	}
// 	defer file.Close()

// 	// Create a new CSV reader
// 	reader := csv.NewReader(file)
// 	reader.TrimLeadingSpace = true // Trim leading spaces around fields

// 	// Read all the records from the CSV file
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		return taskCategoryList, err
// 	}

// 	// Loop through the records and create a new TaskCategory struct for each record
// 	for i, record := range records {
// 		// Skip the first index, which is the header name
// 		if i == 0 {
// 			continue
// 		}

// 		// remove unexpected characters
// 		re := regexp.MustCompile("[^a-zA-Z0-9]+")
// 		name := re.ReplaceAllString(record[0], "")

// 		taskCategory := models.TaskCategory{
// 			Name: name,
// 		}
// 		taskCategoryList.TaskCategories = append(taskCategoryList.TaskCategories, taskCategory)
// 	}

// 	return taskCategoryList, nil
// }
