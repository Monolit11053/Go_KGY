// Команда server — точка входа приложения «Учёт трат».
// Запуск из корня репозитория: go run ./cmd/server
package main

import (
	"log"
	"net/http"

	"go-http-server/internal/handlers"
)

func main() {
	h, err := handlers.New("web/templates")
	if err != nil {
		log.Fatalf("не удалось загрузить шаблоны: %v", err)
	}

	mux := http.NewServeMux()

	// Статика: /static/style.css -> web/static/style.css
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	// "/{$}" — только корень; без {$} шаблон "/" ловил бы все пути.
	mux.HandleFunc("GET /{$}", h.Home)
	mux.HandleFunc("GET /about", h.About)
	mux.HandleFunc("/ping", h.Ping) // метод проверяется внутри (этап 1)

	mux.HandleFunc("GET /expenses", h.ExpenseList)
	mux.HandleFunc("GET /expenses/new", h.ExpenseNewForm)
	mux.HandleFunc("POST /expenses", h.ExpenseCreate)

	addr := ":8080"
	log.Printf("сервер запущен: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
