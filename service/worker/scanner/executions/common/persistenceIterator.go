package common

import (
	"github.com/uber/cadence/common/pagination"
)

type (
	persistenceIterator struct {
		pr PersistenceRetryer
	}
)

func NewPersistenceIterator(pr PersistenceRetryer) ExecutionIterator {
	paginatedIterator := pagination.NewIterator()
	return &persistenceIterator{
		pr: pr,
	}
}

func (i *persistenceIterator) Next() (*Execution, error) {

}
