package app

import (
	"html/template"
	"net/http"

	"github.com/dreamsofcode-io/guestbook/internal/handler"
)

func (a *App) loadRoutes(tmpl *template.Template) {
	guestbook := handler.New(a.logger, a.db, tmpl)

	files := http.FileServer(http.Dir("./static"))

	a.router.Handle("GET /static/", http.StripPrefix("/static", files))

	a.router.Handle("GET /{$}", http.HandlerFunc(guestbook.Home))

	a.router.Handle("POST /{$}", http.HandlerFunc(guestbook.Create))
}
