package cronengine

import (
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

type job struct {
	id       string
	mode     uint // 0 = one time, 1 = interval
	when     time.Time
	interval time.Duration
	nextRun  time.Time
}

var schedulerInstance *Scheduler = nil
var once sync.Once

func StartScheduler() *Scheduler {
	once.Do(initScheduler)
	return schedulerInstance

}
func (s *Scheduler) ScheduleJob(name string, when time.Time) (string, error) {
	schedulerInstance.mux.Lock()
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

	schedulerInstance.chanel <- job

	return sb.String(), nil
}
func (s *Scheduler) ScheduleIntervalJob(name string, interval time.Duration) (string, error) {
	schedulerInstance.mux.Lock()
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

	schedulerInstance.chanel <- intervalJob

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
	go chanelListen()

}
func chanelListen() {

	for job := range schedulerInstance.chanel {

		schedulerInstance.jobs = append(schedulerInstance.jobs, job)
		schedulerInstance.mux.Unlock()
		log.Printf("Job %s added to scheduler\n", job.id)
		if schedulerInstance.ticker == nil {
			schedulerInstance.ticker = time.NewTicker(1 * time.Second)
			go tickerLogic()
			log.Println("Scheduler started")
		}

	}
}
func tickerLogic() {
	for range schedulerInstance.ticker.C {
		schedulerInstance.mux.Lock()
		if len(schedulerInstance.jobs) == 0 {
			log.Println("No jobs scheduled - Scheduler goes to sleep")
			schedulerInstance.ticker.Stop()
			schedulerInstance.ticker = nil
			schedulerInstance.mux.Unlock()
			return

		}
		schedulerInstance.mux.Unlock()
		for i := len(schedulerInstance.jobs) - 1; i >= 0; i-- {
			if schedulerInstance.jobs[i].mode == 0 {
				if time.Now().After(schedulerInstance.jobs[i].when) {
					log.Printf("Job %s executed", schedulerInstance.jobs[i].id)
					schedulerInstance.jobs = slices.Delete(schedulerInstance.jobs, i, i+1)
				}
			} else if schedulerInstance.jobs[i].mode == 1 {
				if time.Now().After(schedulerInstance.jobs[i].nextRun) {
					log.Printf("Interval Job %s executed", schedulerInstance.jobs[i].id)
					schedulerInstance.jobs[i].nextRun = time.Now().Add(schedulerInstance.jobs[i].interval)
				}
			} else {
				log.Fatal("Unknown job mode")

			}

		}

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
