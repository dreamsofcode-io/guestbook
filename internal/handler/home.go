package handler

import (
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/dreamsofcode-io/guestbook/internal/guest"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Guestbook struct {
	logger *slog.Logger
	tmpl   *template.Template
	repo   *guest.Repo
}

func New(
	logger *slog.Logger, db *pgxpool.Pool, tmpl *template.Template,
) *Guestbook {
	repo := guest.NewRepo(db)
	return &Guestbook{
		tmpl:   tmpl,
		repo:   repo,
		logger: logger,
	}
}

type indexPage struct {
	Guests []guest.Guest
	Total  int
}

type errorPage struct {
	ErrorMessage string
}

func (h *Guestbook) Home(w http.ResponseWriter, r *http.Request) {
	guests, err := h.repo.FindAll(r.Context(), 20)
	if err != nil {
		h.logger.Error("failed to find guests", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	count, err := h.repo.Count(r.Context())
	if err != nil {
		h.logger.Error("failed to get count", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html")
	h.tmpl.ExecuteTemplate(w, "index.html", indexPage{
		Guests: guests,
		Total:  count,
	})
}

func (h *Guestbook) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Error("failed to parse form", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg, ok := r.Form["message"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message := strings.Join(msg, " ")

	if strings.TrimSpace(message) == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.tmpl.ExecuteTemplate(w, "error.html", errorPage{
			ErrorMessage: "Blank messages don't count",
		})

		return
	}

	splits := strings.Split(r.RemoteAddr, ":")
	ipStr := strings.Trim(strings.Join(splits[:len(splits)-1], ":"), "[]")
	ip := net.ParseIP(ipStr)

	if goaway.IsProfane(message) {
		w.WriteHeader(http.StatusBadRequest)
		h.tmpl.ExecuteTemplate(w, "error.html", errorPage{
			ErrorMessage: fmt.Sprintf(
				"Please don't use profanity. Your IP has been tracked %s",
				ipStr,
			),
		})
		return
	}

	guest, err := guest.NewGuest(message, ip)
	if err != nil {
		h.logger.Error("failed to create guest", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.repo.Insert(r.Context(), guest)
	if err != nil {
		h.logger.Error("failed to insert guest", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
