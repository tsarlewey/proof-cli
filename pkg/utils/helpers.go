package utils

// Ptr returns a pointer to the given value (helper for optional SDK fields).
func Ptr[T any](v T) *T {
	return &v
}

// PtrIfNotEmpty returns a pointer to the string if non-empty, otherwise nil.
func PtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
