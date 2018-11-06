package port

import (
	"sort"
	"strconv"

	"github.com/yittg/ving/addons/port/types"
	"github.com/yittg/ving/config"
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
var predefinedPortsMap map[int]types.PortDesc

func buildPredefinedPorts() {
	predefinedPorts = append(wellKnownPorts, config.GetConfig().AddOns.Ports.Extra...)
	sort.Sort(predefinedPorts)

	predefinedPortsMap = make(map[int]types.PortDesc, len(predefinedPorts))
	for _, pd := range predefinedPorts {
		predefinedPortsMap[pd.Port] = pd
	}
}

func getPredefinedPorts() []types.PortDesc {
	return predefinedPorts
}

func getPredefinedPortByN(port int) *types.PortDesc {
	pd, ok := predefinedPortsMap[port]
	if !ok {
		return nil
	}
	return &pd
}

func getNameOfPort(port int) string {
	pd := getPredefinedPortByN(port)
	if pd == nil {
		return strconv.Itoa(port)
	}
	return pd.Name
}
