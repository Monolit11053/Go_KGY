package main

import (
	"fmt"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Добро пожаловать в «Учёт трат»!")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Expense Tracker — учебный проект на Go: веб-приложение для учёта личных трат.")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	// Явно проверяем метод: всё, кроме GET, получает 405.
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet) // по стандарту 405 сопровождается Allow
		http.Error(w, "метод не разрешён", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK) // 200 — можно не вызывать, это значение по умолчанию
	fmt.Fprint(w, "pong")
}

func main() {
	mux := http.NewServeMux()

	// "/{$}" — только корень; без {$} шаблон "/" ловил бы все пути.
	mux.HandleFunc("/{$}", homeHandler)
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/ping", pingHandler)

	addr := ":8080"
	log.Printf("сервер запущен: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
