// Package handlers содержит HTTP-обработчики и работу с шаблонами.
package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"go-http-server/internal/models"
)

// pages — страницы, для каждой из которых строится свой набор шаблонов:
// layout.html + <страница>.html.
var pages = []string{"home", "about", "expenses", "expense_new"}

// Handler хранит шаблоны и список трат в памяти.
type Handler struct {
	tmpl map[string]*template.Template

	mu       sync.Mutex // запросы обрабатываются параллельно — срез нужно защищать
	expenses []models.Expense
	nextID   int
}

// New разбирает шаблоны из каталога dir. Ошибка здесь — при старте, а не при первом запросе.
func New(dir string) (*Handler, error) {
	funcs := template.FuncMap{
		"money": func(v float64) string { return fmt.Sprintf("%.2f ₽", v) },
	}

	tmpl := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		t, err := template.New("").Funcs(funcs).ParseFiles(
			filepath.Join(dir, "layout.html"),
			filepath.Join(dir, page+".html"),
		)
		if err != nil {
			return nil, fmt.Errorf("страница %q: %w", page, err)
		}
		tmpl[page] = t
	}
	return &Handler{tmpl: tmpl, nextID: 1}, nil
}

// render выполняет шаблон в буфер и только потом пишет ответ:
// если шаблон упадёт посередине, клиент получит 500, а не половину страницы.
func (h *Handler) render(w http.ResponseWriter, status int, page string, data any) {
	var buf bytes.Buffer
	if err := h.tmpl[page].ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Printf("ошибка шаблона %s: %v", page, err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// Home — главная страница.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "home", map[string]any{"Title": "Главная"})
}

// About — о проекте.
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "about", map[string]any{"Title": "О проекте"})
}

// Ping — проверка живости сервера (из этапа 1).
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "метод не разрешён", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "pong")
}

// ExpenseList показывает таблицу трат.
func (h *Handler) ExpenseList(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	list := make([]models.Expense, len(h.expenses))
	copy(list, h.expenses) // копия, чтобы шаблон не читал срез, который меняют другие запросы
	h.mu.Unlock()

	var total float64
	for _, e := range list {
		total += e.Amount
	}

	h.render(w, http.StatusOK, "expenses", map[string]any{
		"Title":    "Траты",
		"Expenses": list,
		"Total":    total,
	})
}

// formData — данные формы, чтобы при ошибке не терять введённое.
type formData struct {
	Amount      string
	Description string
	Date        string
}

// ExpenseNewForm показывает пустую форму добавления.
func (h *Handler) ExpenseNewForm(w http.ResponseWriter, r *http.Request) {
	h.renderForm(w, http.StatusOK, formData{Date: todayString()}, "")
}

func (h *Handler) renderForm(w http.ResponseWriter, status int, f formData, errMsg string) {
	h.render(w, status, "expense_new", map[string]any{
		"Title": "Новая трата",
		"Form":  f,
		"Error": errMsg,
	})
}

// ExpenseCreate обрабатывает форму: проверяет, добавляет трату и делает redirect (Post/Redirect/Get).
func (h *Handler) ExpenseCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректная форма", http.StatusBadRequest)
		return
	}

	f := formData{
		Amount:      r.FormValue("amount"),
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	}

	exp, errMsg := parseExpense(f)
	if errMsg != "" {
		// Ошибка ввода: показываем форму снова (без redirect) с кодом 400.
		h.renderForm(w, http.StatusBadRequest, f, errMsg)
		return
	}

	h.mu.Lock()
	exp.ID = h.nextID
	h.nextID++
	h.expenses = append(h.expenses, exp)
	h.mu.Unlock()

	// Успех: не рендерим страницу, а отправляем браузер на GET /expenses.
	// Обновление страницы после этого повторит GET, а не POST.
	http.Redirect(w, r, "/expenses", http.StatusSeeOther) // 303
}
