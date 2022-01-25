// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/service/worker/scanner/executions"
)

// AdminDBDataDecodeThrift is the command to decode thrift binary into JSON
func AdminDBDataDecodeThrift(c *cli.Context) {
	input := getRequiredOption(c, FlagInput)

	encoder := codec.NewThriftRWEncoder()
	dataInput, err := hex.DecodeString(input)
	if err != nil {
		ErrorAndExit("input is not a valid hex string", err)
	}
	// this is an inconsistency in the code base, some place use ThriftRWEncoder(version0Thriftrw.go) some use thriftEncoder(thrift_encoder.go)
	dataWithPrepend := []byte{0x59}
	dataWithPrepend = append(dataWithPrepend, dataInput...)
	datas := [][]byte{dataInput, dataWithPrepend}

	found := false
	for _, data := range datas {
		for typeName, t := range decodingTypes {
			err = encoder.Decode(data, t)
			if err == nil {
				// encoding back to confirm
				data2, err := encoder.Encode(t)
				if err != nil {
					ErrorAndExit("cannot encode back to confirm", err)
				}
				if bytes.Compare(data, data2) == 0 {
					fmt.Printf("=======Decode into type %v ========\n", typeName)
					fmt.Println(anyToString(t, true, 0))
					found = true
				}
			}
		}
	}

	if !found {
		ErrorAndExit("input data cannot be decoded into any struct", nil)
	}

}

// AdminDBClean is the command to clean up unhealthy executions.
// Input is a JSON stream provided via STDIN or a file.
func AdminDBClean(c *cli.Context) {
	scanType, err := executions.ScanTypeString(getRequiredOption(c, FlagScanType))

	if err != nil {
		ErrorAndExit("unknown scan type", err)
	}
	collectionSlice := c.StringSlice(FlagInvariantCollection)
	blob := scanType.ToBlobstoreEntity()

	var collections []invariant.Collection
	for _, v := range collectionSlice {
		collection, err := invariant.CollectionString(v)
		if err != nil {
			ErrorAndExit("unknown invariant collection", err)
		}
		collections = append(collections, collection)
	}

	invariants := scanType.ToInvariants(collections)
	if len(invariants) < 1 {
		ErrorAndExit(
			fmt.Sprintf("no invariants for scantype %q and collections %q",
				scanType.String(),
				collectionSlice),
			nil,
		)
	}

	input := getInputFile(c.String(FlagInputFile))

	dec := json.NewDecoder(input)
	if err != nil {
		ErrorAndExit("", err)
	}
	var data []*store.ScanOutputEntity

	for {
		soe := &store.ScanOutputEntity{
			Execution: blob.Clone(),
		}

		if err := dec.Decode(&soe); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
		} else {
			data = append(data, soe)
		}
	}

	for _, e := range data {
		out := store.FixOutputEntity{
			Execution: e.Execution,
			Input:     *e,
			Result:    fixExecution(c, invariants, e),
		}
		data, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		fmt.Println(string(data))
	}
}

func fixExecution(
	c *cli.Context,
	invariants []executions.InvariantFactory,
	execution *store.ScanOutputEntity,
) invariant.ManagerFixResult {
	execManager := initializeExecutionStore(c, execution.Execution.(entity.Entity).GetShardID(), 0)
	defer execManager.Close()

	historyV2Mgr := initializeHistoryManager(c)
	defer historyV2Mgr.Close()

	pr := persistence.NewPersistenceRetryer(
		execManager,
		historyV2Mgr,
		common.CreatePersistenceRetryPolicy(),
	)

	var ivs []invariant.Invariant

	for _, fn := range invariants {
		ivs = append(ivs, fn(pr))
	}

	ctx, cancel := newContext(c)
	defer cancel()

	return invariant.NewInvariantManager(ivs).RunFixes(ctx, execution.Execution)
}
