package cronengine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/google/uuid"
)

type Scheduler struct {
	chanel chan job
	ticker *time.Ticker
	mux    sync.Mutex
	jobs   []job
}
type JobFunction func(f func(args ...any) []any, args ...any) []any
type job struct {
	id       string
	mode     uint // 0 = one time, 1 = interval
	when     time.Time
	interval time.Duration
	nextRun  time.Time
	_        *context.Context
}

var schedulerInstance *Scheduler = nil
var once sync.Once

func StartScheduler() *Scheduler {
	once.Do(initScheduler)
	return schedulerInstance

}
func (s *Scheduler) ScheduleJob(name string, when time.Time) (string, error) {
	if len(name) > 40 {
		return "", fmt.Errorf("job name cannot exceed 40 characters")
	}
	if when.Before(time.Now()) {
		return "", fmt.Errorf("job time cannot be in the past")
	}
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString("-")
	sb.WriteString(uuid.NewString())

	var job job = job{
		mode: 0,
		id:   sb.String(),
		when: when,
	}
	s.mux.Lock()

	log.Printf("Job %s added to scheduler\n", job.id)
	s.jobs = append(s.jobs, job)
	s.mux.Unlock()
	s.chanel <- job

	return sb.String(), nil
}
func (s *Scheduler) ScheduleIntervalJob(name string, interval time.Duration) (string, error) {
	if len(name) > 40 {
		return "", fmt.Errorf("job name cannot exceed 40 characters")
	}
	if interval <= 0 {
		return "", fmt.Errorf("interval must be greater than zero")
	}
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString("-")
	sb.WriteString(uuid.NewString())

	var intervalJob job = job{
		mode:     1,
		id:       sb.String(),
		interval: interval,
		nextRun:  time.Now().Add(interval),
	}
	s.mux.Lock()
	log.Printf("Job %s added to scheduler\n", intervalJob.id)
	s.jobs = append(s.jobs, intervalJob)
	s.mux.Unlock()
	s.chanel <- intervalJob
	return sb.String(), nil
}
func (s *Scheduler) RemoveJob(id string) error {
	if id == "" {
		return fmt.Errorf("job id cannot be empty")

	}
	if err := removeById(id); err != nil {
		return err
	}
	log.Printf("Job %s removed from scheduler\n", id)
	return nil
}

func initScheduler() {
	schedulerInstance = &Scheduler{
		chanel: make(chan job),
		jobs:   make([]job, 0),
		ticker: nil,
	}
	go schedulerInstance.schedulerLoop()

}

func (s *Scheduler) schedulerLoop() {
	for {
		select {
		case job := <-s.chanel:
			s.mux.Lock()
			log.Printf("Processing job from channel: %s", job.id)
			s.awakeTicker()
			s.mux.Unlock()

		case <-s.getTickerChannel():
			s.processJobs()

		}
	}
}

func (s *Scheduler) getTickerChannel() <-chan time.Time {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.ticker != nil {
		return s.ticker.C
	}

	return nil
}
func (s *Scheduler) processJobs() {
	if len(s.jobs) == 0 {
		log.Println("No jobs scheduled - Scheduler goes to sleep")
		s.mux.Lock()
		s.ticker.Stop()
		s.ticker = nil
		s.mux.Unlock()
		return
	}
	for i := len(s.jobs) - 1; i >= 0; i-- {
		if s.jobs[i].mode == 0 && time.Now().After(s.jobs[i].when) {
			log.Printf("Job %s executed", s.jobs[i].id)
			s.mux.Lock()
			s.jobs = slices.Delete(s.jobs, i, i+1)
			s.mux.Unlock()
		} else if s.jobs[i].mode == 1 && time.Now().After(s.jobs[i].nextRun) {
			log.Printf("Interval Job %s executed", s.jobs[i].id)
			s.jobs[i].nextRun = time.Now().Add(s.jobs[i].interval)
		}
	}
}
func (s *Scheduler) awakeTicker() {
	if s.ticker == nil {
		s.ticker = time.NewTicker(1 * time.Second)
		log.Println("Scheduler started")

	}
}
func removeById(id string) error {
	schedulerInstance.mux.Lock()
	for i := len(schedulerInstance.jobs) - 1; i >= 0; i-- {
		if schedulerInstance.jobs[i].id == id {
			schedulerInstance.jobs = slices.Delete(schedulerInstance.jobs, i, i+1)
			return nil
		}
	}
	schedulerInstance.mux.Unlock()
	return fmt.Errorf("job with id %s not found", id)

}
