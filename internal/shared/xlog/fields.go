package xlog

// Common log fields.
const (
	FieldRequestID = "id"
	FieldMethod    = "method"
	FieldPath      = "path"
	FieldClientIP  = "client_ip"
	FieldDuration  = "duration"

	FieldVersion   = "version"
	FieldCommit    = "commit"
	FieldBuildTime = "build_time"
	FieldGoVersion = "go_version"
	FieldCause     = "cause"

	FieldSubject  = "subject"
	FieldServices = "services"
	FieldNext     = "next"
)
