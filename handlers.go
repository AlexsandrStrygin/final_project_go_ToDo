package main

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/jmoiron/sqlx"
)

// Функция для отправки ошибок клиенту
func writeError(w http.ResponseWriter, message string, status int) {
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Функция для отправки ответов клиенту
func writeResponse(w http.ResponseWriter, data interface{}, status int) {
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// Обработчик для добавления задачи
func AddTaskHandler(s *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var task Task
        err := json.NewDecoder(r.Body).Decode(&task)
        if err != nil {
            writeError(w, "Ошибка десериализации JSON: "+err.Error(), http.StatusBadRequest)
            return
        }

        if task.Title == "" {
            writeError(w, "Не указан заголовок задачи", http.StatusBadRequest)
            return
        }

        now := time.Now()
        today := now.Format(DateFormat)

        if task.Date == "" {
            task.Date = today
        }

        taskDate, err := time.Parse(DateFormat, task.Date)
        if err != nil {
            writeError(w, "Неверный формат даты, ожидается формат YYYYMMDD", http.StatusBadRequest)
            return
        }

        // Проверяем, является ли дата задачи сегодняшней
        if taskDate.Year() == now.Year() && taskDate.YearDay() == now.YearDay() {
            task.Date = today
        } else if taskDate.Before(now) {
            if task.Repeat == "" {
                task.Date = today
            } else {
                nextDate, err := NextDate(now, task.Date, task.Repeat)
                if err != nil {
                    writeError(w, "Ошибка при вычислении следующей даты: "+err.Error(), http.StatusBadRequest)
                    return
                }
                task.Date = nextDate
            }
        }

        // Добавление задачи в базу данных
        result, err := AddScheduler(s, Scheduler{
        	ID:      0,
        	Date:    task.Date,
        	Title:   task.Title,
        	Comment: task.Comment,
        	Repeat:  task.Repeat,
        })
        if err != nil {
            writeError(w, "Ошибка добавления задачи: "+err.Error(), http.StatusInternalServerError)
            return
        }

        id, err := result.LastInsertId()
        if err != nil {
            writeError(w, "Ошибка получения ID задачи: "+err.Error(), http.StatusInternalServerError)
            return
        }

        writeResponse(w, map[string]int64{"id": id}, http.StatusCreated)
    }
}
