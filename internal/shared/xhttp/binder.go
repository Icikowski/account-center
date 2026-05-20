package xhttp

import "github.com/go-chi/chi/v5"

// RouteBinder defines an interface for types that can bind routes to a [chi.Router].
type RouteBinder interface {
	// Bind registers routes to the provided [chi.Router].
	Bind(r chi.Router)
}
