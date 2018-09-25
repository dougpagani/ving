package statistic

import (
	"fmt"
	"net"
	"time"

	"github.com/yittg/ving/types"
)

const (
	errChar = "-"
)

// TraceSt for trace
type TraceSt struct {
	ID            int
	PreSuccessful bool
	Cost          []float64
	From          []string
}

func transformFrom(from net.Addr) string {
	if udpFrom, ok := from.(*net.UDPAddr); ok {
		return udpFrom.IP.String()
	}
	return from.String()
}

// DealRecord deal new record at t
func (st *TraceSt) DealRecord(record types.Record) {
	var from string
	if record.Successful {
		from = fmt.Sprintf("%2d:%s", record.TTL, transformFrom(record.From))
	} else {
		from = "   " + errChar
	}
	cost := float64(record.Cost) / float64(time.Millisecond)
	if record.TTL == 1 {
		st.From = []string{from}
		if record.Successful {
			st.Cost = []float64{cost}
		}
	} else {
		if !record.Successful && !st.PreSuccessful {
			st.From[len(st.From)-1] += errChar
		} else {
			st.From = append(st.From, from)
		}
		if record.Successful {
			st.Cost = append(st.Cost, cost)
		}
	}
	st.PreSuccessful = record.Successful
}
