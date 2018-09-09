package statistic

import (
	"math"

	"github.com/yittg/ving/types"
)

const (
	errStatisticWindow = 1000
)

// Detail provide ability for statistic
type Detail struct {
	ID    int
	Title string

	Total             int
	ErrCount          int
	Cost              []int
	Dead              bool
	lastErrRecord     *ErrorRecordWithIter
	lastErrIter       uint64
	lastNIterRecord   []RecordWithIter
	lastNIterErrCount int
	lastNIterCost     int64
}

// DealRecord deal new record at iter
func (s *Detail) DealRecord(iter uint64, record types.Record) {
	s.lastNIterRecord = append(s.lastNIterRecord, RecordWithIter{
		Iter:   iter,
		Record: record,
	})
	s.Total = record.Rounds

	if record.Successful {
		s.lastNIterCost += int64(record.Cost)
		s.Cost = append(s.Cost[1:], int(record.Cost))
	} else {
		s.ErrCount++
		s.lastNIterErrCount++
		s.lastErrRecord = &ErrorRecordWithIter{
			Iter: iter,
			Err:  record.ErrMsg,
		}
		s.Cost = append(s.Cost[1:], 0)
		if record.IsFatal {
			s.Dead = true
		}
	}
}

// RetireRecord retires those records out of window
func (s *Detail) RetireRecord(iter uint64) {
	for i := 0; i < len(s.lastNIterRecord); i++ {
		record := s.lastNIterRecord[i]
		if record.Iter+errStatisticWindow < iter {
			if !record.Record.Successful {
				s.lastNIterErrCount--
			} else {
				s.lastNIterCost -= int64(record.Record.Cost)
			}
			continue
		}
		s.lastNIterRecord = s.lastNIterRecord[i:]
		break
	}
}

// LastRecord represents latest record
func (s *Detail) LastRecord() *RecordWithIter {
	n := len(s.lastNIterRecord)
	if n == 0 {
		return nil
	}
	return &s.lastNIterRecord[n-1]
}

// LastErrorRecord represents latest error record
func (s *Detail) LastErrorRecord() *ErrorRecordWithIter {
	return s.lastErrRecord
}

// LastErrRate represents latest error rate in window
func (s *Detail) LastErrRate() float64 {
	return float64(s.lastNIterErrCount) / float64(len(s.lastNIterRecord))
}

// LastAverageCost represents latest speed in window
func (s *Detail) LastAverageCost() int64 {
	successfulCount := len(s.lastNIterRecord) - s.lastNIterErrCount
	if successfulCount <= 0 {
		return math.MaxInt64
	}
	return s.lastNIterCost / int64(successfulCount)
}

// ResizeViewWindow resize view window to size
func (s *Detail) ResizeViewWindow(size int) {
	crtSize := len(s.Cost)
	if crtSize == size {
		return
	}
	if crtSize == 0 {
		s.Cost = make([]int, size)
	} else if crtSize < size {
		s.Cost = append(make([]int, size-crtSize), s.Cost...)
	} else {
		s.Cost = s.Cost[crtSize-size:]
	}
}

// RecordWithIter represents Record generated of the iteration
type RecordWithIter struct {
	Iter   uint64
	Record types.Record
}

// View of record
func (r *RecordWithIter) View() interface{} {
	if !r.Record.Successful {
		return "Err"
	}
	return r.Record.Cost
}

// ErrorRecordWithIter error record
type ErrorRecordWithIter struct {
	Iter uint64
	Err  string
}
