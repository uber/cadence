// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package postgres

import (
	"fmt"
	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
	"strconv"

	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/uber/cadence/common/persistence/sql/storage"
	"github.com/uber/cadence/common/service/config"
)

const (
	DriverName             = "postgres"
	dataSourceNamePostgres = "user=%v password=%v host=%v port=%v dbname=%v sslmode=disable "
)

type driver struct{}

var _ sqldb.Driver = (*driver)(nil)

func init() {
	storage.RegisterDriver(DriverName, &driver{})
}

// InitDB initialize the db object
func (d *driver) InitDB(cfg *config.SQL) (sqldb.DB, error){
	conn, err := d.createDBConnection(cfg)
	if err != nil {
		return nil, err
	}
	db := NewDB(conn, nil)
	return db, nil
}

// CreateDBConnection creates a returns a reference to a logical connection to the
// underlying SQL database. The returned object is to tied to a single
// SQL database and the object can be used to perform CRUD operations on
// the tables in the database
func (d *driver) createDBConnection(cfg *config.SQL) (*sqlx.DB, error) {
	ss := strings.Split(cfg.ConnectAddr, ":")
	if len(ss) != 2 {
		return nil, fmt.Errorf("invalid connect address, it must be in host:port format, %v", cfg.ConnectAddr)
	}
	host := ss[0]
	port, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v, %v", ss[1], cfg.ConnectAddr)
	}

	dbName := cfg.DatabaseName
	//NOTE: postgres doesn't allow to connect with empty dbName, the admin dbName is "postgres"
	if dbName == ""{
		dbName = "postgres"
	}
	db, err := sqlx.Connect(DriverName, fmt.Sprintf(dataSourceNamePostgres, cfg.User, cfg.Password, host, port, dbName ))

	if err != nil {
		return nil, err
	}

	// Maps struct names in CamelCase to snake without need for db struct tags.
	db.MapperFunc(strcase.ToSnake)
	return db, nil
}


