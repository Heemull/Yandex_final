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
	var task db.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid date format. Use YYYYMMDD"})
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

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save task"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": fmt.Sprintf("%d", id)})
}
