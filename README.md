# Task Management Application
### Introduction
This is a simple API-based task management application that was created as part of an internship assignment. The application is designed to allow users to create, view, update, and delete tasks, as well as manage task categories. The application was developed using the Go programming language and the Chi router library, and it uses a PostgreSQL database for data storage.
### Project Support Features
* Users can signup and login to their accounts
* Public (non-authenticated) users can only access the homepage
* Authenticated users can access all tasks as well as edit their assigned tasks and also edit their information.
* Users who have the role of 'manager' are able to access all features within the app.
### Start the project guide
1. Clone this repository
    ```sh
    git clone https://github.com/qthuy2k1/task-management-app.git
    ```
2. Create an .env file in your project root folder and add your variables. See .env.sample for assistance.

3. Start the project and get all dependencies
    ```sh 
    docker-compose up
    ``` 
    If you have installed Makefile, you can simply run the <code>make</code> command.
    ```sh
    make
    ```
### API Endpoints
| HTTP Verbs | Endpoints | Action |
| --- | --- | --- |
| | HOMEPAGE |
| GET | / | Homepage |
| | USERS |
| POST | /signup | To sign up a new user account |
| POST | /login | To login an existing user account |
| POST | /logout | To log out of an account |
| GET | /users/ | To retrieve all users |
| GET | /users/profile | To retrieve the information of user account |
| POST | /users/change-password | To change the user account password |
| GET | /users/managers | To retrieve all users account that have the role of manager |
| GET | /users/{userID}/ | To retrieve the details of a single user |
| PUT | /users/{userID}/ | To update the information of user account |
| DELETE | /users/{userID}/ | To delete a user account |
| PATCH | /users/{userID}/update-role | To update the role of an user account |
| POST | /users/{userID}/get-tasks | To get all tasks that are assigned to a user |
| | TASKS |
| GET | /tasks/ | To retrieve all tasks |
| POST | /tasks | To add a new task to the database |
| POST | /tasks/csv | To import task data from a CSV file |
| GET | /tasks/filter-name | To retrieve all tasks filtering by name |
| GET | /tasks/{taskID}/ | To retrieve the details of a single task |
| PUT | /tasks/{taskID}/ | To update a task |
| DELETE | /tasks/{taskID}/ | To delete a task |
| PATCH | /tasks/{taskID}/lock | To lock a task |
| PATCH | /tasks/{taskID}/unlock | To unlock a task |
| POST | /tasks/{taskID}/add-user | To assign an user to a task |
| POST | /tasks/{taskID}/delete-user | To delete an user from a task |
| GET | /tasks/{taskID}/get-users | To retrieve all users that are assigned to a task |
| GET | /tasks/{taskID}/get-task-category | To retrieve the task category of a task |
| | TASK CATEGORIES |
| GET | /task-categories/ | To retrieve all task categories |
| POST | /task-categories | To add a new task category to the database |
| POST | /task-categories/csv | To import task category data from a CSV file |
| GET | /task-categories/{taskCategoryID}/ | To retrieve the details of a single task category |
| PUT | /task-categories/{taskCategoryID}/ | To update a task category |
| DELETE | /task-categories/{taskCategoryID}/ | To delete a task category |
### Technologies Used
* [Go](https://go.dev/) This is a simple and efficient programming language created by Google in 2007. It is known for its high performance and built-in support for concurrency.
* [Chi](https://go-chi.io/) A lightweight, idiomatic and composable router for building Go HTTP services.
* [Golang-migrate](https://github.com/golang-migrate/migrate) A popular package for managing database migrations in Golang projects.
* [PostgreSQL](https://www.postgresql.org/) PostgreSQL is a powerful, open source object-relational database system with over 35 years of active development that has earned it a strong reputation for reliability, feature robustness, and performance.
* [Sqlboiler](https://github.com/volatiletech/sqlboiler) A tool to generate a Go ORM tailored to your database schema. It is a "database-first" ORM as opposed to "code-first" (like gorm/gorp).
* [Docker](https://www.docker.com/) An open platform for developing, shipping, and running applications.