package main

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "github.com/jmoiron/sqlx"
)

// Обработчик для получения задачи по идентификатору
func GetTaskHandler(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        if id == "" {
            writeError(w, "Не указан идентификатор", http.StatusBadRequest)
            return
        }

        var task Scheduler
        err := db.Get(&task, "SELECT * FROM scheduler WHERE id = ?", id)
        if err != nil {
            writeError(w, "Задача не найдена", http.StatusNotFound)
            return
        }

        // Создаем ответ в виде словаря
        response := map[string]interface{}{
            "id":      strconv.FormatInt(task.ID, 10),
            "date":    task.Date,
            "title":   task.Title,
            "comment": task.Comment,
            "repeat":  task.Repeat,
        }

        // Установка заголовка для JSON
        w.Header().Set("Content-Type", "application/json")
        // Ответ в формате JSON
        json.NewEncoder(w).Encode(response)
    }
}

// Обработчик для обновления задачи по идентификатору
func UpdateTaskHandler(db *sqlx.DB) http.HandlerFunc {
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
        today := now.Format("20060102")

        if task.Date == "" {
            task.Date = today
        }

        taskDate, err := time.Parse("20060102", task.Date)
        if err != nil {
            writeError(w, "Неверный формат даты, ожидается формат YYYYMMDD", http.StatusBadRequest)
            return
        }

        if taskDate.Equal(now) {
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

        // Обновление задачи в базе данных
        res, err := db.Exec(`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
            task.Date, task.Title, task.Comment, task.Repeat, task.ID)
        
        if err != nil {
            writeError(w, "Ошибка обновления задачи: "+err.Error(), http.StatusInternalServerError)
            return
        }

        rowsAffected, _ := res.RowsAffected()
        if rowsAffected == 0 {
            writeError(w, "Задача не найдена", http.StatusNotFound)
            return
        }

        // Успешное обновление
        writeResponse(w, map[string]string{}, http.StatusOK)
    }
}
