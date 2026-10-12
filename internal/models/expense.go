// Package models содержит структуры данных приложения.
package models

import "time"

// DateLayout — формат даты, который отдаёт <input type="date">.
const DateLayout = "2006-01-02"

// Expense — одна трата.
type Expense struct {
	ID          int
	Amount      float64 // сумма в рублях
	Description string
	Date        time.Time
}

// DateString возвращает дату в формате ГГГГ-ММ-ДД (удобно для шаблонов).
func (e Expense) DateString() string {
	return e.Date.Format(DateLayout)
}
