// Copyright (c) 2016 Uber Technologies, Inc.
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

package environment

import (
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSetupEnv(t *testing.T) {
	if err := SetupEnv(); err != nil {
		t.Fatalf("SetupEnv() failed, err: %v", err)
	}
}

func TestGetCassandraAddress(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: CassandraSeeds,
			envVarVal: "",
			wantVal:   envDefaults[CassandraSeeds],
		},
		{
			name:      "custom address",
			envVarKey: CassandraSeeds,
			envVarVal: "casseed",
			wantVal:   "casseed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetCassandraAddress()

			if gotVal != tt.wantVal {
				t.Fatalf("GetCassandraAddress() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetCassandraPort(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantErr   bool
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: CassandraPort,
			envVarVal: "",
			wantErr:   false,
			wantVal:   mustConvertInt(t, envDefaults[CassandraPort]),
		},
		{
			name:      "non-int port value",
			envVarKey: CassandraPort,
			envVarVal: "xyz",
			wantErr:   true,
		},
		{
			name:      "custom port value",
			envVarKey: CassandraPort,
			envVarVal: "1001",
			wantVal:   1001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal, err := GetCassandraPort()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCassandraPort() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil || tt.wantErr {
				return
			}

			if gotVal != tt.wantVal {
				t.Fatalf("GetCassandraPort() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetCassandraUsername(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: CassandraUsername,
			envVarVal: "",
			wantVal:   CassandraDefaultUsername,
		},
		{
			name:      "custom user name",
			envVarKey: CassandraUsername,
			envVarVal: "user1",
			wantVal:   "user1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetCassandraUsername()

			if gotVal != tt.wantVal {
				t.Fatalf("GetCassandraUsername() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetCassandraPassword(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: CassandraPassword,
			envVarVal: "",
			wantVal:   CassandraDefaultUsername,
		},
		{
			name:      "custom pw",
			envVarKey: CassandraPassword,
			envVarVal: "xyz123",
			wantVal:   "xyz123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetCassandraPassword()

			if gotVal != tt.wantVal {
				t.Fatalf("GetCassandraPassword() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetCassandraAllowedAuthenticators(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   []string
	}{
		{
			name:      "default",
			envVarKey: CassandraAllowedAuthenticators,
			envVarVal: "",
			wantVal:   nil,
		},
		{
			name:      "custom authenticators",
			envVarKey: CassandraAllowedAuthenticators,
			envVarVal: "auth1",
			wantVal:   []string{"auth1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetCassandraAllowedAuthenticators()

			if diff := cmp.Diff(gotVal, tt.wantVal); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetCassandraProtoVersion(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantErr   bool
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: CassandraProtoVersion,
			envVarVal: "",
			wantErr:   false,
			wantVal:   mustConvertInt(t, envDefaults[CassandraProtoVersion]),
		},
		{
			name:      "non-int proto version",
			envVarKey: CassandraProtoVersion,
			envVarVal: "xyz",
			wantErr:   true,
		},
		{
			name:      "custom proto version",
			envVarKey: CassandraProtoVersion,
			envVarVal: "35",
			wantVal:   35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal, err := GetCassandraProtoVersion()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCassandraProtoVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil || tt.wantErr {
				return
			}

			if gotVal != tt.wantVal {
				t.Fatalf("GetCassandraProtoVersion() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetMySQLAddress(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: MySQLSeeds,
			envVarVal: "",
			wantVal:   Localhost,
		},
		{
			name:      "custom",
			envVarKey: MySQLSeeds,
			envVarVal: "mysqlseed",
			wantVal:   "mysqlseed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetMySQLAddress()

			if gotVal != tt.wantVal {
				t.Fatalf("GetMySQLAddress() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetMySQLPort(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantErr   bool
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: MySQLPort,
			envVarVal: "",
			wantErr:   false,
			wantVal:   mustConvertInt(t, MySQLDefaultPort),
		},
		{
			name:      "non-int port",
			envVarKey: MySQLPort,
			envVarVal: "xyz",
			wantErr:   true,
		},
		{
			name:      "custom port",
			envVarKey: MySQLPort,
			envVarVal: "8787",
			wantVal:   8787,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal, err := GetMySQLPort()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetMySQLPort() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil || tt.wantErr {
				return
			}

			if gotVal != tt.wantVal {
				t.Fatalf("GetMySQLPort() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetMySQLUser(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: MySQLUser,
			envVarVal: "",
			wantVal:   MySQLDefaultUser,
		},
		{
			name:      "custom",
			envVarKey: MySQLUser,
			envVarVal: "mysqluser",
			wantVal:   "mysqluser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetMySQLUser()

			if gotVal != tt.wantVal {
				t.Fatalf("GetMySQLUser() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetMySQLPassword(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: MySQLPassword,
			envVarVal: "",
			wantVal:   MySQLDefaultPassword,
		},
		{
			name:      "custom",
			envVarKey: MySQLPassword,
			envVarVal: "klmno",
			wantVal:   "klmno",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetMySQLPassword()

			if gotVal != tt.wantVal {
				t.Fatalf("GetMySQLPassword() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetPostgresAddress(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: PostgresSeeds,
			envVarVal: "",
			wantVal:   Localhost,
		},
		{
			name:      "custom",
			envVarKey: PostgresSeeds,
			envVarVal: "postgresseed",
			wantVal:   "postgresseed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetPostgresAddress()

			if gotVal != tt.wantVal {
				t.Fatalf("GetPostgresAddress() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetPostgresPort(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantErr   bool
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: PostgresPort,
			envVarVal: "",
			wantErr:   false,
			wantVal:   mustConvertInt(t, PostgresDefaultPort),
		},
		{
			name:      "non-int port",
			envVarKey: PostgresPort,
			envVarVal: "xyz",
			wantErr:   true,
		},
		{
			name:      "custom port",
			envVarKey: PostgresPort,
			envVarVal: "8787",
			wantVal:   8787,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal, err := GetPostgresPort()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetPostgresPort() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil || tt.wantErr {
				return
			}

			if gotVal != tt.wantVal {
				t.Fatalf("GetPostgresPort() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetESVersion(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: ESVersion,
			envVarVal: "",
			wantVal:   ESDefaultVersion,
		},
		{
			name:      "custom",
			envVarKey: ESVersion,
			envVarVal: "v1234",
			wantVal:   "v1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetESVersion()

			if gotVal != tt.wantVal {
				t.Fatalf("GetESVersion() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetMongoAddress(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: MongoSeeds,
			envVarVal: "",
			wantVal:   Localhost,
		},
		{
			name:      "custom",
			envVarKey: MongoSeeds,
			envVarVal: "mongoseed",
			wantVal:   "mongoseed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal := GetMongoAddress()

			if gotVal != tt.wantVal {
				t.Fatalf("GetMongoAddress() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestGetMongoPort(t *testing.T) {
	tests := []struct {
		name      string
		envVarKey string
		envVarVal string
		wantErr   bool
		wantVal   any
	}{
		{
			name:      "default",
			envVarKey: MongoPort,
			envVarVal: "",
			wantErr:   false,
			wantVal:   mustConvertInt(t, MongoDefaultPort),
		},
		{
			name:      "non-int port",
			envVarKey: MongoPort,
			envVarVal: "xyz",
			wantErr:   true,
		},
		{
			name:      "custom port",
			envVarKey: MongoPort,
			envVarVal: "8787",
			wantVal:   8787,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVarKey, tt.envVarVal)
			gotVal, err := GetMongoPort()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetMongoPort() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil || tt.wantErr {
				return
			}

			if gotVal != tt.wantVal {
				t.Fatalf("GetMongoPort() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func mustConvertInt(t *testing.T, s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("failed to convert string to int, err: %v", err)
	}
	return v
}
