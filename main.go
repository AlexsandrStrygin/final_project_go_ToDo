package main

import (
    "log"
    "net/http"
    "os"

    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
    "github.com/joho/godotenv"
)

func main() {
    // Загрузка переменных окружения
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file")
    }

    // Получаем путь к файлу базы данных из переменной окружения
    dbFile := os.Getenv("TODO_DBFILE")
    if dbFile == "" {
        dbFile = "scheduler.db" // Если переменная не задана, используем значение по умолчанию
    }

    // Инициализация базы данных
    if err := initDatabase(dbFile); err != nil {
        log.Fatalf("Could not initialize database: %v", err)
    }

    // Соединение с базой данных
    db, err := sqlx.Connect("sqlite3", dbFile)
    if err != nil {
        log.Fatalf("Could not connect to database: %v", err)
    }
    defer db.Close()

    // Настройка маршрутов
    setupRoutes(db)

    // Получение порта из переменных окружения
    port := os.Getenv("PORT")
    if port == "" {
        port = "7540" // Установить значение по умолчанию
    }

    // Запуск HTTP-сервера
    log.Printf("Starting server on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

