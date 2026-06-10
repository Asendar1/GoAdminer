package handler

import (
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
