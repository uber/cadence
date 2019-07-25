package archiver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type (
	URISuite struct {
		*require.Assertions
		suite.Suite
	}
)

func TestURISuite(t *testing.T) {
	suite.Run(t, new(URISuite))
}

func (s *URISuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *URISuite) TestURI() {
	testCases := []struct {
		URIString string
		valid     bool
		scheme    string
		path      string
		hostname  string
		port      string
		username  string
		password  string
	}{
		{
			URIString: "",
			valid:     false,
		},
		{
			URIString: "some random string",
			valid:     false,
		},
		{
			URIString: "mailto:a@b.com",
			valid:     false,
		},
		{
			URIString: "test://",
			valid:     true,
			scheme:    "test",
		},
		{
			URIString: "http://example.com/path",
			valid:     true,
			scheme:    "http",
			hostname:  "example.com",
			path:      "/path",
		},
		{
			URIString: "http://example.com/path with space",
			valid:     true,
			scheme:    "http",
			hostname:  "example.com",
			path:      "/path with space",
		},
		{
			URIString: "https://localhost:8080",
			valid:     true,
			scheme:    "https",
			hostname:  "localhost",
			port:      "8080",
		},
		{
			URIString: "file:///absolute/path/to/dir",
			valid:     true,
			scheme:    "file",
			path:      "/absolute/path/to/dir",
		},
		{
			URIString: "test://person:password@host/path",
			valid:     true,
			scheme:    "test",
			hostname:  "host",
			path:      "/path",
			username:  "person",
			password:  "password",
		},
	}

	for _, tc := range testCases {
		URI, err := NewURI(tc.URIString)
		if !tc.valid {
			s.Error(err)
			continue
		}

		s.NoError(err)
		s.Equal(tc.scheme, URI.Scheme())
		s.Equal(tc.path, URI.Path())
		s.Equal(tc.hostname, URI.Hostname())
		s.Equal(tc.port, URI.Port())
		s.Equal(tc.username, URI.Username())
		s.Equal(tc.password, URI.Password())
	}
}
