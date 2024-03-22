// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

package gocql

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

// Definition of all Consistency levels
const (
	Any Consistency = iota
	One
	Two
	Three
	Quorum
	All
	LocalQuorum
	EachQuorum
	LocalOne
)

// Definition of all SerialConsistency levels
const (
	Serial SerialConsistency = iota
	LocalSerial
)

func mustConvertConsistency(c Consistency) gocql.Consistency {
	switch c {
	case Any:
		return gocql.Any
	case One:
		return gocql.One
	case Two:
		return gocql.Two
	case Three:
		return gocql.Three
	case Quorum:
		return gocql.Quorum
	case All:
		return gocql.All
	case LocalQuorum:
		return gocql.LocalQuorum
	case EachQuorum:
		return gocql.EachQuorum
	case LocalOne:
		return gocql.LocalOne
	default:
		panic(fmt.Sprintf("Unknown gocql Consistency level: %v", c))
	}
}

func mustConvertSerialConsistency(c SerialConsistency) gocql.SerialConsistency {
	switch c {
	case Serial:
		return gocql.Serial
	case LocalSerial:
		return gocql.LocalSerial
	default:
		panic(fmt.Sprintf("Unknown gocql SerialConsistency level: %v", c))
	}
}

func (c Consistency) MarshalText() (text []byte, err error) {
	return []byte(c.String()), nil
}

func (c *Consistency) UnmarshalText(text []byte) error {
	switch string(text) {
	case "ANY":
		*c = Any
	case "ONE":
		*c = One
	case "TWO":
		*c = Two
	case "THREE":
		*c = Three
	case "QUORUM":
		*c = Quorum
	case "ALL":
		*c = All
	case "LOCAL_QUORUM":
		*c = LocalQuorum
	case "EACH_QUORUM":
		*c = EachQuorum
	case "LOCAL_ONE":
		*c = LocalOne
	default:
		return fmt.Errorf("invalid consistency %q", string(text))
	}

	return nil
}

func (c Consistency) String() string {
	switch c {
	case Any:
		return "ANY"
	case One:
		return "ONE"
	case Two:
		return "TWO"
	case Three:
		return "THREE"
	case Quorum:
		return "QUORUM"
	case All:
		return "ALL"
	case LocalQuorum:
		return "LOCAL_QUORUM"
	case EachQuorum:
		return "EACH_QUORUM"
	case LocalOne:
		return "LOCAL_ONE"
	default:
		return fmt.Sprintf("invalid consistency: %d", uint16(c))
	}
}

func ParseConsistency(s string) (Consistency, error) {
	var c Consistency
	if err := c.UnmarshalText([]byte(strings.ToUpper(s))); err != nil {
		return c, fmt.Errorf("parse consistency: %w", err)
	}
	return c, nil
}

func ParseSerialConsistency(s string) (SerialConsistency, error) {
	var sc SerialConsistency
	if err := sc.UnmarshalText([]byte(strings.ToUpper(s))); err != nil {
		return sc, fmt.Errorf("parse serial consistency: %w", err)

	}
	return sc, nil
}

func (s SerialConsistency) String() string {
	switch s {
	case Serial:
		return "SERIAL"
	case LocalSerial:
		return "LOCAL_SERIAL"
	default:
		return fmt.Sprintf("invalid serial consistency %d", uint16(s))
	}
}

func (s SerialConsistency) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

func (s *SerialConsistency) UnmarshalText(text []byte) error {
	switch string(text) {
	case "SERIAL":
		*s = Serial
	case "LOCAL_SERIAL":
		*s = LocalSerial
	default:
		return fmt.Errorf("invalid serial consistency %q", string(text))
	}

	return nil
}
