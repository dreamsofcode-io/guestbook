package app

import (
	"html/template"
	"net/http"

	"github.com/dreamsofcode-io/guestbook/internal/handler"
)

func (a *App) loadRoutes() {
	tmpl := template.Must(template.New("").ParseGlob("./templates/*"))
	guestbook := handler.New(a.logger, a.db, tmpl)

	files := http.FileServer(http.Dir("./static"))
	a.router.Handle("GET /static/", http.StripPrefix("/static", files))

	a.router.HandleFunc("GET /{$}", guestbook.Home)

	a.router.HandleFunc("POST /{$}", guestbook.Create)
}
