package gocql

import (
	"context"
	"time"

	"github.com/uber/cadence/common/service/config"
)

type (
	Client interface {
		CreateSession(ClusterConfig) (Session, error)

		// TODO: we should be able to remove this method
		ParseConsistency(string) Consistency

		IsTimeoutError(error) bool
		IsNotFoundError(error) bool
		IsThrottlingError(error) bool
	}

	Session interface {
		Query(string, ...interface{}) Query
		NewBatch(BatchType) Batch
		MapExecuteBatchCAS(Batch, map[string]interface{}) (bool, Iter, error)
		Close()
	}

	// no implemtation needed if we don't need Consistency
	Query interface {
		Exec() error
		Scan(...interface{}) error
		MapScan(map[string]interface{}) error
		MapScanCAS(map[string]interface{}) (bool, error)
		Iter() Iter
		PageSize(int) Query
		PageState([]byte) Query
		WithContext(context.Context) Query
		// TODO: why do we set consistency level to gocql.One for read queries in visibility store?
		Consistency(Consistency) Query
	}

	// no implemtation needed
	Batch interface {
		Query(string, ...interface{})
		WithContext(context.Context) Batch
	}

	// no implemtation needed
	Iter interface {
		Scan(...interface{}) bool
		MapScan(map[string]interface{}) bool
		PageState() []byte
		Close() error
	}

	// no implemtation needed
	// but note that we do value.([]gocql.UUID) in our code,
	// this need to be changed to use reflection and type assert each element
	UUID interface {
		String() string
	}

	BatchType byte

	Consistency uint16

	SerialConsistency uint16

	ClusterConfig struct {
		config.Cassandra

		ProtoVersion      int
		Consistency       Consistency
		SerialConsistency SerialConsistency
		Timeout           time.Duration
	}
)

// Note: don't do directly type cast for conversion
// need mapper function to map these values to the constant defined by the underlying library

const (
	LoggedBatch   BatchType = 0
	UnloggedBatch BatchType = 1
	CounterBatch  BatchType = 2
)

const (
	Any         Consistency = 0x00
	One         Consistency = 0x01
	Two         Consistency = 0x02
	Three       Consistency = 0x03
	Quorum      Consistency = 0x04
	All         Consistency = 0x05
	LocalQuorum Consistency = 0x06
	EachQuorum  Consistency = 0x07
	LocalOne    Consistency = 0x0A
)

const (
	Serial      SerialConsistency = 0x08
	LocalSerial SerialConsistency = 0x09
)
