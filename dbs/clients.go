package dbs

import (
	"dnscheck/variation"
	"sync"
	"time"

	"gorm.io/gorm"
)

type ClientConfig struct {
	Domains            []string       `json:"domains"`
	DomainBlocks       []*DomainBlock `json:"domainBlocks"`
	PureLogMode        interface{}    `json:"logMode"`
	DNS                []string       `json:"dns" yaml:"dnses"`
	ClientId           string         `json:"clientId"`
	BatchSize          int            `json:"batchSize"`
	processing         bool
	CheckMainDomains   bool `yaml:"check_main_domains"`
	Concurrency        int  `yaml:"concurrency"`
	TrackResponseTimes bool `yaml:"track_response_times"`
}

type DomainBlock struct {
	Concurrency     int
	CheckMainDomain bool
	Domain          string
	Variations      []variation.VariationRecord
	Dns             []string
	Id              string
}

type DomainInfo struct {
	IsValid    bool
	IsMain     bool
	MainDomain string
	Domain     variation.VariationRecord
	IP         []string
	Host       []string
}
type ResultInfo struct {
	gorm.Model
	Domain        string
	All           int
	DomainBlockId string
	Errors        []string     `gorm:"-"`
	ErrorsJSON    string       // Used to avoid foreign tables etc
	InfoJSON      string       // Used to avoid foreign tables etc
	Infos         []DomainInfo `gorm:"-"`
	ClientId      string
}

type ClientStore struct {
	sync.Mutex
	SpawnClientConfig func(string) *ClientConfig
	GetBatchForClient func(string) []*DomainBlock
	data              map[string]*ClientConfig
	activeTasks       []ClientTask
	openDomainBlocks  []*DomainBlock
	domainChan        chan []*DomainBlock
	resultChan        chan *ResultInfo
	rrIndex           int
}

type ClientTask struct {
	ClientId     string
	DomainBlocks []*DomainBlock
	Expires      time.Time
}

var ClientDB ClientStore

func init() {
	ClientDB = ClientStore{
		data:        make(map[string]*ClientConfig, 0),
		activeTasks: make([]ClientTask, 0),
		domainChan:  make(chan []*DomainBlock, 1),
		resultChan:  make(chan *ResultInfo, 1),
	}
}

func (s *ClientStore) Get(id string) *ClientConfig {
	//dbConfig, ok := s.data[id]

	// Always get the newest one...
	//if !ok {
	s.Lock()
	dbConfig := s.SpawnClientConfig(id)
	s.data[dbConfig.ClientId] = dbConfig
	s.Unlock()
	//}
	// Check if client can do a new task
	if !dbConfig.processing {

		s.Lock()
		dBlocks := s.GetBatchForClient(id)

		//select {
		//case block := <-s.domainChan:
		dbConfig.DomainBlocks = dBlocks
		dbConfig.processing = true
		// logrus.Info("Added block to client id ", dbConfig.ClientId)

		task := ClientTask{
			ClientId:     dbConfig.ClientId,
			DomainBlocks: dbConfig.DomainBlocks,
			Expires:      time.Now().Add(60 * time.Minute),
		}
		s.activeTasks = append(s.activeTasks, task)
		s.Unlock()
		//default:
		//}
	}

	return dbConfig
}

// TODO: Check expiration

func (s *ClientStore) AddResult(res ResultInfo) {
	// logrus.Info(s.data)
	//dbConfig, ok := s.data[res.ClientId]
	//if !ok {
	s.Lock()
	// TODO error handling
	// logger.Error("Failed to set results for non existing client")
	dbConfig := s.SpawnClientConfig(res.ClientId)
	s.data[dbConfig.ClientId] = dbConfig
	s.Unlock()
	//}

	dbConfig.processing = false
	dbConfig.DomainBlocks = nil

	// TOOD Mutex for activeTasks
	// Remove task
	s.Lock()
	for i, v := range s.activeTasks {
		if v.ClientId == dbConfig.ClientId {
			s.activeTasks = append(s.activeTasks[:i], s.activeTasks[i+1:]...)
			break
		}
	}
	s.Unlock()

	s.resultChan <- &res
}

func (s *ClientStore) WaitForClientCompletion() *ResultInfo {
	return <-s.resultChan
}

// Round Robin
func (s *ClientStore) AddDomainBlock(block []*DomainBlock) {
	/*i := 0
	for key := range s.data {
		if i == s.rrIndex {
			s.data[key].DomainBlocks = append(s.data[key].DomainBlocks, *block)
			break
		}
	}
	s.rrIndex++
	if s.rrIndex >= len(s.data) {
		s.rrIndex = 0
	}*/
	// Waits until a client is free to pick this one
	s.domainChan <- block
}
