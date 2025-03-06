package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// Структура для документации через swagger
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Структура для документации через swagger
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Handlers struct {
	DB *pgx.Conn
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHandlers(db *pgx.Conn) *Handlers {
	return &Handlers{DB: db}
}

func handleError(c *fiber.Ctx, statusCode int, msg string, err error) error {
	fmt.Printf("%s: %v\n", msg, err)
	return c.Status(statusCode).JSON(ErrorResponse{Error: msg})
}

func parseID(c *fiber.Ctx) (int, error) {
	id := c.Params("id")
	taskID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}

	return taskID, nil
}

// CreateTask создает новую задачу
// @Summary      Создать задачу
// @Description  Создает новую задачу в базе данных
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task body CreateTaskRequest true "Создаваемая задача"
// @Success      201  {object}  Task
// @Failure      400  {object}  ErrorResponse "Ошибка валидации"
// @Failure      500  {object}  ErrorResponse "Ошибка сервера"
// @Router       /tasks [post]
func (h *Handlers) CreateTask(c *fiber.Ctx) error {
	task := new(Task)

	if err := c.BodyParser(task); err != nil {
		return handleError(c, fiber.StatusBadRequest, "Couldn't parse body", err)
	}

	if task.Title == "" {
		return handleError(c, fiber.StatusBadRequest, "Title is required", nil)
	}

	query := "INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id, status, created_at, updated_at"
	err := h.DB.QueryRow(context.Background(), query, task.Title, task.Description).Scan(&task.ID, &task.Status, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Couldn't create task", err)
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

// GetTask возвращает список задач
// @Summary      Получить список задач
// @Description  Возвращает все задачи из базы данных
// @Tags         tasks
// @Produce      json
// @Success      200  {array}  Task
// @Failure      500  {object}  ErrorResponse "Ошибка сервера"
// @Router       /tasks [get]
func (h *Handlers) GetTasks(c *fiber.Ctx) error {
	rows, err := h.DB.Query(context.Background(), "SELECT id, title, description, status, created_at, updated_at FROM tasks")
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Couldn't get tasks", err)
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return handleError(c, fiber.StatusInternalServerError, "Couldn't scan task", err)
		}
		tasks = append(tasks, task)
	}

	return c.JSON(tasks)
}

// UpdateTask обновляет задачу
// @Summary      Обновить задачу
// @Description  Обновляет существующую задачу по ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path int true "ID задачи"
// @Param        task body UpdateTaskRequest true "Создаваемая задача"
// @Success      200  {object}  Task
// @Failure      400  {object}  ErrorResponse "Ошибка валидации"
// @Failure      404  {object}  ErrorResponse "Задача не найдена"
// @Failure      500  {object}  ErrorResponse "Ошибка сервера"
// @Router       /tasks/{id} [put]
func (h *Handlers) UpdateTask(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return handleError(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	task := new(Task)
	if err := c.BodyParser(task); err != nil {
		return handleError(c, fiber.StatusBadRequest, "Couldn't parse body", err)
	}

	if task.Title == "" {
		return handleError(c, fiber.StatusBadRequest, "Title is required", nil)
	}

	query := "UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = now() WHERE id = $4 RETURNING id, title, description, status, created_at, updated_at"
	err = h.DB.QueryRow(context.Background(), query, task.Title, task.Description, task.Status, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Couldn't update task", err)
	}

	return c.Status(fiber.StatusOK).JSON(task)
}

// DeleteTask удаляет задачу
// @Summary      Удалить задачу
// @Description  Удаляет задачу по ID
// @Tags         tasks
// @Param        id path int true "ID задачи"
// @Success      204 "Задача удалена"
// @Failure      400  {object}  ErrorResponse "Неверный ID"
// @Failure      404  {object}  ErrorResponse "Задача не найдена"
// @Failure      500  {object}  ErrorResponse "Ошибка сервера"
// @Router       /tasks/{id} [delete]
func (h *Handlers) DeleteTask(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return handleError(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	query := "DELETE FROM tasks WHERE id = $1"
	tag, err := h.DB.Exec(context.Background(), query, id)

	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Couldn't delete task", err)
	}

	if tag.RowsAffected() == 0 {
		return handleError(c, fiber.StatusNotFound, "Task not found", nil)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
