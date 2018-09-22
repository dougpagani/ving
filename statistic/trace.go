package statistic

import (
	"fmt"
	"net"
	"time"

	"github.com/yittg/ving/types"
)

// TraceSt for trace
type TraceSt struct {
	ID   int
	Cost []float64
	From []string
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
		from = transformFrom(record.From)
	} else {
		from = record.ErrMsg
	}
	from = fmt.Sprintf("%2d:%s", record.TTL, from)
	cost := float64(record.Cost) / float64(time.Millisecond)
	if record.TTL == 1 {
		st.From = []string{from}
		if record.Successful {
			st.Cost = []float64{cost}
		}
	} else {
		st.From = append(st.From, from)
		if record.Successful {
			st.Cost = append(st.Cost, cost)
		}
	}
}
