package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/angelbarreiros/Penguin/scheduler"
	"github.com/google/uuid"
)

// expiringMap is the shared storage behind every keyed cache. Entries are
// stored as pointers so a scheduled cleaner only deletes the exact entry it
// was created for (CompareAndDelete) and never a value stored later under the
// same key. Expiry is also checked on every read, so correctness does not
// depend on the cleaner running.
type expiringMap[K comparable, V any] struct {
	cache     *sync.Map
	cleaner   *scheduler.Scheduler
	closeOnce sync.Once
}

type expiringEntry[V any] struct {
	value      V
	expiration time.Duration
	generated  time.Time
	expiresAt  time.Time // zero = never expires
}

func (e *expiringEntry[V]) expired(now time.Time) bool {
	return !e.expiresAt.IsZero() && !now.Before(e.expiresAt)
}

type entryCleaner[K comparable, V any] struct {
	cache *sync.Map
	key   K
	entry *expiringEntry[V]
}

func (c *entryCleaner[K, V]) Execute() []any {
	c.cache.CompareAndDelete(c.key, c.entry)
	return nil
}

func newExpiringMap[K comparable, V any]() *expiringMap[K, V] {
	return &expiringMap[K, V]{
		cache:   &sync.Map{},
		cleaner: scheduler.NewScheduler(),
	}
}

// store keeps value for expiration; expiration <= 0 means no expiry.
func (m *expiringMap[K, V]) store(key K, value V, expiration time.Duration) {
	now := time.Now()
	entry := &expiringEntry[V]{value: value, expiration: expiration, generated: now}
	if expiration > 0 {
		entry.expiresAt = now.Add(expiration)
	}
	m.cache.Store(key, entry)
	if expiration > 0 {
		// If scheduling fails (cache closed) the entry still expires lazily.
		_, _ = m.cleaner.ScheduleOneTimeJob(expiration, scheduler.JobFunction(&entryCleaner[K, V]{cache: m.cache, key: key, entry: entry}))
	}
}

func (m *expiringMap[K, V]) load(key K) (*expiringEntry[V], bool) {
	v, ok := m.cache.Load(key)
	if !ok {
		return nil, false
	}
	entry := v.(*expiringEntry[V])
	if entry.expired(time.Now()) {
		m.cache.CompareAndDelete(key, entry)
		return nil, false
	}
	return entry, true
}

func (m *expiringMap[K, V]) delete(key K) {
	m.cache.Delete(key)
}

func (m *expiringMap[K, V]) rangeLive(fn func(key K, entry *expiringEntry[V]) bool) {
	now := time.Now()
	m.cache.Range(func(k, v any) bool {
		entry := v.(*expiringEntry[V])
		if entry.expired(now) {
			m.cache.CompareAndDelete(k, entry)
			return true
		}
		return fn(k.(K), entry)
	})
}

func (m *expiringMap[K, V]) clear() {
	m.cache.Clear()
}

func (m *expiringMap[K, V]) close() {
	m.closeOnce.Do(func() {
		m.cleaner.Stop()
		m.cache.Clear()
	})
}

// HTTP response cache

type cleanerCache struct {
	entries *expiringMap[string, CacheItem]
}
type CacheItem struct {
	value      any
	expiration time.Duration
	generated  time.Time
}

func (c CacheItem) GetValue() any {
	return c.value
}

// NewCacheItem creates an item; expiration <= 0 means it never expires.
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
	once.Do(func() {
		instance = &cleanerCache{entries: newExpiringMap[string, CacheItem]()}
	})
	return instance
}
func (c *cleanerCache) Close() {
	c.entries.close()
}
func (c *cleanerCache) Set(key string, i CacheItem) {
	c.entries.store(key, i, i.expiration)
}
func (c *cleanerCache) Get(w http.ResponseWriter, key string) (CacheItem, bool) {
	entry, ok := c.entries.load(key)
	if !ok {
		if w != nil {
			w.Header().Set("Cache-Status", "MISS")
		}
		return CacheItem{}, false
	}
	cacheItem := entry.value
	if w != nil {
		w.Header().Set("Cache-Status", "HIT")
		// private: entries may be keyed on Authorization, so shared caches
		// must never store them.
		cacheControl := "private"
		if cacheItem.expiration > 0 {
			cacheControl += ", max-age=" + strconv.Itoa(int(cacheItem.expiration.Seconds()))
		}
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("Age", strconv.Itoa(int(time.Since(cacheItem.generated).Seconds())))
		w.Header().Add("Vary", "Accept, Accept-Language, Authorization")
	}
	return cacheItem, true
}

// GenerateCacheKey builds a key from the method, path, every query value
// (escaped and sorted) and the Accept, Accept-Language and Authorization
// headers. The Authorization value is hashed so tokens are never kept in memory
// as cache keys.
func GenerateCacheKey(r *http.Request) string {
	var sb strings.Builder

	sb.WriteString(r.Method)
	sb.WriteString(" ")
	sb.WriteString(url.PathEscape(r.URL.Path))

	if query := r.URL.Query(); len(query) > 0 {
		sb.WriteString("?")
		sb.WriteString(query.Encode())
	}

	for _, header := range []string{"Accept", "Accept-Language", "Authorization"} {
		value := r.Header.Get(header)
		if value == "" {
			continue
		}
		if header == "Authorization" {
			sum := sha256.Sum256([]byte(value))
			value = hex.EncodeToString(sum[:])
		}
		sb.WriteString("#")
		sb.WriteString(header)
		sb.WriteString("=")
		sb.WriteString(url.QueryEscape(value))
	}

	return sb.String()
}

func registryKey[T any](cacheType string) string {
	return cacheType + "|" + reflect.TypeFor[T]().String()
}

// UUID-based cache implementation with parametrized types

// UUIDCache es la estructura principal - solo algunos métodos serán públicos
type uuidCache[T any] struct {
	entries    *expiringMap[uuid.UUID, T]
	defaultTTL time.Duration
	ttlMutex   sync.RWMutex
}

func newUUIDCache[T any]() *uuidCache[T] {
	return &uuidCache[T]{
		entries:    newExpiringMap[uuid.UUID, T](),
		defaultTTL: 0, // No default expiration
	}
}

// Close stops the cache's cleaner and drops every entry.
func (c *uuidCache[T]) Close() {
	c.entries.close()
}

// Store - Método público para almacenar un valor con clave específica (expiration <= 0: sin expiración)
func (c *uuidCache[T]) Store(key uuid.UUID, value T, expiration time.Duration) {
	c.entries.store(key, value, expiration)
}

// StoreWithDefaultTTL - Método público para almacenar un valor usando el TTL predefinido
func (c *uuidCache[T]) StoreWithDefaultTTL(key uuid.UUID, value T) {
	c.entries.store(key, value, c.GetDefaultTTL())
}

// SetDefaultTTL - Establecer tiempo de expiración predefinido para todos los elementos nuevos
func (c *uuidCache[T]) SetDefaultTTL(duration time.Duration) {
	c.ttlMutex.Lock()
	defer c.ttlMutex.Unlock()
	c.defaultTTL = duration
}

// GetDefaultTTL - Obtener el TTL predefinido actual
func (c *uuidCache[T]) GetDefaultTTL() time.Duration {
	c.ttlMutex.RLock()
	defer c.ttlMutex.RUnlock()
	return c.defaultTTL
}

// StoreWithNewKey - Método público para almacenar un valor generando una nueva clave UUID
func (c *uuidCache[T]) StoreWithNewKey(value T, expiration time.Duration) uuid.UUID {
	key := uuid.New()
	c.Store(key, value, expiration)
	return key
}

// StoreWithNewKeyAndDefaultTTL - Método público para almacenar un valor con nueva clave UUID usando TTL predefinido
func (c *uuidCache[T]) StoreWithNewKeyAndDefaultTTL(value T) uuid.UUID {
	key := uuid.New()
	c.StoreWithDefaultTTL(key, value)
	return key
}

// Load - Método público para obtener un valor del cache
func (c *uuidCache[T]) Load(key uuid.UUID) (T, bool) {
	if entry, ok := c.entries.load(key); ok {
		return entry.value, true
	}
	var zero T
	return zero, false
}

// Delete - Método público para eliminar una entrada del cache
func (c *uuidCache[T]) Delete(key uuid.UUID) {
	c.entries.delete(key)
}

// Has - Método público para verificar si existe una clave
func (c *uuidCache[T]) Has(key uuid.UUID) bool {
	_, ok := c.entries.load(key)
	return ok
}

// Range - Método público para iterar sobre todas las entradas del cache
func (c *uuidCache[T]) Range(fn func(key uuid.UUID, value T) bool) {
	c.entries.rangeLive(func(key uuid.UUID, entry *expiringEntry[T]) bool {
		return fn(key, entry.value)
	})
}

// Keys - Método público para obtener todas las claves
func (c *uuidCache[T]) Keys() []uuid.UUID {
	var keys []uuid.UUID
	c.Range(func(key uuid.UUID, _ T) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// Values - Método público para obtener todos los valores
func (c *uuidCache[T]) Values() []T {
	var values []T
	c.Range(func(_ uuid.UUID, value T) bool {
		values = append(values, value)
		return true
	})
	return values
}

// GetAll - Método público para obtener un mapa regular (copia) para range tradicional
func (c *uuidCache[T]) GetAll() map[uuid.UUID]T {
	result := make(map[uuid.UUID]T)
	c.Range(func(key uuid.UUID, value T) bool {
		result[key] = value
		return true
	})
	return result
}

// Len - Método público para obtener el número de elementos en el cache
func (c *uuidCache[T]) Len() int {
	count := 0
	c.Range(func(uuid.UUID, T) bool {
		count++
		return true
	})
	return count
}

// Clear - Método público para limpiar completamente el cache
func (c *uuidCache[T]) Clear() {
	c.entries.clear()
}

// Generic UUID cache instance management - privado
var (
	uuidCacheInstances = make(map[string]any)
	uuidCacheMutex     sync.Mutex
)

// UUIDCache - Interfaz pública que expone solo los métodos necesarios
type UUIDCache[T any] interface {
	Store(key uuid.UUID, value T, expiration time.Duration)
	StoreWithDefaultTTL(key uuid.UUID, value T)
	StoreWithNewKey(value T, expiration time.Duration) uuid.UUID
	StoreWithNewKeyAndDefaultTTL(value T) uuid.UUID
	Load(key uuid.UUID) (T, bool)
	Delete(key uuid.UUID)
	Has(key uuid.UUID) bool
	// Métodos de iteración
	Range(fn func(key uuid.UUID, value T) bool)
	Keys() []uuid.UUID
	Values() []T
	GetAll() map[uuid.UUID]T
	Len() int
	// Control de TTL predefinido
	SetDefaultTTL(duration time.Duration)
	GetDefaultTTL() time.Duration
	// Control de limpieza
	Clear()
	Close()
}

// NewUUIDCache - Función pública para crear una nueva instancia de cache
func NewUUIDCache[T any]() UUIDCache[T] {
	return newUUIDCache[T]()
}

// GetUUIDCacheInstance - Función pública para obtener una instancia singleton.
// Instances are keyed by cacheType and T, so the same name with different
// value types yields separate caches.
func GetUUIDCacheInstance[T any](cacheType string) UUIDCache[T] {
	key := registryKey[T](cacheType)
	uuidCacheMutex.Lock()
	defer uuidCacheMutex.Unlock()
	if instance, ok := uuidCacheInstances[key]; ok {
		return instance.(*uuidCache[T])
	}
	newCache := newUUIDCache[T]()
	uuidCacheInstances[key] = newCache
	return newCache
}

// String-based cache implementation with parametrized types

// stringCache es la estructura principal - solo algunos métodos serán públicos
type stringCache[T any] struct {
	entries    *expiringMap[string, T]
	defaultTTL time.Duration
	ttlMutex   sync.RWMutex
}

func newStringCache[T any]() *stringCache[T] {
	return &stringCache[T]{
		entries:    newExpiringMap[string, T](),
		defaultTTL: 0, // No default expiration
	}
}

// Close stops the cache's cleaner and drops every entry.
func (c *stringCache[T]) Close() {
	c.entries.close()
}

// Store - Método público para almacenar un valor con clave específica (expiration <= 0: sin expiración)
func (c *stringCache[T]) Store(key string, value T, expiration time.Duration) {
	c.entries.store(key, value, expiration)
}

// StoreWithDefaultTTL - Método público para almacenar un valor usando el TTL predefinido
func (c *stringCache[T]) StoreWithDefaultTTL(key string, value T) {
	c.entries.store(key, value, c.GetDefaultTTL())
}

// SetDefaultTTL - Establecer tiempo de expiración predefinido para todos los elementos nuevos
func (c *stringCache[T]) SetDefaultTTL(duration time.Duration) {
	c.ttlMutex.Lock()
	defer c.ttlMutex.Unlock()
	c.defaultTTL = duration
}

// GetDefaultTTL - Obtener el TTL predefinido actual
func (c *stringCache[T]) GetDefaultTTL() time.Duration {
	c.ttlMutex.RLock()
	defer c.ttlMutex.RUnlock()
	return c.defaultTTL
}

// Load - Método público para obtener un valor del cache
func (c *stringCache[T]) Load(key string) (T, bool) {
	if entry, ok := c.entries.load(key); ok {
		return entry.value, true
	}
	var zero T
	return zero, false
}

// Delete - Método público para eliminar una entrada del cache
func (c *stringCache[T]) Delete(key string) {
	c.entries.delete(key)
}

// Has - Método público para verificar si existe una clave
func (c *stringCache[T]) Has(key string) bool {
	_, ok := c.entries.load(key)
	return ok
}

// Range - Método público para iterar sobre todas las entradas del cache
func (c *stringCache[T]) Range(fn func(key string, value T) bool) {
	c.entries.rangeLive(func(key string, entry *expiringEntry[T]) bool {
		return fn(key, entry.value)
	})
}

// Keys - Método público para obtener todas las claves
func (c *stringCache[T]) Keys() []string {
	var keys []string
	c.Range(func(key string, _ T) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// Values - Método público para obtener todos los valores
func (c *stringCache[T]) Values() []T {
	var values []T
	c.Range(func(_ string, value T) bool {
		values = append(values, value)
		return true
	})
	return values
}

// GetAll - Método público para obtener un mapa regular (copia) para range tradicional
func (c *stringCache[T]) GetAll() map[string]T {
	result := make(map[string]T)
	c.Range(func(key string, value T) bool {
		result[key] = value
		return true
	})
	return result
}

// Len - Método público para obtener el número de elementos en el cache
func (c *stringCache[T]) Len() int {
	count := 0
	c.Range(func(string, T) bool {
		count++
		return true
	})
	return count
}

// Clear - Método público para limpiar completamente el cache
func (c *stringCache[T]) Clear() {
	c.entries.clear()
}

// Generic string cache instance management - privado
var (
	stringCacheInstances = make(map[string]any)
	stringCacheMutex     sync.Mutex
)

// StringCache - Interfaz pública que expone solo los métodos necesarios
type StringCache[T any] interface {
	Store(key string, value T, expiration time.Duration)
	StoreWithDefaultTTL(key string, value T)
	Load(key string) (T, bool)
	Delete(key string)
	Has(key string) bool
	// Métodos de iteración
	Range(fn func(key string, value T) bool)
	Keys() []string
	Values() []T
	GetAll() map[string]T
	Len() int
	// Control de TTL predefinido
	SetDefaultTTL(duration time.Duration)
	GetDefaultTTL() time.Duration
	// Control de limpieza
	Clear()
	Close()
}

// NewStringCache - Función pública para crear una nueva instancia de cache de strings
func NewStringCache[T any]() StringCache[T] {
	return newStringCache[T]()
}

// GetStringCacheInstance - Función pública para obtener una instancia singleton de cache de strings.
// Instances are keyed by cacheType and T.
func GetStringCacheInstance[T any](cacheType string) StringCache[T] {
	key := registryKey[T](cacheType)
	stringCacheMutex.Lock()
	defer stringCacheMutex.Unlock()
	if instance, ok := stringCacheInstances[key]; ok {
		return instance.(*stringCache[T])
	}
	newCache := newStringCache[T]()
	stringCacheInstances[key] = newCache
	return newCache
}

// List-based cache implementation with automatic list expiration.
// Expiry is lazy: the whole list is treated as empty once its TTL has passed
// and is reset on the next write. No background job is used, so changing the
// TTL can never wipe items added afterwards.

// listCache es la implementación privada - toda la lista expira junta
type listCache[T any] struct {
	items     []T
	mutex     sync.RWMutex
	createdAt time.Time
	ttl       time.Duration
}

func newListCache[T any]() *listCache[T] {
	return &listCache[T]{
		items:     make([]T, 0),
		createdAt: time.Now(),
		ttl:       0, // Sin expiración por defecto
	}
}

// SetTTL - Establecer tiempo de vida para toda la lista (reinicia su edad)
func (c *listCache[T]) SetTTL(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resetIfExpired()
	c.ttl = duration
	c.createdAt = time.Now()
}

// isExpired - Verificar si toda la lista ha expirado
func (c *listCache[T]) isExpired() bool {
	if c.ttl <= 0 {
		return false // Sin expiración
	}
	return time.Since(c.createdAt) > c.ttl
}

// resetIfExpired drops expired items; callers must hold the write lock.
func (c *listCache[T]) resetIfExpired() {
	if c.isExpired() {
		clear(c.items)
		c.items = c.items[:0]
		c.createdAt = time.Now()
	}
}

// Add - Agregar un elemento al final de la lista
func (c *listCache[T]) Add(value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resetIfExpired()
	c.items = append(c.items, value)
}

// AddFirst - Agregar un elemento al inicio de la lista
func (c *listCache[T]) AddFirst(value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resetIfExpired()
	c.items = append([]T{value}, c.items...)
}

// Get - Obtener elemento por índice
func (c *listCache[T]) Get(index int) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isExpired() || index < 0 || index >= len(c.items) {
		var zero T
		return zero, false
	}

	return c.items[index], true
}

// GetFirst - Obtener primer elemento
func (c *listCache[T]) GetFirst() (T, bool) {
	return c.Get(0)
}

// GetLast - Obtener último elemento
func (c *listCache[T]) GetLast() (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isExpired() || len(c.items) == 0 {
		var zero T
		return zero, false
	}

	return c.items[len(c.items)-1], true
}

// GetAll - Obtener todos los elementos (copia)
func (c *listCache[T]) GetAll() []T {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isExpired() {
		return []T{}
	}

	// Crear copia
	result := make([]T, len(c.items))
	copy(result, c.items)
	return result
}

// GetAllWithIndex - Obtener todos los elementos con sus índices
func (c *listCache[T]) GetAllWithIndex() map[int]T {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[int]T)
	if c.isExpired() {
		return result
	}

	for i, item := range c.items {
		result[i] = item
	}
	return result
}

// Range - Iterar sobre todos los elementos (sobre una copia, fn puede modificar la lista)
func (c *listCache[T]) Range(fn func(index int, value T) bool) {
	for i, item := range c.GetAll() {
		if !fn(i, item) {
			break
		}
	}
}

// RemoveAt - Remover elemento por índice
func (c *listCache[T]) RemoveAt(index int) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isExpired() || index < 0 || index >= len(c.items) {
		return false
	}

	c.items = append(c.items[:index], c.items[index+1:]...)
	return true
}

// RemoveFirst - Remover primer elemento
func (c *listCache[T]) RemoveFirst() bool {
	return c.RemoveAt(0)
}

// RemoveLast - Remover último elemento
func (c *listCache[T]) RemoveLast() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isExpired() || len(c.items) == 0 {
		return false
	}

	c.items = c.items[:len(c.items)-1]
	return true
}

// Clear - Limpiar toda la lista manualmente
func (c *listCache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	clear(c.items)
	c.items = c.items[:0]
	c.createdAt = time.Now()
}

// Len - Obtener número de elementos
func (c *listCache[T]) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isExpired() {
		return 0
	}

	return len(c.items)
}

// IsEmpty - Verificar si la lista está vacía
func (c *listCache[T]) IsEmpty() bool {
	return c.Len() == 0
}

// IsExpired - Método público para verificar si la lista ha expirado
func (c *listCache[T]) IsExpired() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isExpired()
}

// GetTTL - Obtener el TTL actual
func (c *listCache[T]) GetTTL() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ttl
}

// GetAge - Obtener la edad de la lista
func (c *listCache[T]) GetAge() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.createdAt)
}

// GetRemainingTTL - Obtener tiempo restante antes de expiración
func (c *listCache[T]) GetRemainingTTL() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.ttl <= 0 {
		return -1 // Sin expiración
	}

	remaining := c.ttl - time.Since(c.createdAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ListCache - Interfaz pública para cache de lista (simplificada)
type ListCache[T any] interface {
	// Métodos de escritura
	Add(value T)
	AddFirst(value T)

	// Métodos de lectura
	Get(index int) (T, bool)
	GetFirst() (T, bool)
	GetLast() (T, bool)
	GetAll() []T
	GetAllWithIndex() map[int]T
	Range(fn func(index int, value T) bool)

	// Métodos de eliminación
	RemoveAt(index int) bool
	RemoveFirst() bool
	RemoveLast() bool
	Clear() // Control manual total

	// Métodos de información
	Len() int
	IsEmpty() bool
	IsExpired() bool

	// Control de TTL para toda la lista
	SetTTL(duration time.Duration)
	GetTTL() time.Duration
	GetAge() time.Duration
	GetRemainingTTL() time.Duration
}

// NewListCache - Función pública para crear una nueva instancia de cache de lista
func NewListCache[T any]() ListCache[T] {
	return newListCache[T]()
}

// Generic list cache instance management - privado
var (
	listCacheInstances = make(map[string]any)
	listCacheMutex     sync.Mutex
)

// GetListCacheInstance - Función pública para obtener una instancia singleton de cache de lista.
// Instances are keyed by cacheType and T.
func GetListCacheInstance[T any](cacheType string) ListCache[T] {
	key := registryKey[T](cacheType)
	listCacheMutex.Lock()
	defer listCacheMutex.Unlock()
	if instance, ok := listCacheInstances[key]; ok {
		return instance.(*listCache[T])
	}
	newCache := newListCache[T]()
	listCacheInstances[key] = newCache
	return newCache
}
