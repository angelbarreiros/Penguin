package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/angelbarreiros/Penguin/router"
	"github.com/angelbarreiros/Penguin/router/helpers"
	"github.com/angelbarreiros/Penguin/router/types"
	"github.com/angelbarreiros/Penguin/scheduler"
)

func TestNewDateRangeRejectsInvertedRange(t *testing.T) {
	dr, err := types.NewDateRange(time.Now(), time.Now().Add(-time.Hour))
	if dr != nil || !errors.Is(err, types.ErrInvalidDateRange) {
		t.Fatalf("got (%v, %v)", dr, err)
	}
}

func TestDateRangeSplitNonPositive(t *testing.T) {
	dr := types.DateRange{Start: time.Now(), End: time.Now().Add(48 * time.Hour)}
	if got := dr.Split(0); got != nil {
		t.Fatalf("Split(0) = %v", got)
	}
	if got := dr.Split(-1); got != nil {
		t.Fatalf("Split(-1) = %v", got)
	}
}

func TestDateRangeDaysAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Skip("tzdata unavailable")
	}
	dr := types.DateRange{
		Start: time.Date(2026, 3, 29, 0, 0, 0, 0, loc), // 23-hour day
		End:   time.Date(2026, 3, 30, 0, 0, 0, 0, loc),
	}
	if dr.Days() != 1 {
		t.Fatalf("Days = %d, want 1", dr.Days())
	}
}

func TestPostgreSQLTypesJSONRoundTrip(t *testing.T) {
	ts := types.PostgreSQLTimeTimeStamp{Year: 2026, Month: 1, Day: 2, Hour: 3, Minute: 4, Second: 5, Microsecond: 6, Timezone: "+02:00"}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"2026-01-02 03:04:05.000006+02:00"` {
		t.Fatalf("marshal = %s", data)
	}
	var back types.PostgreSQLTimeTimeStamp
	if err := json.Unmarshal(data, &back); err != nil || back != ts {
		t.Fatalf("round trip = %+v, %v", back, err)
	}
	if err := back.Unmarshal([]byte("2026-01-02 03:04:05.5+02")); err != nil || back.Microsecond != 500000 {
		t.Fatalf("short fraction parse = %+v, %v", back, err)
	}

	date := types.PostgreSQLDate{Year: 2026, Month: 10, Day: 27}
	data, _ = json.Marshal(date)
	if string(data) != `"2026-10-27"` {
		t.Fatalf("date marshal = %s", data)
	}
	var unquoted types.PostgreSQLTime
	if err := json.Unmarshal([]byte(`2026`), &unquoted); err == nil {
		t.Fatal("unquoted time accepted")
	}
}

func TestNumericParamValidation(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?big=18446744073709551615&hex=0x1p3&neg=-1.0&abc=abc1&n=-5&lead=0abc", nil)
	if v, err := helpers.GetNullUint64QueryParam("big", r); err == nil {
		t.Errorf("uint64 overflow accepted: %v", v)
	}
	if v, err := helpers.GetNullFloat64QueryParam("hex", r); err == nil {
		t.Errorf("hex float accepted: %v", v)
	}
	if v, err := helpers.GetNullFloat64QueryParam("neg", r); err != nil || v.Float64 != -1 {
		t.Errorf("-1.0 = %v, %v", v, err)
	}
	if _, err := helpers.GetNullUint64QueryParam("abc", r); err == nil {
		t.Error("abc1 accepted")
	}
	if _, err := helpers.GetNullUint64QueryParam("lead", r); err == nil {
		t.Error("0abc accepted")
	}
	if v, err := helpers.GetNullInt64QueryParam("n", r); err != nil || v.Int64 != -5 {
		t.Errorf("-5 = %v, %v", v, err)
	}
}

func TestDeserializeBody(t *testing.T) {
	var v map[string]any
	err := helpers.DeserializeBodyWithLimit(httptest.NewRequest(http.MethodPost, "/", strings.NewReader("")), &v, 100)
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Errorf("empty body err = %v", err)
	}
	err = helpers.DeserializeBodyWithLimit(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1}{"b":2}`)), &v, 100)
	if err == nil {
		t.Error("trailing data accepted")
	}
	err = helpers.DeserializeBodyWithLimit(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"`+strings.Repeat("x", 200)+`"}`)), &v, 100)
	if err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Errorf("too large err = %v", err)
	}
}

func TestSendSuccessResponseMarshalError(t *testing.T) {
	w := httptest.NewRecorder()
	helpers.SendSuccessResponse(w, map[string]any{"bad": make(chan int)})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestPaginationBounds(t *testing.T) {
	p := helpers.GetPaginationParams(httptest.NewRequest(http.MethodGet, "/?page=0&pageSize=10000000000", nil), 20)
	if p.Page != 1 || p.PageSize != 20 {
		t.Fatalf("got %+v", p)
	}
	p = helpers.GetPaginationParams(httptest.NewRequest(http.MethodGet, "/?pageSize=5000", nil), 20)
	if p.PageSize != helpers.DefaultMaxPageSize {
		t.Fatalf("pageSize = %d, want cap %d", p.PageSize, helpers.DefaultMaxPageSize)
	}
}

func TestStringCacheRefreshIsNotEvictedByOldTimer(t *testing.T) {
	c := helpers.NewStringCache[string]()
	defer c.Close()
	c.Store("k", "v1", 100*time.Millisecond)
	c.Store("k", "v2", 10*time.Second)
	time.Sleep(300 * time.Millisecond)
	if v, ok := c.Load("k"); !ok || v != "v2" {
		t.Fatalf("Load = %q, %v", v, ok)
	}
}

func TestStringCacheExpiresAndClosingOneCacheKeepsOthersWorking(t *testing.T) {
	a := helpers.NewStringCache[int]()
	b := helpers.NewStringCache[int]()
	defer b.Close()
	a.Close()

	b.Store("k", 1, 50*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	if b.Has("k") || b.Len() != 0 {
		t.Fatal("entry did not expire")
	}
	b.Store("forever", 1, 0)
	if !b.Has("forever") {
		t.Fatal("zero TTL entry missing")
	}
}

func TestCacheRegistrySameNameDifferentType(t *testing.T) {
	ints := helpers.GetStringCacheInstance[int]("shared-name")
	strs := helpers.GetStringCacheInstance[string]("shared-name")
	ints.Store("k", 1, 0)
	if strs.Has("k") {
		t.Fatal("caches with different types share storage")
	}
}

func TestListCacheSetTTLDoesNotWipeLaterItems(t *testing.T) {
	l := helpers.NewListCache[int]()
	l.SetTTL(100 * time.Millisecond)
	l.SetTTL(time.Hour)
	l.Add(1)
	time.Sleep(200 * time.Millisecond)
	if l.Len() != 1 {
		t.Fatalf("Len = %d, want 1", l.Len())
	}
}

func TestGenerateCacheKey(t *testing.T) {
	a := httptest.NewRequest(http.MethodGet, "/p?a=1&a=2", nil)
	b := httptest.NewRequest(http.MethodGet, "/p?a=1", nil)
	if helpers.GenerateCacheKey(a) == helpers.GenerateCacheKey(b) {
		t.Fatal("multi-value query collides with single value")
	}
	b.Header.Set("Authorization", "Bearer secret-token")
	if strings.Contains(helpers.GenerateCacheKey(b), "secret-token") {
		t.Fatal("token stored in key")
	}
}

type slowJob struct{ calls *atomic.Int32 }

func (s slowJob) Execute() []any {
	s.calls.Add(1)
	time.Sleep(100 * time.Millisecond)
	return []any{1}
}

func TestSchedulerRemoveWhileRunningDoesNotPanic(t *testing.T) {
	s := scheduler.NewScheduler()
	defer s.Stop()
	var calls atomic.Int32
	jf, ch := scheduler.JobFunction(slowJob{&calls}).WithReturnChannel()
	id, err := s.ScheduleIntervalJob(20*time.Millisecond, jf)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(40 * time.Millisecond)
	if err := s.RemoveJob(id); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	for range ch {
	}
}

type panicJob struct{}

func (panicJob) Execute() []any { panic("boom") }

func TestSchedulerRecoversJobPanic(t *testing.T) {
	s := scheduler.NewScheduler()
	defer s.Stop()
	if _, err := s.ScheduleOneTimeJob(10*time.Millisecond, scheduler.JobFunction(panicJob{})); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if !s.IsRunning() {
		t.Fatal("scheduler stopped after job panic")
	}
}

func TestSchedulerRejectsReusedJobFunction(t *testing.T) {
	s := scheduler.NewScheduler()
	defer s.Stop()
	jf := scheduler.JobFunction(panicJob{})
	if _, err := s.ScheduleOneTimeJob(time.Hour, jf); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ScheduleOneTimeJob(time.Hour, jf); err == nil {
		t.Fatal("reused jobFunction accepted")
	}
}

func TestRouterHeadOptionsAndConcurrentRegistration(t *testing.T) {
	r := router.NewRouter()
	r.NewRoute(router.Route{Path: "/x", Method: router.GET, Handler: func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("body"))
	}})
	h := r.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodHead, "/x", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("HEAD status = %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, "/x", nil))
	if w.Code != http.StatusNoContent || !strings.Contains(w.Header().Get("Allow"), "GET") {
		t.Fatalf("OPTIONS = %d %v", w.Code, w.Header())
	}

	done := make(chan struct{})
	go func() {
		for i := range 200 {
			r.NewRoute(router.Route{Path: "/x", Method: router.HTTPMethod("M" + string(rune('A'+i%26)) + string(rune('A'+i/26))), Handler: func(http.ResponseWriter, *http.Request) {}})
		}
		close(done)
	}()
	for range 200 {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	}
	<-done
}

func TestI18nAcceptLanguageQValues(t *testing.T) {
	i := helpers.GetI18nInstance()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Language", "es;q=0.1, fr-CA;q=0.9, *;q=1")
	if got := i.GetLanguage(r); got != "fr" {
		t.Fatalf("GetLanguage = %q, want fr", got)
	}
	r.Header.Set("Accept-Language", "../")
	if got := i.GetLanguage(r); got != i.DefaultLocale {
		t.Fatalf("GetLanguage = %q, want default", got)
	}
}
