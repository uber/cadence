// Copyright (c) 2017 Uber Technologies, Inc.
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

package sqlshared

import (
	"database/sql"
	"errors"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

const (
	shardID = 54321
)

var errMissingArgs = errors.New("missing one or more args for API")

// InsertIntoDomain inserts a single row into domains table
func (mdb *DB) InsertIntoDomain(row *sqldb.DomainRow) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.CreateDomainQuery(), row.ID, row.Name, row.IsGlobal, row.Data, row.DataEncoding)
}

// UpdateDomain updates a single row in domains table
func (mdb *DB) UpdateDomain(row *sqldb.DomainRow) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.UpdateDomainQuery(), row.Name, row.Data, row.DataEncoding, row.ID)
}

// SelectFromDomain reads one or more rows from domains table
func (mdb *DB) SelectFromDomain(filter *sqldb.DomainFilter) ([]sqldb.DomainRow, error) {
	switch {
	case filter.ID != nil || filter.Name != nil:
		return mdb.selectFromDomain(filter)
	case filter.PageSize != nil && *filter.PageSize > 0:
		return mdb.selectAllFromDomain(filter)
	default:
		return nil, errMissingArgs
	}
}

func (mdb *DB) selectFromDomain(filter *sqldb.DomainFilter) ([]sqldb.DomainRow, error) {
	var err error
	var row sqldb.DomainRow
	switch {
	case filter.ID != nil:
		err = mdb.conn.Get(&row, mdb.driver.GetDomainByIDQuery(), shardID, *filter.ID)
	case filter.Name != nil:
		err = mdb.conn.Get(&row, mdb.driver.GetDomainByNameQuery(), shardID, *filter.Name)
	}
	if err != nil {
		return nil, err
	}
	return []sqldb.DomainRow{row}, err
}

func (mdb *DB) selectAllFromDomain(filter *sqldb.DomainFilter) ([]sqldb.DomainRow, error) {
	var err error
	var rows []sqldb.DomainRow
	switch {
	case filter.GreaterThanID != nil:
		err = mdb.conn.Select(&rows, mdb.driver.ListDomainsRangeQuery(), shardID, *filter.GreaterThanID, *filter.PageSize)
	default:
		err = mdb.conn.Select(&rows, mdb.driver.ListDomainsQuery(), shardID, filter.PageSize)
	}
	return rows, err
}

// DeleteFromDomain deletes a single row in domains table
func (mdb *DB) DeleteFromDomain(filter *sqldb.DomainFilter) (sql.Result, error) {
	var err error
	var result sql.Result
	switch {
	case filter.ID != nil:
		result, err = mdb.conn.Exec(mdb.driver.DeleteDomainByIDQuery(), shardID, filter.ID)
	default:
		result, err = mdb.conn.Exec(mdb.driver.DeleteDomainByNameQuery(), shardID, filter.Name)
	}
	return result, err
}

// LockDomainMetadata acquires a write lock on a single row in domain_metadata table
func (mdb *DB) LockDomainMetadata() error {
	var row sqldb.DomainMetadataRow
	err := mdb.conn.Get(&row.NotificationVersion, mdb.driver.LockDomainMetadataQuery())
	return err
}

// SelectFromDomainMetadata reads a single row in domain_metadata table
func (mdb *DB) SelectFromDomainMetadata() (*sqldb.DomainMetadataRow, error) {
	var row sqldb.DomainMetadataRow
	err := mdb.conn.Get(&row.NotificationVersion, mdb.driver.GetDomainMetadataQuery())
	return &row, err
}

// UpdateDomainMetadata updates a single row in domain_metadata table
func (mdb *DB) UpdateDomainMetadata(row *sqldb.DomainMetadataRow) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.UpdateDomainMetadataQuery(), row.NotificationVersion+1, row.NotificationVersion)
}
