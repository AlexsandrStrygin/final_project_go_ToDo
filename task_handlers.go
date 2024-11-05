package main

import (
    "net/http"
    "time"

    "github.com/jmoiron/sqlx"
)

// Обработчик для выполнения задачи
func DoneTaskHandler(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            writeError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

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

        // Если задача одноразовая, удаляем её
        if task.Repeat == "" {
            _, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
            if err != nil {
                writeError(w, "Ошибка при удалении задачи: "+err.Error(), http.StatusInternalServerError)
                return
            }
            writeResponse(w, map[string]struct{}{}, http.StatusOK) // Возвращаем пустой JSON
            return
        }

        // Если задача периодическая, рассчитываем следующую дату
        now := time.Now()
        nextDate, err := NextDate(now, task.Date, task.Repeat)
        if err != nil {
            writeError(w, "Ошибка при вычислении следующей даты: "+err.Error(), http.StatusBadRequest)
            return
        }

        // Обновляем задачу с новой датой
        _, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
        if err != nil {
            writeError(w, "Ошибка при обновлении задачи: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Успешное выполнение
        writeResponse(w, map[string]struct{}{}, http.StatusOK) // Возвращаем пустой JSON
    }
}

// Обработчик для удаления задачи
func DeleteTaskHandler(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodDelete {
            writeError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

        id := r.URL.Query().Get("id")
        if id == "" {
            writeError(w, "Не указан идентификатор", http.StatusBadRequest)
            return
        }

        // Удаление задачи из базы данных
        result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
        if err != nil {
            writeError(w, "Ошибка при удалении задачи: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Проверяем, были ли удалены строки
        rowsAffected, err := result.RowsAffected()
        if err != nil {
            writeError(w, "Ошибка при проверке количества удаленных строк: "+err.Error(), http.StatusInternalServerError)
            return
        }

        if rowsAffected == 0 {
            writeError(w, "Задача не найдена", http.StatusNotFound) // Возвращаем ошибку, если задача не найдена
            return
        }

        // Возвращаем пустой JSON
        writeResponse(w, map[string]struct{}{}, http.StatusOK) // Пустой JSON {}
    }
}
