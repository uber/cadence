package filestore

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ConfigSuite struct {
	*require.Assertions
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *ConfigSuite) TestValidate() {
	testCases := []struct{
		config *Config
		isValid bool
	}{
		{
			config: &Config{
				StoreDirectory: "",
			},
			isValid: false,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{},
			},
			isValid: false,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
				},
			},
			isValid: false,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
					Owner: "test-default-bucket-owner",
				},
			},
			isValid: true,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
					Owner: "test-default-bucket-owner",
					RetentionDays: -1,
				},
			},
			isValid: false,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
					Owner: "test-default-bucket-owner",
					RetentionDays: 10,
				},
			},
			isValid: true,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
					Owner: "test-default-bucket-owner",
					RetentionDays: 10,
				},
				CustomBuckets: []BucketConfig{},
			},
			isValid: true,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
					Owner: "test-default-bucket-owner",
					RetentionDays: 10,
				},
				CustomBuckets: []BucketConfig{
					{},
				},
			},
			isValid: false,
		},
		{
			config: &Config{
				StoreDirectory: "test-store-directory",
				DefaultBucket: BucketConfig{
					Name: "test-default-bucket-name",
					Owner: "test-default-bucket-owner",
					RetentionDays: 10,
				},
				CustomBuckets: []BucketConfig{
					{
						Name: "test-custom-bucket-name",
						Owner: "test-custom-bucket-owner",
						RetentionDays: 10,
					},
				},
			},
			isValid: true,
		},
	}

	for _, tc := range testCases {
		if tc.isValid {
			s.NoError(tc.config.Validate())
		} else {
			s.Error(tc.config.Validate())
		}
	}
}

func (s *ConfigSuite) TestSerialization() {
	inCfg := &BucketConfig{
		Name:          "test-custom-bucket-name",
		Owner:         "test-custom-bucket-owner",
		RetentionDays: 10,
	}
	bytes, err := serialize(inCfg)
	s.NoError(err)

	outCfg, err := deserialize(bytes)
	s.NoError(err)
	s.Equal(inCfg, outCfg)
}