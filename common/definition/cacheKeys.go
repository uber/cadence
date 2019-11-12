package definition

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	keySeparator       = "::"
	decimal            = 10
	cacheKeyTemplate   = "%v" + keySeparator + "%v"
	errorCacheKeyEmpty = "cache key cannot be empty"
)

const (
	reapplyEventCacheKey = iota
)

type (
	CacheKey interface {
		// Generate generates the cache key
		Generate() string
	}

	reapplyEventCacheKeyImpl struct {
		key string
	}
)

func NewReapplyEventCacheKeyImpl(
	runID string,
	eventID int64,
	version int64,
) *reapplyEventCacheKeyImpl {
	fields := []string{runID, strconv.FormatInt(eventID, decimal), strconv.FormatInt(version, decimal)}
	cacheKey := strings.Join(fields, keySeparator)

	return &reapplyEventCacheKeyImpl{
		key: fmt.Sprintf(cacheKeyTemplate, reapplyEventCacheKey, cacheKey),
	}
}

func (k *reapplyEventCacheKeyImpl) Generate() string {
	if len(k.key) == 0 {
		panic(errorCacheKeyEmpty)
	}
	return k.key
}
