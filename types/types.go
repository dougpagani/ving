package types

// ItemHeader describe id name and order
type ItemHeader struct {
	ID     int
	Target string
	Iter   int
}

// SpItem for spark line
type SpItem struct {
	ItemHeader
	Value   int
	Display interface{}
}

// ErrItem for errors
type ErrItem struct {
	ItemHeader
	Err string
}

// DataSet to display
type DataSet struct {
	SpItems  []SpItem
	ErrItems []ErrItem
}
