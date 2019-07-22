package citool

// IsEmpty is a helper function which returns true if value is nil or empty
func IsEmpty(value *string) bool {
	return value == nil || len(*value) == 0
}
