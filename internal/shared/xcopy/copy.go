package xcopy

import "github.com/tiendc/go-deepcopy"

// DeepCopy creates a deep copy of the provided value.
func DeepCopy[T any](src *T) (*T, error) {
	dst := new(T)
	err := deepcopy.Copy(dst, src)
	return dst, err
}
