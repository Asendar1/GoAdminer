package handler

import (
	"fmt"
	"net/http"

	"github.com/Asendar1/GoAdminer/internal/driver"
	"github.com/Asendar1/GoAdminer/internal/session"
)

type Handler struct {
	Sessions *session.Store
	Drivers  map[string]func() driver.Driver
}

func New(sessions *session.Store) *Handler {
	return &Handler{
		Sessions: sessions,
		Drivers: map[string]func() driver.Driver{
			"postgres": driver.NewPostgres,
			"sqlite":   driver.NewSQLite,
		},
	}
}

func (h *Handler) getSession(r *http.Request) (*session.Session, error) {
	id := r.Header.Get("X-Session-ID")
	if id == "" {
		return nil, fmt.Errorf("missing X-Session-ID header")
	}
	sess, ok := h.Sessions.Get(id)
	if !ok {
		return nil, fmt.Errorf("invalid or expired session")
	}
	return sess, nil
}

func (h *Handler) getDriver(name string) driver.Driver {
	fn, ok := h.Drivers[name]
	if !ok {
		return nil
	}
	return fn()
}
