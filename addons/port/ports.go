package port

var knownPorts = []port{
	{"echo", 7},
	{"ftp-t", 20},
	{"ftp-c", 21},
	{"ssh", 22},
	{"http", 80},
	{"https", 443},
	{"DNSs", 853},
	{"openVPN", 1194},
	{"docker", 2375},
	{"docker(ssl)", 2376},
	{"etcd", 2379},
	{"etcd(inner)", 2380},
	{"mysql", 3306},
	{"PostgreSQL", 5432},
	{"AMQP", 5671},
	{"redis", 6379},
	{"http(8008)", 8008},
	{"http(8080)", 8080},
	{"zabbix", 10050},
}

type port struct {
	name string
	port int
}
