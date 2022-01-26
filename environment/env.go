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

// SetupEnv setup the necessary env
func SetupEnv() {
	if os.Getenv(CassandraSeeds) == "" {
		err := os.Setenv(CassandraSeeds, Localhost)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", CassandraSeeds))
		}
	}

	if os.Getenv(CassandraPort) == "" {
		err := os.Setenv(CassandraPort, CassandraDefaultPort)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", CassandraPort))
		}
	}

	if os.Getenv(CassandraProtoVersion) == "" {
		err := os.Setenv(CassandraProtoVersion, CassandraDefaultProtoVersion)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", CassandraProtoVersion))
		}
	}

	if os.Getenv(MySQLSeeds) == "" {
		err := os.Setenv(MySQLSeeds, Localhost)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", MySQLSeeds))
		}
	}

	if os.Getenv(MySQLPort) == "" {
		err := os.Setenv(MySQLPort, MySQLDefaultPort)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", MySQLPort))
		}
	}

	if os.Getenv(PostgresSeeds) == "" {
		err := os.Setenv(PostgresSeeds, Localhost)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", PostgresSeeds))
		}
	}

	if os.Getenv(PostgresPort) == "" {
		err := os.Setenv(PostgresPort, PostgresDefaultPort)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", PostgresPort))
		}
	}

	if os.Getenv(KafkaSeeds) == "" {
		err := os.Setenv(KafkaSeeds, Localhost)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", KafkaSeeds))
		}
	}

	if os.Getenv(KafkaPort) == "" {
		err := os.Setenv(KafkaPort, KafkaDefaultPort)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", KafkaPort))
		}
	}

	if os.Getenv(ESSeeds) == "" {
		err := os.Setenv(ESSeeds, Localhost)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", ESSeeds))
		}
	}

	if os.Getenv(ESPort) == "" {
		err := os.Setenv(ESPort, ESDefaultPort)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", ESPort))
		}
	}

	if os.Getenv(CLITransportProtocol) == "" {
		err := os.Setenv(CLITransportProtocol, DefaultCLITransportProtocol)
		if err != nil {
			panic(fmt.Sprintf("error setting env %v", CLITransportProtocol))
		}
	}
}

// GetCassandraAddress return the cassandra address
func GetCassandraAddress() string {
	addr := os.Getenv(CassandraSeeds)
	if addr == "" {
		addr = Localhost
	}
	return addr
}

// GetCassandraPort return the cassandra port
func GetCassandraPort() int {
	port := os.Getenv(CassandraPort)
	if port == "" {
		port = CassandraDefaultPort
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("error getting env %v", CassandraPort))
	}
	return p
}

// GetCassandraUsername return the cassandra username
func GetCassandraUsername() string {
	user := os.Getenv(CassandraUsername)
	if user == "" {
		user = CassandraDefaultUsername
	}
	return user
}

// GetCassandraPassword return the cassandra password
func GetCassandraPassword() string {
	pass := os.Getenv(CassandraPassword)
	if pass == "" {
		pass = CassandraDefaultPassword
	}
	return pass
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
func GetCassandraProtoVersion() int {
	protoVersion := os.Getenv(CassandraProtoVersion)
	if protoVersion == "" {
		protoVersion = CassandraDefaultProtoVersion
	}
	p, err := strconv.Atoi(protoVersion)
	if err != nil {
		panic(fmt.Sprintf("error getting env %v", CassandraProtoVersion))
	}
	return p
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
func GetMySQLPort() int {
	port := os.Getenv(MySQLPort)
	if port == "" {
		port = MySQLDefaultPort
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("error getting env %v", MySQLPort))
	}
	return p
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
func GetPostgresPort() int {
	port := os.Getenv(PostgresPort)
	if port == "" {
		port = PostgresDefaultPort
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("error getting env %v", PostgresPort))
	}
	return p
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
func GetMongoPort() int {
	port := os.Getenv(MongoPort)
	if port == "" {
		port = MongoDefaultPort
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("error getting env %v", MongoPort))
	}
	return p
}
