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

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/uber/cadence/common/log/loggerimpl"

	"github.com/pkg/errors"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/cassandra"
	"github.com/urfave/cli"
)

type (
	domainMigrationImpl struct {
		metadataV1Mgr persistence.MetadataManager
		metadataV2Mgr *cassandra.CassandraMetadataPersistenceV2
	}
)

// AdminDomainMigration command migrate the domain from V1 table to V2 table
func AdminDomainMigration(c *cli.Context) {
	currentCluster := c.String(FlagCurrentClusterName)
	domainName := c.GlobalString(FlagDomain)

	if currentCluster == "" || domainName == "" {
		ErrorAndExit("missing parameters",
			errors.Errorf("must specify domain: %v and current cluster: %v", domainName, currentCluster),
		)
	}

	logger := loggerimpl.NewNopLogger()
	session := connectToCassandra(c)

	metadataV1Mgr := cassandra.NewMetadataPersistenceWithSession(session, currentCluster, logger)
	metadataV2Mgr := cassandra.NewMetadataPersistenceV2WithSession(session, currentCluster, logger)

	domainMigration := &domainMigrationImpl{
		metadataV1Mgr: metadataV1Mgr,
		metadataV2Mgr: metadataV2Mgr.(*cassandra.CassandraMetadataPersistenceV2),
	}

	domainMigration.migrateDomainFromV1ToV2(domainName)
}

func (d *domainMigrationImpl) migrateDomainFromV1ToV2(domainName string) {
	req := &persistence.GetDomainRequest{
		Name: domainName,
	}
	respV1, err := d.metadataV1Mgr.GetDomain(req)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("fail to get domain: %v in v1 table", domainName), err)
	}

	respV2, err := d.metadataV2Mgr.GetDomain(req)
	switch err.(type) {
	case *shared.EntityNotExistsError:
		_, err = d.metadataV2Mgr.CreateDomainInV2Table(&persistence.CreateDomainRequest{
			Info:              respV1.Info,
			Config:            respV1.Config,
			ReplicationConfig: respV1.ReplicationConfig,
			IsGlobalDomain:    false,
			ConfigVersion:     0,
			FailoverVersion:   common.EmptyVersion,
		})
		if err != nil {
			ErrorAndExit(fmt.Sprintf("fail to migrate domain: %v to v2 table", domainName), err)
		}
	case nil:
		// this means domain with same name exists in v2 table
		// do some sanity check
		if respV2.Info.ID == respV1.Info.ID {
			fmt.Printf("domain: %v is already migrated to v2 table", domainName)
		} else {
			ErrorAndExit(fmt.Sprintf("fail to migrate domain: %v to v2 table due to domain name collition", domainName), err)
		}
	default:
		ErrorAndExit(fmt.Sprintf("fail to migrate domain: %v to v2 table", domainName), err)
	}

	respV1, err = d.metadataV1Mgr.GetDomain(req)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("fail to get domain: %v in v1 table", domainName), err)
	}
	respV2, err = d.metadataV2Mgr.GetDomain(req)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("fail to get domain: %v in v2 table", domainName), err)
	}

	print := func(value interface{}) string {
		bytes, _ := json.MarshalIndent(value, "", "  ")
		return string(bytes)
	}

	fmt.Printf("domain in v1 table: %v\n", print(respV1))
	fmt.Printf("domain in v2 table: %v\n", print(respV2))
	fmt.Println("domain migration is success")
}
