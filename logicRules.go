package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
        "sort"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	fmt.Printf("Now: %s, Date: %s, Repeat: %s\n", now.Format(DateFormat), date, repeat)

	parsedDate, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", errors.New("некорректная дата")
	}

	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	parts := strings.Fields(repeat)
	var nextDate time.Time

        switch parts[0] {
	case "y":
		for {
			nextDate = parsedDate.AddDate(1, 0, 0) // Добавляем 1 год к дате
			if nextDate.After(now) {
				break // Прерываем цикл, если nextDate больше now
			}
			parsedDate = nextDate // Обновляем parsedDate на следующую дату
		}

	case "d":
		if len(parts) != 2 {
			return "", errors.New("некорректный формат правила 'd'")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 { // лимит в 400 дней
			return "", fmt.Errorf(`некорректное правило повторения в "d"`)
		}

		nextDate = parsedDate
		// Найти первую дату больше текущей
		nextDate = nextDate.AddDate(0, 0, days)

		// Цикл для поиска следующей даты
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}

	case "w":
		return handleWeekRule(parsedDate, now, parts)

	case "m":
		return handleMonthRule(parsedDate, now, parts)

	default:
		return "", errors.New("неподдерживаемый формат")
	}

	if !nextDate.After(now) {
		return "", nil
	}

	nextDateFormatted := nextDate.Format(DateFormat)
	fmt.Printf("Calculated next date: %s\n", nextDateFormatted)
	return nextDateFormatted, nil
}

// Обработка недельного правила "w"
func handleWeekRule(date time.Time, now time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("некорректный формат правила 'w'")
	}

	// Разбиение входного параметра на дни недели
	days := strings.Split(parts[1], ",")
	nextDate := date

	// Если заданная дата раньше текущей
	if nextDate.Before(now) {
		nextDate = now
	}

	// Перебор дней недели от 1 до 7
	for _, day := range days {
		dayInt, err := strconv.Atoi(day)
		if err != nil || dayInt < 1 || dayInt > 7 {
			return "", errors.New("некорректный день недели: " + day)
		}

		// Поиск ближайшего нужного дня
		for {
			weekDay := int(nextDate.Weekday())
			if weekDay == 0 { // Если воскресенье
				weekDay = 7
			}

			if weekDay == dayInt && nextDate.After(now) {
				// Если совпадает день недели и дата после now
				return nextDate.Format(DateFormat), nil
			}

			// Переход к следующему дню
			nextDate = nextDate.AddDate(0, 0, 1) // Добавляем 1 день
		}
	}

	return "", nil
}

// handleMonthRule обрабатывает месячные правила m
// handleMonthRule обрабатывает месячные правила и возвращает ближайшую подходящую дату
func handleMonthRule(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", fmt.Errorf("неверный формат повторения для правила 'm'")
	}

	if date.Before(now) {
		date = now // Устанавливаем на текущее время, если дата меньше
	}

	var days []int
	var months []int

	// Обработка дней
	for _, d := range strings.Split(parts[1], ",") {
		day, err := strconv.Atoi(d)
		if err != nil || (day < -2 || day > 31) {
			return "", fmt.Errorf("некорректный день месяца: %s", d)
		}
		days = append(days, day)
	}

	// Обработка месяцев
	if len(parts) > 2 {
		for _, m := range strings.Split(parts[2], ",") {
			month, err := strconv.Atoi(m)
			if err != nil || month < 1 || month > 12 {
				return "", fmt.Errorf("некорректный номер месяца: %s", m)
			}
			months = append(months, month)
		}
	} else {
		// Если месяцы не указаны, добавляем все месяцы
		for month := 1; month <= 12; month++ {
			months = append(months, month)
		}
	}

	validDates := make([]time.Time, 0) // Хранение валидных дат

	// Переход по месяцам и дням для поиска валидных дат
	for _, month := range months {
		for _, d := range days {
			var targetDate time.Time // для следующего значения

			if d == -1 { // Последний день месяца
				targetDate = time.Date(date.Year(), time.Month(month+1), 1, 0, 0, 0, 0, date.Location()).AddDate(0, 0, -1)
			} else if d == -2 { // Предпоследний день месяца
				targetDate = time.Date(date.Year(), time.Month(month+1), 1, 0, 0, 0, 0, date.Location()).AddDate(0, 0, -2)
			} else { // Указанный день месяца
				if d > lastDayInMonth(time.Date(date.Year(), time.Month(month), 1, 0, 0, 0, 0, date.Location())) {
					continue // Пропустить, если день больше допустимого
				}
				targetDate = time.Date(date.Year(), time.Month(month), d, 0, 0, 0, 0, date.Location())
			}

			// Проверяем, что targetDate больше чем 'now'
			if targetDate.After(now) {
				validDates = append(validDates, targetDate) // Сохраняем только будущие даты
			}
		}
	}

	if len(validDates) == 0 {
		return "", errors.New("нет валидных дат")
	}

	// Возвращаем ближайшую валидную дату
	sort.Slice(validDates, func(i, j int) bool {
		return validDates[i].Before(validDates[j]) // Сортировка по времени
	})

	return validDates[0].Format(DateFormat), nil // Форматируем и возвращаем дату
}

// lastDayInMonth вычисляет последний день месяца с учетом високосного года
func lastDayInMonth(t time.Time) int {
	month := t.Month()
	year := t.Year()

	switch month {
	case time.February:
		if isLeapYear(year) {
			return 29
		}
		return 28
	case time.April, time.June, time.September, time.November:
		return 30
	default:
		return 31
	}
}

// isLeapYear определяет, является ли год високосным
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
