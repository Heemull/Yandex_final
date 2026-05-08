package api

import (
	"net/http"
	"time"

	"Yandex_final/pkg/db"
)

// doneHandler изменение статуса задачи на Выполнено
func doneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	now := time.Now()

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete task"})
			return
		}
	} else {
		nextDateStr, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid repeat rule: " + err.Error()})
			return
		}

		err = db.UpdateDate(id, nextDateStr)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update task date"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}
