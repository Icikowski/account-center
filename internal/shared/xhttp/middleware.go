package xhttp

import "net/http"

// Middleware is a type that represents an HTTP middleware function.
type Middleware func(next http.Handler) http.Handler
