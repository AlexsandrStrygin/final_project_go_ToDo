package main

import (
    "database/sql"
    "github.com/jmoiron/sqlx"
)

// Scheduler представляет собой структуру таблицы
type Scheduler struct {
    ID      int64  `db:"id"`       // Храним id как int64 для базы данных
    IDStr   string `json:"id"`      // Отправляем id как string в JSON
    Date    string `db:"date" json:"date"`
    Title   string `db:"title" json:"title"`
    Comment string `db:"comment" json:"comment,omitempty"`
    Repeat  string `db:"repeat" json:"repeat,omitempty"`
}

// initDatabase инициализирует базу данных и создаёт таблицу
func initDatabase(dbFile string) error {
    db, err := sqlx.Connect("sqlite3", dbFile)
    if err != nil {
        return err
    }
    defer db.Close()

    // Создаём таблицу, если она не существует
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT
        )
    `)

    return err
}

// AddScheduler добавляет новую задачу в таблицу
func AddScheduler(db *sqlx.DB, s Scheduler) (sql.Result, error) {
    return db.NamedExec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`, s)
}
