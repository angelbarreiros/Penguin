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

// List-based cache implementation with automatic list expiration

// listCacheCleaner es para limpiar toda la lista cuando expire
type listCacheCleaner[T any] struct {
	cache *listCache[T]
}

func (c *listCacheCleaner[T]) Execute() []any {
	c.cache.Clear()
	return nil
}

// listCache es la implementación privada - toda la lista expira junta
type listCache[T any] struct {
	items     []T
	mutex     sync.RWMutex
	cleaner   *scheduler.Scheduler
	createdAt time.Time
	ttl       time.Duration
}

func newListCache[T any]() *listCache[T] {
	return &listCache[T]{
		items:     make([]T, 0),
		cleaner:   scheduler.StartScheduler(),
		createdAt: time.Now(),
		ttl:       0, // Sin expiración por defecto
	}
}

// SetTTL - Establecer tiempo de vida para toda la lista
func (c *listCache[T]) SetTTL(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ttl = duration
	c.createdAt = time.Now()

	if duration > 0 {
		// Programar limpieza automática
		job := scheduler.JobFunction(&listCacheCleaner[T]{cache: c})
		c.cleaner.ScheduleJob(time.Now().Add(duration), job)
	}
}

// isExpired - Verificar si toda la lista ha expirado
func (c *listCache[T]) isExpired() bool {
	if c.ttl <= 0 {
		return false // Sin expiración
	}
	return time.Since(c.createdAt) > c.ttl
}

// Add - Agregar un elemento al final de la lista
func (c *listCache[T]) Add(value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isExpired() {
		c.items = c.items[:0]    // Limpiar si está expirada
		c.createdAt = time.Now() // Resetear tiempo
	}

	c.items = append(c.items, value)
}

// AddFirst - Agregar un elemento al inicio de la lista
func (c *listCache[T]) AddFirst(value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isExpired() {
		c.items = c.items[:0]    // Limpiar si está expirada
		c.createdAt = time.Now() // Resetear tiempo
	}

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

// Range - Iterar sobre todos los elementos
func (c *listCache[T]) Range(fn func(index int, value T) bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isExpired() {
		return
	}

	for i, item := range c.items {
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
	listCacheMutex     sync.RWMutex
)

// GetListCacheInstance - Función pública para obtener una instancia singleton de cache de lista
func GetListCacheInstance[T any](cacheType string) ListCache[T] {
	listCacheMutex.RLock()
	if instance, ok := listCacheInstances[cacheType]; ok {
		listCacheMutex.RUnlock()
		return instance.(*listCache[T])
	}
	listCacheMutex.RUnlock()

	listCacheMutex.Lock()
	defer listCacheMutex.Unlock()

	// Double check after acquiring write lock
	if instance, ok := listCacheInstances[cacheType]; ok {
		return instance.(*listCache[T])
	}

	newCache := newListCache[T]()
	listCacheInstances[cacheType] = newCache
	return newCache
}
