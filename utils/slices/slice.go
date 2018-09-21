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

// ContainStr check whether `slice` contain `target`
func ContainStr(slice []string, target string) bool {
	return IndexStrOf(slice, target) >= 0
}

// IndexStrOf for `target` index in `slice`, return -1 if `target` not found
func IndexStrOf(slice []string, target string) int {
	for i, item := range slice {
		if item == target {
			return i
		}
	}
	return -1
}
