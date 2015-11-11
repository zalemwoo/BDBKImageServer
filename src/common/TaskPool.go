package common

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"gopkg.in/fatih/pool.v2"
)

type ServerInfo struct {
	Host string
	Port string
	Url  string
}

type TaskPoolConfig struct {
	ConnectionsPerServer int
	Servers              []ServerInfo
}

type TaskPool struct {
	pools          []pool.Pool
	npools         int
	connPerPool    int
	Queue          chan func(pool.Pool) error
	Result         chan error
	Total          int
	finishMutex    sync.Mutex
	finish         bool
	finishCallback func()
}

var wg sync.WaitGroup

func (self *TaskPool) Init(config *TaskPoolConfig) {
	self.connPerPool = config.ConnectionsPerServer
	self.npools = len(config.Servers)
	self.Total = self.connPerPool * self.npools
	self.pools = make([]pool.Pool, 0, self.npools)
	for _, v := range config.Servers {
		url := v.Url
		pool, err := pool.NewChannelPool(1, self.connPerPool,
			func() (net.Conn, error) {
				return net.DialTimeout("tcp", url, 20*time.Second)
			})

		if err != nil {
			log.Fatal("create connection pool error. err: %v", err)
			return
		}
		log.Printf("Connection pool created. url: %s", v.Url)
		self.pools = append(self.pools, pool)
	}

	self.finish = false
	self.Queue = make(chan func(pool.Pool) error, self.Total)
	self.Result = make(chan error, self.Total)
}

func (self *TaskPool) Start() {
	defer func() {
		close(self.Queue)
		close(self.Result)
		for _, p := range self.pools {
			p.Close()
		}
	}()

	for {
		if self.finish {
			break
		}
		wg.Add(self.Total)
		for i := 0; i < self.Total; i++ {
			self.doTask()
		}
		wg.Wait()
	}

	if self.finishCallback != nil {
		self.finishCallback()
	}
}

func (self *TaskPool) doTask() {
	task, ok := <-self.Queue
	if !ok {
		self.Stop()
		return
	}
	go func() {
		defer wg.Done()
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		p := self.pools[r.Intn(self.npools)]
		err := task(p)
		if err != nil {
		}
		// self.Result <- err
	}()
}

func (self *TaskPool) AddTask(task func(pool.Pool) error) {
	self.Queue <- task
}

func (self *TaskPool) Stop() {
	self.finishMutex.Lock()
	self.finish = true
	self.finishMutex.Unlock()
}

func (self *TaskPool) SetFinishCallback(callback func()) {
	self.finishCallback = callback
}
