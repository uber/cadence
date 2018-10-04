package persistence

import (
	"sync"

	"fmt"
	"strings"

	"strconv"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/service/config"
)

type (
	Factory interface {
		NewTaskManager() (TaskManager, error)
		NewShardManager() (ShardManager, error)
		NewHistoryManager() (HistoryManager, error)
		NewMetadataManager() (MetadataManager, error)
		NewExecutionManagerFactory() (ExecutionManagerFactory, error)
		NewVisibilityManager() (VisibilityManager, error)
	}
	factoryImpl struct {
		sync.RWMutex
		config              *config.Persistence
		taskMgr             TaskManager
		shardMgr            ShardManager
		historyMgr          HistoryManager
		metadataMgr         MetadataManager
		executionMgrFactory ExecutionManagerFactory
		logger              *bark.Logger
	}
	cassandraFactory struct {
	}
)

func NewFactory(cfg *config.Persistence, logger *bark.Logger) Factory {
	return &factoryImpl{config: cfg, logger: logger}
}

func (f *factoryImpl) NewTaskManager() (TaskManager, error) {
}

func (f *factoryImpl) NewShardManager() (ShardManager, error) {

}

func (f *factoryImpl) NewHistoryManager() (HistoryManager, error) {

}

func (f *factoryImpl) NewMetadataManager() (MetadataManager, error) {

}

func (f *factoryImpl) NewExecutionManagerFactory() (ExecutionManagerFactory, error) {

}

func (f *factoryImpl) NewVisibilityManager() (VisibilityManager, error) {

}

func (f *factoryImpl) getTaskManager() TaskManager {
	f.RLock()
	defer f.RUnlock()
	return f.taskMgr
}

func (f *factoryImpl) newTaskManager() (TaskManager, error) {
	f.Lock()
	defer f.Unlock()
	if f.taskMgr != nil {
		return f.taskMgr, nil
	}
	storeInfo := f.config.DataStores[f.config.DefaultStore]
	var err error
	var result TaskManager
	if storeInfo.SQL != nil {
		cfg := storeInfo.SQL
		host, port, err1 := splitHostPort(cfg.ConnectAddr)
		if err1 != nil {
			return nil, err
		}
		result, err = sql.NewTaskPersistence(host, port, cfg.User, cfg.Password, cfg.DatabaseName, f.logger)
	}
	if storeInfo.Cassandra != nil {
		result, err = cassandra.NewTaskPersistence(storeInfo.Cassandra, f.logger)
	}
	if err != nil {
		return nil, err
	}
	f.taskMgr = result
	return result, nil
}

func splitHostPort(addr string) (string, int, error) {
	hostPort := strings.Split(strings.TrimSpace(addr), ":")
	if len(hostPort) != 2 {
		return "", 0, fmt.Errorf("invalid address: %v", addr)
	}
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return "", 0, fmt.Errorf("error parsing host:port from %v: %v", addr, err)
	}
	return hostPort[0], port, nil
}
