package sqldriver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewDriver returns a driver to SQL, either using singleton Driver or sharded Driver 
func NewDriver(xdbs []*sqlx.DB, tx *sqlx.Tx, dbShardID int) (Driver, error) {

	if len(xdbs) == 1 {
		return newSingletonSQLDriver(xdbs[0], tx, dbShardID), nil
	}

	if len(xdbs) <= 1 {
		return nil, fmt.Errorf("invalid number of connection for sharded SQL driver")
	}
	// this is the case of multiple database with sharding
	return  newShardedSQLDriver(xdbs, tx, dbShardID), nil
}
