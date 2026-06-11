package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	sess, err := h.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	drv := h.getDriver(sess.Driver)
	if drv == nil {
		http.Error(w, "unknown driver", http.StatusInternalServerError)
		return
	}

	dbs, err := drv.ListDatabases(sess.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbs)
}
