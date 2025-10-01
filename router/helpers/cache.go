package helpers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/angelbarreiros/Penguin/scheduler"
	"github.com/google/uuid"
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
		var cacheItem CacheItem = item.(CacheItem)
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

	var relevantHeaders []string = []string{"Accept", "Accept-Language", "Authorization"}
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

// UUID-based cache implementation with parametrized types

type uuidCleaner struct {
	key   uuid.UUID
	cache *sync.Map
}

func (c *uuidCleaner) Execute() []any {
	c.cache.Delete(c.key)
	return nil
}

// uuidCacheItem es privado - implementación interna
type uuidCacheItem[T any] struct {
	value      T
	expiration time.Duration
	generated  time.Time
}

func (c uuidCacheItem[T]) getValue() T {
	return c.value
}

func newUUIDCacheItem[T any](value T, expiration time.Duration) uuidCacheItem[T] {
	return uuidCacheItem[T]{
		value:      value,
		expiration: expiration,
		generated:  time.Now(),
	}
}

// UUIDCache es la estructura principal - solo algunos métodos serán públicos
type uuidCache[T any] struct {
	cache   *sync.Map
	cleaner *scheduler.Scheduler
}

func newUUIDCache[T any]() *uuidCache[T] {
	return &uuidCache[T]{
		cache:   &sync.Map{},
		cleaner: scheduler.StartScheduler(),
	}
}

func (c *uuidCache[T]) Close() {
	defer c.cleaner.Stop()
	defer c.cache.Clear()
}

func (c *uuidCache[T]) set(key uuid.UUID, item uuidCacheItem[T]) {
	c.cache.Store(key, item)
	var job = scheduler.JobFunction(&uuidCleaner{cache: c.cache, key: key})
	c.cleaner.ScheduleJob(time.Now().Add(item.expiration), job)
}

// Store - Método público para almacenar un valor con clave específica
func (c *uuidCache[T]) Store(key uuid.UUID, value T, expiration time.Duration) {
	item := newUUIDCacheItem(value, expiration)
	c.set(key, item)
}

// StoreWithNewKey - Método público para almacenar un valor generando una nueva clave UUID
func (c *uuidCache[T]) StoreWithNewKey(value T, expiration time.Duration) uuid.UUID {
	key := uuid.New()
	item := newUUIDCacheItem(value, expiration)
	c.set(key, item)
	return key
}

func (c *uuidCache[T]) get(key uuid.UUID) (uuidCacheItem[T], bool) {
	if item, ok := c.cache.Load(key); ok {
		var cacheItem uuidCacheItem[T] = item.(uuidCacheItem[T])
		return cacheItem, true
	}
	return uuidCacheItem[T]{}, false
}

// Load - Método público para obtener un valor del cache
func (c *uuidCache[T]) Load(key uuid.UUID) (T, bool) {
	if item, ok := c.get(key); ok {
		return item.getValue(), true
	}
	var zero T
	return zero, false
}

// Delete - Método público para eliminar una entrada del cache
func (c *uuidCache[T]) Delete(key uuid.UUID) {
	c.cache.Delete(key)
}

// Has - Método público para verificar si existe una clave
func (c *uuidCache[T]) Has(key uuid.UUID) bool {
	_, ok := c.cache.Load(key)
	return ok
}

// Range - Método público para iterar sobre todas las entradas del cache
func (c *uuidCache[T]) Range(fn func(key uuid.UUID, value T) bool) {
	c.cache.Range(func(k, v interface{}) bool {
		key := k.(uuid.UUID)
		item := v.(uuidCacheItem[T])
		return fn(key, item.getValue())
	})
}

// Keys - Método público para obtener todas las claves
func (c *uuidCache[T]) Keys() []uuid.UUID {
	var keys []uuid.UUID
	c.cache.Range(func(k, v interface{}) bool {
		keys = append(keys, k.(uuid.UUID))
		return true
	})
	return keys
}

// Values - Método público para obtener todos los valores
func (c *uuidCache[T]) Values() []T {
	var values []T
	c.cache.Range(func(k, v interface{}) bool {
		item := v.(uuidCacheItem[T])
		values = append(values, item.getValue())
		return true
	})
	return values
}

// GetAll - Método público para obtener un mapa regular (copia) para range tradicional
func (c *uuidCache[T]) GetAll() map[uuid.UUID]T {
	result := make(map[uuid.UUID]T)
	c.cache.Range(func(k, v interface{}) bool {
		key := k.(uuid.UUID)
		item := v.(uuidCacheItem[T])
		result[key] = item.getValue()
		return true
	})
	return result
}

// Len - Método público para obtener el número de elementos en el cache
func (c *uuidCache[T]) Len() int {
	count := 0
	c.cache.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}

// Generic UUID cache instance management - privado
var (
	uuidCacheInstances = make(map[string]any)
	uuidCacheMutex     sync.RWMutex
)

// UUIDCache - Interfaz pública que expone solo los métodos necesarios
type UUIDCache[T any] interface {
	Store(key uuid.UUID, value T, expiration time.Duration)
	StoreWithNewKey(value T, expiration time.Duration) uuid.UUID
	Load(key uuid.UUID) (T, bool)
	Delete(key uuid.UUID)
	Has(key uuid.UUID) bool
	// Métodos de iteración
	Range(fn func(key uuid.UUID, value T) bool)
	Keys() []uuid.UUID
	Values() []T
	GetAll() map[uuid.UUID]T
	Len() int
}

// NewUUIDCache - Función pública para crear una nueva instancia de cache
func NewUUIDCache[T any]() UUIDCache[T] {
	return newUUIDCache[T]()
}

// GetUUIDCacheInstance - Función pública para obtener una instancia singleton
func GetUUIDCacheInstance[T any](cacheType string) UUIDCache[T] {
	uuidCacheMutex.RLock()
	if instance, ok := uuidCacheInstances[cacheType]; ok {
		uuidCacheMutex.RUnlock()
		return instance.(*uuidCache[T])
	}
	uuidCacheMutex.RUnlock()

	uuidCacheMutex.Lock()
	defer uuidCacheMutex.Unlock()

	// Double check after acquiring write lock
	if instance, ok := uuidCacheInstances[cacheType]; ok {
		return instance.(*uuidCache[T])
	}

	newCache := newUUIDCache[T]()
	uuidCacheInstances[cacheType] = newCache
	return newCache
}
