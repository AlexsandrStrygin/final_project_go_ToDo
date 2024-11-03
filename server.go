package main

import (
    "html/template"
    "net/http"
    "os"
    "log"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
)

// Эта функция настраивает маршруты
func setupRoutes(db *sqlx.DB) {
    // Обработчик корневого маршрута
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/scheduler", http.StatusFound) // Перенаправление на /scheduler
    })
    
    // Обработчик для выполнения задачи
    http.HandleFunc("/api/task/done", DoneTaskHandler(db))

    // Основной обработчик для маршрута /api/task
    http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            GetTaskHandler(db).ServeHTTP(w, r)   // Обработчик для получения одной задачи
        case http.MethodPost:
            AddTaskHandler(db).ServeHTTP(w, r)   // Обработчик для добавления задачи
        case http.MethodPut:
            UpdateTaskHandler(db).ServeHTTP(w, r) // Обработчик для обновления задачи
        case http.MethodDelete:
            DeleteTaskHandler(db).ServeHTTP(w, r) // Обработчик для удаления задачи
        default:
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
        }
    })

    http.HandleFunc("/api/nextdate", nextDateHandler)
    http.HandleFunc("/api/tasks", GetTasksHandler(db)) // Обработчик для получения всех задач
    http.HandleFunc("/scheduler", SchedulerHandler(db)) // Обработчик для страницы планировщика
    http.Handle("/favicon.ico", http.FileServer(http.Dir("web/"))) // Обработчик для favicon

    // Обработка статических файлов CSS
    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css"))))

    // Обработка статических файлов JS
    http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js"))))
}

// Обработчик для маршрута /scheduler
func SchedulerHandler(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Логирование пути и метода запроса
        log.Printf("Received request for: %s with method: %s", r.URL.Path, r.Method)

        if r.Method == http.MethodGet {
            // Загрузка данных из базы данных
            var tasks []Scheduler
            err := db.Select(&tasks, "SELECT * FROM scheduler")
            if err != nil {
                log.Printf("Error getting tasks from database: %v", err) // Логирование ошибки получения задач
                http.Error(w, "Ошибка при получении задач", http.StatusInternalServerError)
                return
            }

            // Чтение HTML-шаблона из файла
            tmplFile, err := os.ReadFile("web/index.html") // Укажите путь к вашему HTML-файлу
            if err != nil {
                log.Printf("Error reading template file: %v", err) // Логирование ошибки чтения шаблона
                http.Error(w, "Ошибка при чтении шаблона", http.StatusInternalServerError)
                return
            }

            tmpl, err := template.New("scheduler").Parse(string(tmplFile))
            if err != nil {
                log.Printf("Error creating template: %v", err) // Логирование ошибки создания шаблона
                http.Error(w, "Ошибка при создании шаблона", http.StatusInternalServerError)
                return
            }

            // Отправка данных в шаблон
            err = tmpl.Execute(w, tasks)
            if err != nil {
                log.Printf("Error executing template: %v", err) // Логирование ошибки при отображении данных
                http.Error(w, "Ошибка при отображении данных", http.StatusInternalServerError)
            }
        } else {
            log.Printf("Unsupported method: %s", r.Method) // Логирование неподдерживаемого метода
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
        }
    }
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "некорректная текущая дата", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

        // Отправка результата
        w.Header().Set("Content-Type", "text/plain")
        fmt.Fprint(w, nextDate)
}
