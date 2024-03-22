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
	"fmt"
	"os"
	"strconv"
)

const (
	// Localhost default localhost
	Localhost = "127.0.0.1"

	// CassandraSeeds env
	CassandraSeeds = "CASSANDRA_SEEDS"
	// CassandraPort env
	CassandraPort = "CASSANDRA_DB_PORT"
	// CassandraDefaultPort Cassandra default port
	CassandraDefaultPort = "9042"
	// CassandraUsername env
	CassandraUsername = "CASSANDRA_DB_USERNAME"
	// CassandraDefaultUsername Cassandra default username
	CassandraDefaultUsername = ""
	// CassandraPassword env
	CassandraPassword = "CASSANDRA_DB_PASSWORD"
	// CassandraDefaultPassword Cassandra default password
	CassandraDefaultPassword = ""
	// CassandraAllowedAuthenticators env
	CassandraAllowedAuthenticators = "CASSANDRA_DB_ALLOWED_AUTHENTICATORS"
	// CassandraProtoVersion env
	CassandraProtoVersion = "CASSANDRA_PROTO_VERSION"
	// CassandraDefaultProtoVersion Cassandra default protocol version
	CassandraDefaultProtoVersion = "4"
	// CassandraDefaultProtoVersionInteger Cassandra default protocol version int version
	CassandraDefaultProtoVersionInteger = 4

	// MySQLSeeds env
	MySQLSeeds = "MYSQL_SEEDS"
	// MySQLPort env
	MySQLPort = "MYSQL_PORT"
	// MySQLDefaultPort is MySQL default port
	MySQLDefaultPort = "3306"
	// MySQLUser env
	MySQLUser = "MYSQL_USER"
	// MySQLDefaultUser is default user
	MySQLDefaultUser = "root"
	// MySQLPassword env
	MySQLPassword = "MYSQL_PASSWORD"
	// MySQLDefaultPassword is default password
	MySQLDefaultPassword = "cadence"

	// MongoSeeds env
	MongoSeeds = "MONGO_SEEDS"
	// MongoPort env
	MongoPort = "MONGO_PORT"
	// MongoDefaultPort is Mongo default port
	MongoDefaultPort = "27017"

	// KafkaSeeds env
	KafkaSeeds = "KAFKA_SEEDS"
	// KafkaPort env
	KafkaPort = "KAFKA_PORT"
	// KafkaDefaultPort Kafka default port
	KafkaDefaultPort = "9092"

	// ESSeeds env
	ESSeeds = "ES_SEEDS"
	// ESPort env
	ESPort = "ES_PORT"
	// ESDefaultPort ES default port
	ESDefaultPort = "9200"
	// ESVersion is the ElasticSearch version
	ESVersion = "ES_VERSION"
	// ESDefaultVersion is the default version
	ESDefaultVersion = "v6"

	// PostgresSeeds env
	PostgresSeeds = "POSTGRES_SEEDS"
	// PostgresPort env
	PostgresPort = "POSTGRES_PORT"
	// PostgresDefaultPort Postgres default port
	PostgresDefaultPort = "5432"

	// CLITransportProtocol env
	CLITransportProtocol = "CADENCE_CLI_TRANSPORT_PROTOCOL"
	// DefaultCLITransportProtocol  CLI default channel
	DefaultCLITransportProtocol = "tchannel"
)

var envDefaults = map[string]string{
	CassandraSeeds:        Localhost,
	CassandraPort:         CassandraDefaultPort,
	CassandraProtoVersion: CassandraDefaultProtoVersion,
	MySQLSeeds:            Localhost,
	MySQLPort:             MySQLDefaultPort,
	PostgresSeeds:         Localhost,
	PostgresPort:          PostgresDefaultPort,
	KafkaSeeds:            Localhost,
	KafkaPort:             KafkaDefaultPort,
	ESSeeds:               Localhost,
	ESPort:                ESDefaultPort,
	CLITransportProtocol:  DefaultCLITransportProtocol,
}

// SetupEnv setup the necessary env
func SetupEnv() error {
	for k, v := range envDefaults {
		if os.Getenv(k) == "" {
			if err := setEnv(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetCassandraAddress return the cassandra address
func GetCassandraAddress() string {
	if addr := os.Getenv(CassandraSeeds); addr != "" {
		return addr
	}

	return envDefaults[CassandraSeeds]
}

// GetCassandraPort return the cassandra port
func GetCassandraPort() (int, error) {
	port := os.Getenv(CassandraPort)
	if port == "" {
		port = envDefaults[CassandraPort]
	}

	return strconv.Atoi(port)
}

// GetCassandraUsername return the cassandra username
func GetCassandraUsername() string {
	user := os.Getenv(CassandraUsername)
	if user != "" {
		return user
	}
	return CassandraDefaultUsername
}

// GetCassandraPassword return the cassandra password
func GetCassandraPassword() string {
	pass := os.Getenv(CassandraPassword)
	if pass != "" {
		return pass
	}

	return CassandraDefaultPassword
}

// GetCassandraAllowedAuthenticators return the cassandra allowed authenticators
func GetCassandraAllowedAuthenticators() []string {
	var authenticators []string
	configuredAuthenticators := os.Getenv(CassandraAllowedAuthenticators)
	if configuredAuthenticators == "" {
		return authenticators
	}

	authenticators = append(authenticators, configuredAuthenticators)
	return authenticators
}

// GetCassandraProtoVersion return the cassandra protocol version
func GetCassandraProtoVersion() (int, error) {
	protoVersion := os.Getenv(CassandraProtoVersion)
	if protoVersion == "" {
		protoVersion = envDefaults[CassandraProtoVersion]
	}

	return strconv.Atoi(protoVersion)
}

// GetMySQLAddress return the MySQL address
func GetMySQLAddress() string {
	addr := os.Getenv(MySQLSeeds)
	if addr == "" {
		addr = Localhost
	}
	return addr
}

// GetMySQLPort return the MySQL port
func GetMySQLPort() (int, error) {
	port := os.Getenv(MySQLPort)
	if port == "" {
		port = MySQLDefaultPort
	}

	return strconv.Atoi(port)
}

// GetMySQLUser return the MySQL user
func GetMySQLUser() string {
	user := os.Getenv(MySQLUser)
	if user == "" {
		user = MySQLDefaultUser
	}
	return user
}

// GetMySQLPassword return the MySQL password
func GetMySQLPassword() string {
	pw := os.Getenv(MySQLPassword)
	if pw == "" {
		pw = MySQLDefaultPassword
	}
	return pw
}

// GetPostgresAddress return the Postgres address
func GetPostgresAddress() string {
	addr := os.Getenv(PostgresSeeds)
	if addr == "" {
		addr = Localhost
	}
	return addr
}

// GetPostgresPort return the Postgres port
func GetPostgresPort() (int, error) {
	port := os.Getenv(PostgresPort)
	if port == "" {
		port = PostgresDefaultPort
	}

	return strconv.Atoi(port)
}

// GetESVersion return the ElasticSearch version
func GetESVersion() string {
	version := os.Getenv(ESVersion)
	if version == "" {
		version = ESDefaultVersion
	}
	return version
}

// GetMongoAddress return the MySQL address
func GetMongoAddress() string {
	addr := os.Getenv(MongoSeeds)
	if addr == "" {
		addr = Localhost
	}
	return addr
}

// GetMongoPort return the MySQL port
func GetMongoPort() (int, error) {
	port := os.Getenv(MongoPort)
	if port == "" {
		port = MongoDefaultPort
	}

	return strconv.Atoi(port)
}

func setEnv(key string, val string) error {
	if err := os.Setenv(key, val); err != nil {
		return fmt.Errorf("setting env %q: %w", key, err)
	}
	return nil
}
