package blobstore

import (
	"gopkg.in/yaml.v2"
)

type (
	// Config describes the configuration needed to construct a blobstore client backed by file system
	Config struct {
		StoreDirectory         string `yaml:"storeDirectory"`
		DefaultBucket BucketConfig `yaml:"defaultBucket"`
		CustomBuckets []BucketConfig `yaml:"customBuckets"`
	}

	// BucketConfig describes the config for a bucket
	BucketConfig struct {
		Name string `yaml:"name"`
		Owner string `yaml:"owner"`
		RetentionDays int `yaml:"retentionDays"`
	}
)

func (c *Config) Validate() {
	if len(c.StoreDirectory) == 0 {
		panic("Empty store directory")
	}
	c.DefaultBucket.validate()
	for _, b := range c.CustomBuckets {
		b.validate()
	}
}

func (b *BucketConfig) validate() {
	if len(b.Name) == 0 {
		panic("Empty bucket name")
	}
	if len(b.Owner) == 0 {
		panic("Empty bucket owner")
	}
	if b.RetentionDays < 0 {
		panic("Negative retention days")
	}
}

func serialize(b *BucketConfig) ([]byte, error) {
	return yaml.Marshal(b)
}

func deserialize(data []byte) (*BucketConfig, error) {
	bc := &BucketConfig{}
	if err := yaml.Unmarshal(data, bc); err != nil {
		return nil, err
	}
	return bc, nil

}

