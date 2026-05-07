package api

import (
	"Yandex_final/pkg/auth"
	"net/http"
	"time"
)

import "github.com/go-chi/chi/v5"

func Init(r *chi.Mux) {
	// Публичные маршруты (без проверки токена)
	r.Get("/api/nextdate", nextDateHandler)
	r.Post("/api/signin", signInHandler)

	// Защищенные маршруты
	r.Group(func(r chi.Router) {
		// Применяем middleware ко всем роутам внутри этой группы
		r.Use(auth.Middleware)
		// Создание задач
		r.Post("/api/task", addTaskHandler)

		// Получение списка задач
		r.Get("/api/tasks", tasksHandler)

		// Отметка о выполнении
		r.Post("/api/task/done", doneHandler)
		// Получение задачи
		r.Get("/api/task", taskHandler)
		// Редактирование задачи
		r.Put("/api/task", taskHandler)
		//Удаление задачи
		r.Delete("/api/task", taskHandler)
	})
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	var nowTime time.Time
	var err error

	if nowStr == "" {
		nowTime = time.Now()
	} else {
		nowTime, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, "Invalid 'now' date format. Use YYYYMMDD", http.StatusBadRequest)
			return
		}
	}

	if dateStr == "" || repeatStr == "" {
		http.Error(w, "Missing 'date' or 'repeat' parameters", http.StatusBadRequest)
		return
	}

	nextDateStr, err := NextDate(nowTime, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDateStr))
}
