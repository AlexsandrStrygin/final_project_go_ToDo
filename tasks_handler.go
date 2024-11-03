package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

// JsonTask - временная структура для возврата в формате JSON
type JsonTask struct {
	ID      string `json:"id"`      // ID как строка
	Date    string `json:"date"`    // Дата задачи
	Title   string `json:"title"`   // Заголовок задачи
	Comment string `json:"comment"` // Комментарий задачи, если есть
	Repeat  string `json:"repeat"`  // Правило повторения, если есть
}

// Получает форматированную дату
func formatDate(dateStr string) (string, error) {
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return "", err
	}
	return date.Format("20060102"), nil
}

// GetTasksHandler получает список задач
func GetTasksHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		limit := 50 // Максимальное количество задач для возврата

		var tasks []Scheduler
		var err error

		if search != "" {
			// Проверка, является ли строка поиска датой
			formattedDate, formatErr := formatDate(search) // Используем другую переменную для ошибки
			if formatErr == nil {
				// Используем обычное сравнение для даты
				err = db.Select(&tasks, `SELECT * FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`, formattedDate, limit)
			} else {
				// Поиск по заголовку и комментарию
				searchPattern := "%" + search + "%"
				err = db.Select(&tasks, `SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`, searchPattern, searchPattern, limit)
			}
		} else {
			err = db.Select(&tasks, `SELECT * FROM scheduler ORDER BY date LIMIT ?`, limit)
		}

		if err != nil {
			writeError(w, "Ошибка при получении задач: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Если tasks == nil, создаем пустой слайс
		if tasks == nil {
			tasks = []Scheduler{}
		}

		// Формируем ответ
		jsonTasks := make([]JsonTask, len(tasks))
		for i, task := range tasks {
			jsonTasks[i] = JsonTask{
				ID:      strconv.FormatInt(task.ID, 10), // ID как строка
				Date:    task.Date,
				Title:   task.Title,
				Comment: task.Comment,
				Repeat:  task.Repeat,
			}
		}

		writeResponse(w, map[string]interface{}{"tasks": jsonTasks}, http.StatusOK)
	}
}
