// Copyright (c) 2020 Uber Technologies, Inc.
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

package db2

import (
	"database/sql"
	"fmt"

	db2 "github.com/ibmdb/go_ibm_db"
	"github.com/jmoiron/sqlx"

	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

// db represents a logical connection to db2 database
type db struct {
	dbStr     string
	db        *sqlx.DB
	tx        *sqlx.Tx
	conn      sqlplugin.Conn
	converter DataConverter
}

var _ sqlplugin.AdminDB = (*db)(nil)
var _ sqlplugin.DB = (*db)(nil)
var _ sqlplugin.Tx = (*db)(nil)

// ErrDupEntry DB2 Error 23505 indicates a duplicate primary key i.e. the row already exists,
// so we don't do the insert and return a ConditionalUpdate error.
const ErrDupEntry = 23505

func (mdb *db) IsDupEntryError(err error) bool {
	sqlErr, ok := err.(*db2.Error)
	return ok && sqlErr.Diag[0].NativeError == ErrDupEntry
}

// newDB returns an instance of DB, which is a logical
// connection to the underlying database
func newDB(xdb *sqlx.DB, tx *sqlx.Tx, dbStr string) *db {
	mdb := &db{db: xdb, tx: tx, dbStr: dbStr}
	mdb.conn = xdb
	if tx != nil {
		mdb.conn = tx
	}
	mdb.converter = &converter{}
	return mdb
}

func (mdb *db) getContextStmt(query string) (*sqlx.Stmt, error) {
	if mdb.tx != nil {
		return mdb.tx.Preparex(query)
	}
	return mdb.db.Preparex(query)
}

// Get using this DB
func (mdb *db) Get(dest interface{}, query string, args ...interface{}) error {
	stmt, _ := mdb.getContextStmt(query)
	err := stmt.Get(dest, args...)
	// pay price for non-cached statements
	// as db2 driver does not support bindvars in a regular way
	stmt.Close()
	if err != nil {
		fmt.Printf("Get >>>>>>>>>:  %v, %v, [dest=%v] [err=%v]\n", mdb.tx, query, args, dest, err)
	}
	return err
}

func (mdb *db) Select(dest interface{}, query string, args ...interface{}) error {
	stmt, _ := mdb.getContextStmt(query)
	err := stmt.Select(dest, args...)
	stmt.Close()
	if err != nil {
		fmt.Printf("[Transaction=%v]Selet >>>>>>>>>:  %v, %v, [dest=%v] [err=%v]\n", mdb.tx, query, args, dest, err)
	}
	return err
}

func (mdb *db) NamedExec(query string, arg interface{}) (sql.Result, error) {
	r, err := mdb.conn.NamedExec(query, arg)
	if err != nil {
		fmt.Printf("[Transaction=%v]NamedExec >>>>>>>>>:  %v, %v, [err=%v]\n", mdb.tx, query, arg, err)
	}
	return r, err
}

func (mdb *db) Exec2(query string, args ...interface{}) (sql.Result, error) {
	stmt, _ := mdb.getContextStmt(query)
	r, err := stmt.Exec(args...)
	stmt.Close()
	if err != nil {
		fmt.Printf("[Transaction=%v]Exec2 >>>>>>>>>:  %v, %v, [result=%v] [err=%v]\n", mdb.tx, query, args, r, err)
	}
	return r, err
}

// BeginTx starts a new transaction and returns a reference to the Tx object
func (mdb *db) BeginTx() (sqlplugin.Tx, error) {
	xtx, err := mdb.db.Beginx()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("!!!!!! Transaction START: [%v] ", xtx)
	return newDB(mdb.db, xtx, mdb.dbStr), nil
}

// Commit commits a previously started transaction
func (mdb *db) Commit() error {
	//fmt.Printf("!!!!!! COMMIT: [%v] ", mdb.tx)
	return mdb.tx.Commit()
}

// Rollback triggers rollback of a previously started transaction
func (mdb *db) Rollback() error {
	//fmt.Printf("!!!!!! ROLLBACK: [%v] ", mdb.tx)
	return mdb.tx.Rollback()
}

// Close closes the connection to the mysql db
func (mdb *db) Close() error {
	return mdb.db.Close()
}

// PluginName returns the name of the mysql plugin
func (mdb *db) PluginName() string {
	return PluginName
}
