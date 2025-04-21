package scheduler

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type jobStatus struct {
	state       int16 // "running" : 1 , "paused" 0:  "failed" : -1
	lastRunTime time.Time
	nextRunTime time.Time
	errorCount  int
	lastError   error
}
type Scheduler struct {
	chanel     chan job
	ticker     *time.Ticker
	mux        sync.RWMutex
	jobs       map[uint64]*job
	wg         *sync.WaitGroup
	isRunning  bool
	isSleeping bool
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
	mode        uint16 // 0 = one time, 1 = interval
	when        time.Time
	interval    time.Duration
	nextRun     time.Time
	jobFunction *jobFunction
	jobStatus   jobStatus
}

var schedulerInstance *Scheduler = nil
var once sync.Once
var jobCounter uint64

func StartScheduler() *Scheduler {
	once.Do(initScheduler)
	log.Println("Scheduled started")
	return schedulerInstance
}
func JobFunction(f JobFuncInterface) *jobFunction {
	return &jobFunction{
		function: f,
	}
}
func (j *jobFunction) WithReturnChannel() (*jobFunction, chan []any) {
	returnChannel := make(chan []any, 1)
	j.returnChannel = returnChannel
	return j, j.returnChannel

}
func (s *Scheduler) ScheduleJob(when time.Time, jobFunction *jobFunction) (uint64, error) {
	if when.Before(time.Now()) {
		return 0, fmt.Errorf("job time cannot be in the past")
	}
	if jobFunction == nil {
		return 0, fmt.Errorf("job function cannot be nil")
	}

	var id uint64 = atomic.AddUint64(&jobCounter, 1)

	var job job = job{
		mode:        0,
		id:          id,
		when:        when,
		jobFunction: jobFunction,
		jobStatus: jobStatus{
			state:       1,
			nextRunTime: when,
		},
	}
	s.mux.Lock()
	s.jobs[job.id] = &job
	s.mux.Unlock()
	s.chanel <- job

	return id, nil
}
func (s *Scheduler) ScheduleIntervalJob(interval time.Duration, jobFunction *jobFunction) (uint64, error) {
	if interval <= 0 {
		return 0, fmt.Errorf("interval must be greater than zero")
	}

	var id uint64 = atomic.AddUint64(&jobCounter, 1)
	var intervalJob job = job{
		mode:        1,
		id:          id,
		interval:    interval,
		nextRun:     time.Now().Add(interval),
		jobFunction: jobFunction,
		jobStatus: jobStatus{
			state:       1,
			nextRunTime: time.Now().Add(interval),
		},
	}
	s.mux.Lock()

	s.jobs[intervalJob.id] = &intervalJob
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
func (s *Scheduler) PauseJob(id uint64) error {
	if id == 0 {
		return fmt.Errorf("job id cannot be empty")
	}
	if err := s.pauseById(id); err != nil {
		return err
	}

	return nil
}
func (s *Scheduler) UnPauseJob(id uint64) error {
	if id == 0 {
		return fmt.Errorf("job id cannot be empty")
	}
	if err := s.unPauseById(id); err != nil {
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
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}

	close(s.chanel)
	s.wg.Wait()
	log.Println("Scheduler terminated")
}
func (s *Scheduler) IsRunning() bool {
	return s.isRunning
}

func (s *Scheduler) Pause() {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.isRunning {
		log.Println("Scheduler Paused")
		s.ticker.Stop()
		s.ticker = nil
		s.isRunning = false
		s.isSleeping = false
	}

}
func (s *Scheduler) UnPause() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.isRunning {
		log.Println("Scheduler Unpaused")
		s.isRunning = true
		s.ticker = time.NewTicker(1 * time.Second)
		s.isSleeping = false
	}

}
func (s *Scheduler) sleep() {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.isSleeping {
		log.Println("Scheduler sleeps")
		s.ticker.Stop()
		s.ticker = nil
		s.isRunning = false
		s.isSleeping = true
	}

}
func (s *Scheduler) awake() {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.isSleeping {
		log.Println("Scheduler awakes")
		s.ticker = time.NewTicker(1 * time.Second)
		s.isRunning = true
		s.isSleeping = false
	}

}
func initScheduler() {
	schedulerInstance = &Scheduler{
		chanel:     make(chan job),
		jobs:       make(map[uint64]*job, 10000),
		ticker:     time.NewTicker(1 * time.Second),
		wg:         new(sync.WaitGroup),
		isRunning:  true,
		isSleeping: false,
	}

	schedulerInstance.wg.Add(1)
	go schedulerInstance.schedulerLoop()

}
func (s *Scheduler) schedulerLoop() {
	for {
		select {
		case <-s.getTickerChannel():
			s.processJobs()

		case _, ok := <-s.chanel:
			if !ok {
				s.mux.Lock()
				s.wg.Done()
				s.mux.Unlock()
				return
			}
			s.mux.RLock()
			isSleeping := s.isSleeping
			s.mux.RUnlock()
			if isSleeping {
				s.awake()
			}
		}
	}
}
func (s *Scheduler) getTickerChannel() <-chan time.Time {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.ticker != nil {
		return s.ticker.C
	}
	return make(<-chan time.Time)
}

func (s *Scheduler) processJobs() {
	s.mux.RLock()
	if len(s.jobs) == 0 && s.isRunning && !s.isSleeping {
		s.mux.RUnlock()
		s.sleep()
		return
	}

	var jobsToDelete []uint64
	for k, v := range s.jobs {
		if 0 == v.jobStatus.state {
			continue
		}
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
func (s *Scheduler) pauseById(id uint64) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if job, exists := s.jobs[id]; exists {
		job.jobStatus.state = 0
		return nil
	}
	return fmt.Errorf("job with id %d not found", id)

}
func (s *Scheduler) unPauseById(id uint64) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if job, exists := s.jobs[id]; exists {
		job.jobStatus.state = 1
		return nil
	}
	return fmt.Errorf("job with id %d not found", id)

}
