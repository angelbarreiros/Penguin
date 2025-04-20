package cronengine

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Scheduler struct {
	chanel chan job
	ticker *time.Ticker
	mux    sync.RWMutex
	jobs   map[uint64]job
	wg     *sync.WaitGroup
}
type JobFuncInterface interface {
	Execute() []any
}
type jobFunction struct {
	function      JobFuncInterface
	returnChannel chan []any
}
type job struct {
	id          uint64
	mode        uint // 0 = one time, 1 = interval
	when        time.Time
	interval    time.Duration
	nextRun     time.Time
	jobFunction *jobFunction
}

var schedulerInstance *Scheduler = nil
var once sync.Once
var jobCounter uint64

func StartScheduler() *Scheduler {
	once.Do(initScheduler)
	return schedulerInstance
}

func JobFunction(f JobFuncInterface, args ...any) *jobFunction {
	// checkReflectionErr := checkReflection(f, args...)
	// if checkReflectionErr != nil {
	// 	panic(checkReflectionErr)
	// }
	return &jobFunction{
		function: f,
	}
}
func (j *jobFunction) WithReturnChannel() (*jobFunction, chan []any) {
	returnChannel := make(chan []any, 1)
	j.returnChannel = returnChannel
	return j, j.returnChannel

}
func (s *Scheduler) ScheduleJob(when time.Time, todo *jobFunction) (uint64, error) {
	if when.Before(time.Now()) {
		return 0, fmt.Errorf("job time cannot be in the past")
	}
	if todo == nil {
		return 0, fmt.Errorf("job function cannot be nil")
	}

	var id uint64 = atomic.AddUint64(&jobCounter, 1)

	var job job = job{
		mode:        0,
		id:          id,
		when:        when,
		jobFunction: todo,
	}
	s.mux.Lock()
	s.jobs[job.id] = job
	s.mux.Unlock()
	s.chanel <- job

	return id, nil
}
func (s *Scheduler) ScheduleIntervalJob(interval time.Duration, todo *jobFunction) (uint64, error) {
	if interval <= 0 {
		return 0, fmt.Errorf("interval must be greater than zero")
	}

	var id uint64 = atomic.AddUint64(&jobCounter, 1)
	var intervalJob job = job{
		mode:        1,
		id:          id,
		interval:    interval,
		nextRun:     time.Now().Add(interval),
		jobFunction: todo,
	}
	s.mux.Lock()

	s.jobs[intervalJob.id] = intervalJob
	s.mux.Unlock()
	s.chanel <- intervalJob
	return id, nil
}
func (s *Scheduler) RemoveJob(id uint64) error {
	if id == 0 {
		return fmt.Errorf("job id cannot be empty")
	}
	if err := s.removeById(id); err != nil {
		return err
	}

	return nil
}
func (s *Scheduler) GetChannel(id uint64) chan []any {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if job, exists := s.jobs[id]; exists {
		if job.jobFunction == nil {
			return nil
		}
		return job.jobFunction.returnChannel
	}
	return nil
}
func (s *Scheduler) Stop() {
	s.mux.Lock()
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
	s.mux.Unlock()
	close(s.chanel)

	s.wg.Wait()

}
func initScheduler() {
	schedulerInstance = &Scheduler{
		chanel: make(chan job),
		jobs:   make(map[uint64]job, 10000),
		ticker: nil,
		wg:     new(sync.WaitGroup),
	}

	schedulerInstance.wg.Add(1)
	go schedulerInstance.schedulerLoop()

}

func (s *Scheduler) schedulerLoop() {
	defer s.wg.Done()
	for {
		select {
		case <-s.getTickerChannel():
			s.processJobs()

		case _, ok := <-s.chanel:
			if !ok {
				return
			}
			s.awakeTicker()
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
func (s *Scheduler) awakeTicker() {
	s.mux.Lock()
	if s.ticker == nil {
		s.ticker = time.NewTicker(1 * time.Second)

	}
	s.mux.Unlock()
}
func (s *Scheduler) processJobs() {
	if len(s.jobs) == 0 {
		s.mux.Lock()
		s.ticker.Stop()
		s.ticker = nil
		s.jobs = make(map[uint64]job, 10000)
		s.mux.Unlock()
		return
	}
	s.mux.RLock()
	var jobsToDelete []uint64
	for k, v := range s.jobs {
		if v.mode == 0 && time.Now().After(v.when) {
			v.jobFunction.executeJob()
			jobsToDelete = append(jobsToDelete, k)
		} else if v.mode == 1 && time.Now().After(v.nextRun) {
			v.jobFunction.executeJob()
			v.nextRun = time.Now().Add(v.interval)
		}
	}
	s.mux.RUnlock()

	if len(jobsToDelete) > 0 {
		s.mux.Lock()
		for _, k := range jobsToDelete {
			delete(s.jobs, k)
		}
		s.mux.Unlock()
	}

}

func (s *Scheduler) removeById(id uint64) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.jobs[id]; exists {
		delete(s.jobs, id)
		return nil
	}
	return fmt.Errorf("job with id %d not found", id)
}
func (j jobFunction) executeJob() {
	var results = j.function.Execute()
	if j.returnChannel != nil && len(results) > 0 {
		j.returnChannel <- results
		close(j.returnChannel)
	}
}
