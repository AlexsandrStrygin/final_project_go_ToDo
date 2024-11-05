package main

// Task представляет собой структуру задачи
type Task struct {
    ID      string `json:"id"`      // ID как строка
    Date    string `json:"date"`    // Дата задачи
    Title   string `json:"title"`   // Заголовок задачи
    Comment string `json:"comment,omitempty"`  // Комментарий задачи, если есть
    Repeat  string `json:"repeat,omitempty"`   // Правило повторения, если есть
}
