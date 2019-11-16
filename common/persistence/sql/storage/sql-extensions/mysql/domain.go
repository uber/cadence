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

package mysql

const (
    createDomainQry = `INSERT INTO 
 domains (id, name, is_global, data, data_encoding)
 VALUES(?, ?, ?, ?, ?)`

    updateDomainQry = `UPDATE domains 
 SET name = ?, data = ?, data_encoding = ?
 WHERE shard_id=54321 AND id = ?`

    getDomainPart = `SELECT id, name, is_global, data, data_encoding FROM domains`

    getDomainByIDQry   = getDomainPart + ` WHERE shard_id=? AND id = ?`
    getDomainByNameQry = getDomainPart + ` WHERE shard_id=? AND name = ?`

    listDomainsQry      = getDomainPart + ` WHERE shard_id=? ORDER BY id LIMIT ?`
    listDomainsRangeQry = getDomainPart + ` WHERE shard_id=? AND id > ? ORDER BY id LIMIT ?`

    deleteDomainByIDQry   = `DELETE FROM domains WHERE shard_id=? AND id = ?`
    deleteDomainByNameQry = `DELETE FROM domains WHERE shard_id=? AND name = ?`

    getDomainMetadataQry    = `SELECT notification_version FROM domain_metadata`
    lockDomainMetadataQry   = `SELECT notification_version FROM domain_metadata FOR UPDATE`
    updateDomainMetadataQry = `UPDATE domain_metadata SET notification_version = ? WHERE notification_version = ?`
)

func (d *driver) CreateDomainQry() string {
    return createDomainQry
}

func (d *driver) UpdateDomainQry() string {
    return updateDomainQry
}

func (d *driver) GetDomainByIDQry() string {
    return getDomainByIDQry
}

func (d *driver) GetDomainByNameQry() string {
    return getDomainByNameQry
}

func (d *driver) ListDomainsQry() string {
    return listDomainsQry
}

func (d *driver) ListDomainsRangeQry() string {
    return listDomainsRangeQry
}

func (d *driver) DeleteDomainByIDQry() string {
    return deleteDomainByIDQry
}

func (d *driver) DeleteDomainByNameQry() string {
    return deleteDomainByNameQry
}

func (d *driver) GetDomainMetadataQry() string {
    return getDomainMetadataQry
}

func (d *driver) LockDomainMetadataQry() string {
    return lockDomainMetadataQry
}

func (d *driver) UpdateDomainMetadataQry() string {
    return updateDomainMetadataQry
}

