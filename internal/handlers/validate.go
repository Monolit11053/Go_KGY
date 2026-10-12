package handlers

import (
	"math"
	"strconv"
	"strings"
	"time"

	"go-http-server/internal/models"
)

func todayString() string {
	return time.Now().Format(models.DateLayout)
}

// parseExpense превращает строки формы в Expense. При ошибке возвращает текст для пользователя.
func parseExpense(f formData) (models.Expense, string) {
	amountStr := strings.ReplaceAll(strings.TrimSpace(f.Amount), ",", ".") // «12,50» -> «12.50»
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return models.Expense{}, "Сумма должна быть положительным числом."
	}

	desc := strings.TrimSpace(f.Description)
	if desc == "" {
		return models.Expense{}, "Введите описание траты."
	}

	date, err := time.Parse(models.DateLayout, f.Date)
	if err != nil {
		return models.Expense{}, "Укажите дату в формате ГГГГ-ММ-ДД."
	}

	return models.Expense{Amount: amount, Description: desc, Date: date}, ""
}
