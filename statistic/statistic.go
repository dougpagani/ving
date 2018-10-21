package statistic

import (
	"math"
	"strings"
	"time"

	"github.com/yittg/ving/config"
	"github.com/yittg/ving/types"
)

var (
	statisticConfig    = config.GetConfig().Statistic
	errStatisticWindow = statisticConfig.Window.Value
)

// Detail provide ability for statistic
type Detail struct {
	ID    int
	Title string

	Total             int
	ErrCount          int
	Cost              []int
	Dead              bool
	lastErrRecord     *ErrorRecordAt
	lastNIterRecord   []RecordAt
	lastNIterErrCount int
	lastNIterCost     int64
}

// DealRecord deal new record at t
func (s *Detail) DealRecord(t time.Time, record types.Record) {
	s.lastNIterRecord = append(s.lastNIterRecord, RecordAt{
		T:      t,
		Record: record,
	})
	s.Total = record.Rounds

	if record.Successful {
		s.lastNIterCost += int64(record.Cost)
		s.Cost = append(s.Cost[1:], int(record.Cost))
	} else {
		s.ErrCount++
		s.lastNIterErrCount++
		s.lastErrRecord = &ErrorRecordAt{
			T:   t,
			Err: record.ErrMsg,
		}
		s.Cost = append(s.Cost[1:], 0)
		if record.IsFatal {
			s.Dead = true
		}
	}
}

// RetireRecord retires those records out of window
func (s *Detail) RetireRecord(t time.Time) {
	for i := 0; i < len(s.lastNIterRecord); i++ {
		record := s.lastNIterRecord[i]
		if record.T.Add(errStatisticWindow).Before(t) {
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
func (s *Detail) LastRecord() *RecordAt {
	n := len(s.lastNIterRecord)
	if n == 0 {
		return nil
	}
	return &s.lastNIterRecord[n-1]
}

// LastErrorRecord represents latest error record
func (s *Detail) LastErrorRecord() *ErrorRecordAt {
	return s.lastErrRecord
}

// LastErrRate represents latest error rate in window
func (s *Detail) LastErrRate() float64 {
	return float64(s.lastNIterErrCount) / float64(len(s.lastNIterRecord))
}

// LastErrRateLevel represents the level of last error rate
func (s *Detail) LastErrRateLevel() int {
	r := s.LastErrRate()
	for level, thresh := range statisticConfig.ErrorRateThresh {
		if r < thresh {
			return level
		}
	}
	return len(statisticConfig.ErrorRateThresh)
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

// RecordAt represents Record generated of the iteration
type RecordAt struct {
	T      time.Time
	Record types.Record
}

// View of record
func (r *RecordAt) View() interface{} {
	if !r.Record.Successful {
		errMsg := strings.Split(r.Record.ErrMsg, ":")
		return errMsg[len(errMsg)-1]
	}
	v := r.Record.Cost
	if v > time.Second {
		v = v.Truncate(10 * time.Millisecond)
	} else if v > time.Millisecond {
		v = v.Truncate(10 * time.Microsecond)
	} else {
		v = v.Truncate(10 * time.Nanosecond)
	}
	return v
}

// ErrorRecordAt error record
type ErrorRecordAt struct {
	T   time.Time
	Err string
}
