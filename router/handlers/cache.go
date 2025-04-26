package handlers

import (
	"angelotero/commonBackend/scheduler"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cleaner struct {
	key   string
	cache *sync.Map
}

func (c *cleaner) Execute() []any {
	c.cache.Delete(c.key)
	return nil
}

type cleanerCache struct {
	cache   *sync.Map
	cleaner *scheduler.Scheduler
}
type CacheItem struct {
	value      any
	expiration time.Duration
	generated  time.Time
}

func (c CacheItem) GetValue() any {
	return c.value
}
func NewCacheItem(value any, expiration time.Duration) CacheItem {
	return CacheItem{
		value:      value,
		expiration: expiration,
		generated:  time.Now(),
	}
}

var (
	instance *cleanerCache
	once     sync.Once
)

func NewCleanerCacheInstance() *cleanerCache {
	if instance == nil {
		once.Do(func() {
			instance = &cleanerCache{
				cache:   &sync.Map{},
				cleaner: scheduler.StartScheduler(),
			}
		})
	}
	return instance
}
func (c *cleanerCache) Close() {
	defer c.cleaner.Stop()
	defer c.cache.Clear()
}
func (c *cleanerCache) Set(key string, i CacheItem) {
	c.cache.Store(key, i)
	var job = scheduler.JobFunction(&cleaner{cache: c.cache, key: key})
	c.cleaner.ScheduleJob(time.Now().Add(i.expiration), job)
}
func (c *cleanerCache) Get(w http.ResponseWriter, key string) (CacheItem, bool) {
	if item, ok := c.cache.Load(key); ok {
		cacheItem := item.(CacheItem)
		if w != nil {
			w.Header().Set("Cache-Status", "HIT")

			var sb strings.Builder
			sb.WriteString("max-age=")
			sb.WriteString(strconv.Itoa(int(cacheItem.expiration.Seconds())))
			w.Header().Set("Cache-Control", sb.String())

			sb.Reset()
			sb.WriteString(strconv.Itoa(int(time.Since(cacheItem.generated).Seconds())))
			w.Header().Set("Age", sb.String())
		}

		return cacheItem, true
	}
	if w != nil {
		w.Header().Set("Cache-Status", "MISS")
	}

	return CacheItem{}, false
}
func GenerateCacheKey(r *http.Request) string {
	var sb strings.Builder

	sb.WriteString(r.Method)
	sb.WriteString(":")
	sb.WriteString(r.URL.Path)

	if len(r.URL.Query()) > 0 {
		params := make([]string, 0, len(r.URL.Query()))
		for k := range r.URL.Query() {
			params = append(params, k+"="+r.URL.Query().Get(k))
		}
		sort.Strings(params)
		sb.WriteString("?")
		sb.WriteString(strings.Join(params, "&"))
	}

	relevantHeaders := []string{"Accept", "Accept-Language", "Authorization"}
	for _, header := range relevantHeaders {
		if value := r.Header.Get(header); value != "" {
			sb.WriteString("#")
			sb.WriteString(header)
			sb.WriteString("=")
			sb.WriteString(value)
		}
	}

	return sb.String()
}
