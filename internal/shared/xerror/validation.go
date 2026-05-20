package xerror

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError represents an error that occurs during validation, containing a message
// and optional list of underlying causes.
type ValidationError struct {
	msg    string
	causes []error
}

// NewValidationError creates a new [ValidationError] with the given message and optional causes.
func NewValidationError(msg string, causes ...error) error {
	return ValidationError{
		msg:    msg,
		causes: causes,
	}
}

// Error implements [error].
func (e ValidationError) Error() string {
	causes := e.joinCauses()
	if causes == nil {
		return e.msg
	}
	return fmt.Sprintf("%s: [%s]", e.msg, *causes)
}

func (e ValidationError) joinCauses() *string {
	if len(e.causes) == 0 {
		return nil
	}

	causes := make([]string, len(e.causes))
	for i, cause := range e.causes {
		causes[i] = cause.Error()
	}
	return new(strings.Join(causes, "; "))
}

// ItemValidationError represents an error that occurs during validation of a specific item.
//
// If the cause of the error is a [ValidationError], its message will be included in the error
// message along with the index of the item.
type ItemValidationError struct {
	index int
	cause error
}

// NewItemValidationError creates a new [ItemValidationError] for the item at the given index
// with the provided cause.
func NewItemValidationError(index int, cause error) error {
	if cause == nil {
		return nil
	}

	return ItemValidationError{
		index: index,
		cause: cause,
	}
}

// Error implements [error].
func (e ItemValidationError) Error() string {
	var ve ValidationError
	if errors.As(e.cause, &ve) {
		causes := ve.joinCauses()
		if causes == nil {
			return fmt.Sprintf("%s at index %d", ve.msg, e.index)
		}
		return fmt.Sprintf("%s at index %d: [%s]", ve.msg, e.index, *causes)
	}

	return fmt.Sprintf("at index %d: %s", e.index, e.cause.Error())
}
