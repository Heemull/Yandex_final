package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"Yandex_final/pkg/api" // Добавляем импорт api
	"Yandex_final/pkg/db"
)

// loadEnv загружает переменные из файла .env.example
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded. Using system environment variables or defaults.")
	}
}

func Start() {
	loadEnv()

	dbFile := os.Getenv("TODO_DBFILE")
	fmt.Println(dbFile)
	// Инициализация бд
	err := db.Init(dbFile)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer func() {
		if db.DB != nil {
			db.DB.Close()
		}
	}()
	// Получаем порт (если не указан, то используем стандарный 7540)
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	// Получаем директорию с файлами фронта
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "./web"
	}

	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		log.Fatalf("Ошибка: директория статики '%s' не найдена", webDir)
	}

	r := chi.NewRouter()

	// Инициализация  роутов
	api.Init(r)

	fileServer := http.FileServer(http.Dir(webDir))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	addr := ":" + port
	log.Printf("Откройте в браузере http://127.0.0.1%s/", addr)

	// Запуск сервера
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
