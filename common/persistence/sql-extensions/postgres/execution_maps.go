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
	deleteMapQueryTemplate = `DELETE FROM %v
WHERE
shard_id = $1 AND
domain_id = $2 AND
workflow_id = $3 AND
run_id = $4`

	// %[2]v is the columns of the value struct (i.e. no primary key columns), comma separated
	// %[3]v should be %[2]v with colons prepended.
	// i.e. %[3]v = ",".join(":" + s for s in %[2]v)
	// So that this query can be used with BindNamed
	// %[4]v should be the name of the key associated with the map
	// e.g. for ActivityInfo it is "schedule_id"
	setKeyInMapQueryTemplate = `REPLACE INTO %[1]v
(shard_id, domain_id, workflow_id, run_id, %[4]v, %[2]v)
VALUES
(:shard_id, :domain_id, :workflow_id, :run_id, :%[4]v, %[3]v)`

	// %[2]v is the name of the key
	deleteKeyInMapQueryTemplate = `DELETE FROM %[1]v
WHERE
shard_id = $1 AND
domain_id = $2 AND
workflow_id = $3 AND
run_id = $4 AND
%[2]v = $5`

	// %[1]v is the name of the table
	// %[2]v is the name of the key
	// %[3]v is the value columns, separated by commas
	getMapQueryTemplate = `SELECT %[2]v, %[3]v FROM %[1]v
WHERE
shard_id = $1 AND
domain_id = $2 AND
workflow_id = $3 AND
run_id = $4`
)

func (d *driver) DeleteMapQueryTemplate() string {
	return deleteMapQueryTemplate
}

func (d *driver) SetKeyInMapQueryTemplate() string {
	return setKeyInMapQueryTemplate
}

func (d *driver) DeleteKeyInMapQueryTemplate() string {
	return deleteKeyInMapQueryTemplate
}

func (d *driver) GetMapQueryTemplate() string {
	return getMapQueryTemplate
}

const (
	deleteAllSignalsRequestedSetQuery = `DELETE FROM signals_requested_sets
WHERE
shard_id = $1 AND
domain_id = $2 AND
workflow_id = $3 AND
run_id = $4
`

	createSignalsRequestedSetQuery = `INSERT IGNORE INTO signals_requested_sets
(shard_id, domain_id, workflow_id, run_id, signal_id) VALUES
(:shard_id, :domain_id, :workflow_id, :run_id, :signal_id)`

	deleteSignalsRequestedSetQuery = `DELETE FROM signals_requested_sets
WHERE
shard_id = $1 AND
domain_id = $2 AND
workflow_id = $3 AND
run_id = $4 AND
signal_id = $5`

	getSignalsRequestedSetQuery = `SELECT signal_id FROM signals_requested_sets WHERE
shard_id = $1 AND
domain_id = $2 AND
workflow_id = $3 AND
run_id = $4`
)

func (d *driver) DeleteAllSignalsRequestedSetQuery() string {
	return deleteAllSignalsRequestedSetQuery
}

func (d *driver) CreateSignalsRequestedSetQuery() string {
	return createSignalsRequestedSetQuery
}

func (d *driver) DeleteSignalsRequestedSetQuery() string {
	return deleteSignalsRequestedSetQuery
}

func (d *driver) GetSignalsRequestedSetQuery() string {
	return getSignalsRequestedSetQuery
}
