package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"Yandex_final/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler получение всех задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {

	search := r.URL.Query().Get("search")

	limit := 50

	tasks, err := db.Tasks(limit, search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, TasksResp{
		Tasks: tasks,
	})
}

// taskHandler маршрутизирует запросы /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

// getTaskHandler получение определенной задачи
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// updateTaskHandler изменение задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if task.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required for update"})
		return
	}

	_, err = strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID format"})
		return
	}

	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Title is required"})
		return
	}

	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}

	taskTime, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid date format"})
		return
	}

	if task.Repeat != "" {
		nextDateStr, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid repeat rule: " + err.Error()})
			return
		}

		if !taskTime.After(now) {
			task.Date = nextDateStr
		}
	} else {
		if !taskTime.After(now) {
			task.Date = now.Format(DateFormat)
		}
	}

	err = db.UpdateTask(&task)
	if err != nil {
		if err.Error() == "task not found or no changes made" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update task"})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}

// deleteTaskHandler удаление задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}
