package xiaomipush

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var hostSwitch HostSwitch
var once sync.Once

type Server struct {
	host        string
	priority    int
	minPriority int
	maxPriority int
	decrStep    int
	incrStep    int
	mux         sync.Mutex
}

type HostSwitch struct {
	defaultServer   Server
	servers         []Server
	lastRefreshTime time.Time
}

func (this Server) incrPriority() {
	this.changePriority(true, this.incrStep)
}

func (this Server) decrPriority() {
	this.changePriority(false, this.decrStep)
}

func (this Server) changePriority(isIncr bool, step int) {
	this.mux.Lock()
	var priority int
	if isIncr {
		priority = this.priority + step
	} else {
		priority = this.priority - step
	}
	if priority > this.maxPriority {
		priority = this.maxPriority
	}
	if priority < this.minPriority {
		priority = this.minPriority
	}
	this.priority = priority
	this.mux.Unlock()

}

func (this HostSwitch) needRefresh() bool {
	if time.Since(this.lastRefreshTime).Nanoseconds() > RefreshServerHostInterval {
		return true
	}
	return false
}

func (this HostSwitch) init(hostList string) {
	hosts := strings.Split(hostList, ",")
	serverSlice := make([]Server, len(hosts))
	for index, value := range hosts {
		servers := strings.Split(value, ":")
		if len(servers) < 5 {
			continue
		}
		minPriority, _ := strconv.Atoi(servers[1])
		maxPriority, _ := strconv.Atoi(servers[2])
		decrStep, _ := strconv.Atoi(servers[3])
		incrStep, _ := strconv.Atoi(servers[4])
		server := Server{
			host:        servers[0],
			priority:    maxPriority,
			maxPriority: maxPriority,
			minPriority: minPriority,
			incrStep:    incrStep,
			decrStep:    decrStep,
		}
		serverSlice[index] = server
	}
	this.lastRefreshTime = time.Now()
	this.servers = serverSlice
	this.defaultServer = Server{
		host:        ProductionHost,
		priority:    90,
		maxPriority: 90,
		minPriority: 1,
		incrStep:    5,
		decrStep:    10,
	}
}

func (this HostSwitch) selectServer() Server {
	allPriority := 0
	for _, server := range this.servers {
		allPriority = allPriority + server.priority
	}
	point := rand.Intn(allPriority)
	prioritySum := 0
	for _, server := range this.servers {
		prioritySum = prioritySum + server.priority
		if prioritySum > point {
			return server
		}
	}
	return this.defaultServer
}

func getHostSwitch() HostSwitch {
	once.Do(func() {
		hostSwitch = HostSwitch{
			lastRefreshTime: time.Now(),
		}
	})

	return hostSwitch
}
