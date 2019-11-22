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

const (
	createDomainQuery = `INSERT INTO 
 domains (id, name, is_global, data, data_encoding)
 VALUES(?, ?, ?, ?, ?)`

	updateDomainQuery = `UPDATE domains 
 SET name = ?, data = ?, data_encoding = ?
 WHERE shard_id=54321 AND id = ?`

	getDomainPart = `SELECT id, name, is_global, data, data_encoding FROM domains`

	getDomainByIDQuery   = getDomainPart + ` WHERE shard_id=? AND id = ?`
	getDomainByNameQuery = getDomainPart + ` WHERE shard_id=? AND name = ?`

	listDomainsQuery      = getDomainPart + ` WHERE shard_id=? ORDER BY id LIMIT ?`
	listDomainsRangeQuery = getDomainPart + ` WHERE shard_id=? AND id > ? ORDER BY id LIMIT ?`

	deleteDomainByIDQuery   = `DELETE FROM domains WHERE shard_id=? AND id = ?`
	deleteDomainByNameQuery = `DELETE FROM domains WHERE shard_id=? AND name = ?`

	getDomainMetadataQuery    = `SELECT notification_version FROM domain_metadata`
	lockDomainMetadataQuery   = `SELECT notification_version FROM domain_metadata FOR UPDATE`
	updateDomainMetadataQuery = `UPDATE domain_metadata SET notification_version = ? WHERE notification_version = ?`
)

func (d *driver) CreateDomainQuery() string {
	return createDomainQuery
}

func (d *driver) UpdateDomainQuery() string {
	return updateDomainQuery
}

func (d *driver) GetDomainByIDQuery() string {
	return getDomainByIDQuery
}

func (d *driver) GetDomainByNameQuery() string {
	return getDomainByNameQuery
}

func (d *driver) ListDomainsQuery() string {
	return listDomainsQuery
}

func (d *driver) ListDomainsRangeQuery() string {
	return listDomainsRangeQuery
}

func (d *driver) DeleteDomainByIDQuery() string {
	return deleteDomainByIDQuery
}

func (d *driver) DeleteDomainByNameQuery() string {
	return deleteDomainByNameQuery
}

func (d *driver) GetDomainMetadataQuery() string {
	return getDomainMetadataQuery
}

func (d *driver) LockDomainMetadataQuery() string {
	return lockDomainMetadataQuery
}

func (d *driver) UpdateDomainMetadataQuery() string {
	return updateDomainMetadataQuery
}
