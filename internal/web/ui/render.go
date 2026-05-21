package ui

import (
	"bytes"
	"net/http"

	"github.com/a-h/templ"
	"github.com/rs/zerolog"
)

func renderTemplate(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	l := zerolog.Ctx(r.Context())

	buffer := new(bytes.Buffer)
	if err := c.Render(r.Context(), buffer); err != nil {
		l.Error().Err(err).Msg("failed to render template")
		http.Error(w, "failed to render template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := buffer.WriteTo(w); err != nil {
		l.Error().Err(err).Msg("failed to write rendered template to response")
	}
}
