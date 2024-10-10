package handler

import (
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/x-way/crawlerdetect"

	"github.com/dreamsofcode-io/guestbook/internal/guest"
	"github.com/dreamsofcode-io/guestbook/internal/repository"
)

type Guestbook struct {
	logger *slog.Logger
	tmpl   *template.Template
	repo   *repository.Queries
}

func New(
	logger *slog.Logger, db *pgxpool.Pool, tmpl *template.Template,
) *Guestbook {
	return &Guestbook{
		tmpl:   tmpl,
		repo:   repository.New(db),
		logger: logger,
	}
}

type indexPage struct {
	Guests []repository.Guest
	Total  int64
}

type errorPage struct {
	ErrorMessage string
}

func (h *Guestbook) Home(w http.ResponseWriter, r *http.Request) {
	guests, err := h.repo.FindAll(r.Context(), 200)
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
	// if crawlerdetect.IsCrawler(r.Header.Get("User-Agent")) {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	return
	// }
	//
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

	_, err = h.repo.Insert(r.Context(), repository.InsertParams{
		ID:        guest.ID,
		Message:   guest.Message,
		CreatedAt: guest.CreatedAt,
		Ip:        guest.IP,
	})
	if err != nil {
		h.logger.Error("failed to insert guest", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
