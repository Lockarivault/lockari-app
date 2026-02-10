package cache

type CacheClientMessage string

const (
	CacheClientMessageCacheMiss        CacheClientMessage = "cache_miss"
	CacheClientMessageCacheHit         CacheClientMessage = "cache_hit"
	CacheClientMessageKeyNotFound      CacheClientMessage = "key_not_found"
	CacheClientMessageFailedToGetValue CacheClientMessage = "failed_to_get_value"
)

func (c CacheClientMessage) String() string {
	return string(c)
}
