package port

import (
	"github.com/yittg/ving/addons/port/types"
)

var wellKnownPorts = []types.PortDesc{
	{Name: "ssh", Port: 22},
	{Name: "http", Port: 80},
	{Name: "https", Port: 443},
	{Name: "docker", Port: 2375},
	{Name: "etcd", Port: 2379},
	{Name: "mysql", Port: 3306},
	{Name: "PostgreSQL", Port: 5432},
	{Name: "AMQP", Port: 5671},
	{Name: "redis", Port: 6379},
	{Name: "zabbix", Port: 10050},
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
