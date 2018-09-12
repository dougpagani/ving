package statistic

import "math"

// SortStrategy represents how to sort statistics
type SortStrategy int

// Provided strategys
const (
	Default SortStrategy = iota
)

// StSlice helps sort statistics
type StSlice struct {
	Details      []*Detail
	SortStrategy SortStrategy
}

// Len of statistics
func (st StSlice) Len() int {
	return len(st.Details)
}

// Less compares statistics
func (st StSlice) Less(i, j int) bool {
	if st.Details[i].Dead {
		return false
	}
	eri := st.Details[i].LastErrRate()
	erj := st.Details[j].LastErrRate()
	if math.Abs(eri-erj) < 0.00001 {
		return st.Details[i].LastAverageCost() < st.Details[j].LastAverageCost()
	}
	return eri < erj
}

// Swap btw two statistics
func (st StSlice) Swap(i, j int) {
	st.Details[i], st.Details[j] = st.Details[j], st.Details[i]
}
