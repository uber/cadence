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

package sqlshared

import (
    "github.com/uber/cadence/common/persistence/sql/storage/sqldb"
    "github.com/uber/cadence/common/service/config"
)

type (
    // Driver is the driver interface that each SQL database needs to implement
    Driver interface {
        GetDriverName() string
        CreateDBConnection(cfg *config.SQL) (sqldb.Interface, error)

        //domain
        CreateDomainQry() string
        UpdateDomainQry() string
        GetDomainByIDQry() string
        GetDomainByNameQry() string
        ListDomainsQry() string
        ListDomainsRangeQry() string
        DeleteDomainByIDQry() string
        DeleteDomainByNameQry() string
        GetDomainMetadataQry() string
        LockDomainMetadataQry() string
        UpdateDomainMetadataQry() string

        //events
        AddHistoryNodesQry() string
        GetHistoryNodesQry() string
        DeleteHistoryNodesQry() string
        AddHistoryTreeQry() string
        GetHistoryTreeQry() string
        DeleteHistoryTreeQry() string

        //execution
        //execution_map
        //queue
        //shard
        //task
        //visibility
    }
)