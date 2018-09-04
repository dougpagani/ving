package slices

// Repeat returns a new slice consisting of count copies of the item s.
//
// It panics if count is negative
func Repeat(item interface{}, count int) []interface{} {
	if count < 0 {
		panic("strings: negative Repeat count")
	}

	s := make([]interface{}, count)
	for idx := range s {
		s[idx] = item
	}
	return s
}
