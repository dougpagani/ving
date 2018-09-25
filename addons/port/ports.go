package port

import (
	"github.com/yittg/ving/addons/port/types"
)

var wellKnownPorts = []types.PortDesc{
	{"ssh", 22},
	{"http", 80},
	{"https", 443},
	{"DNSs", 853},
	{"docker", 2375},
	{"etcd", 2379},
	{"mysql", 3306},
	{"PostgreSQL", 5432},
	{"AMQP", 5671},
	{"redis", 6379},
	{"zabbix", 10050},
}

type sortable []types.PortDesc

// Len of sortable ports describe
func (s sortable) Len() int {
	return len(s)
}

// Less compare btw ports describe
func (s sortable) Less(i, j int) bool {
	return s[i].Port < s[j].Port
}

// Swap btw ports describe
func (s sortable) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

var predefinedPorts sortable

func getPredefinedPorts() []types.PortDesc {
	return predefinedPorts
}
