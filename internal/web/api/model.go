package api

// SessionResponse describes the state of the caller's session.
type SessionResponse struct {
	Authenticated bool  `json:"authenticated"`
	User          *User `json:"user,omitempty"`
}

// User is the public view of an authenticated user.
type User struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture,omitempty"`
}

// ServicesResponse lists the services visible to the caller.
type ServicesResponse struct {
	Services []Service `json:"services"`
}

// Service is the public view of a catalog service.
type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon,omitempty"`
	Role string `json:"role"`
}

// ErrorResponse is returned for every failed API call.
type ErrorResponse struct {
	Error string `json:"error"`
}
