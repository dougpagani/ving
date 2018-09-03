package types

// WithId describe id name and order
type WithId struct {
	Id    string
	Order int
}

// SpItem for spark line
type SpItem struct {
	WithId
	Value   int
	Display interface{}
}

// ErrItem for errors
type ErrItem struct {
	WithId
	Err string
}

// DataSet to display
type DataSet struct {
	SpItems  []SpItem
	ErrItems []ErrItem
}
