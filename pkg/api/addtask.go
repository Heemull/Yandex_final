package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"Yandex_final/pkg/db"
)

// writeJSON отправляет JSON-ответ
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// addTaskHandler обрабатывает POST /api/task
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Явно используем UTC, чтобы избежать сдвигов из-за локального пояса сервера/теста
	now := time.Now().UTC()
	todayStr := now.Format("20060102")
	todayTime, _ := time.Parse("20060102", todayStr)

	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Title is required"})
		return
	}

	if task.Date == "" {
		task.Date = todayStr
	}

	taskTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid date format. Use YYYYMMDD"})
		return
	}
	taskTime = taskTime.UTC()

	if task.Repeat != "" {
		nextDateStr, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid repeat rule: " + err.Error()})
			return
		}

		if taskTime.Before(todayTime) {
			task.Date = nextDateStr
		}
	} else {
		if taskTime.Before(todayTime) {
			task.Date = todayStr
		}
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save task"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": fmt.Sprintf("%d", id)})
}
