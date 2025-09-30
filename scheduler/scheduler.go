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
	nextRunTime time.Time
}
type Scheduler struct {
	chanel     chan job
	ticker     *time.Ticker
	stateMux   sync.RWMutex // Solo para estado del scheduler
	jobs       *sync.Map    // Cambio a sync.Map
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
	s.jobs.Store(job.id, &job)
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
	s.jobs.Store(intervalJob.id, &intervalJob)
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
	if jobInterface, exists := s.jobs.Load(id); exists {
		if job := jobInterface.(*job); job.jobFunction != nil {
			return job.jobFunction.returnChannel
		}
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

// Pause stops the Scheduler's ticker and sets its state to not running.
// This effectively pauses the Scheduler, similar to turning it off and on again,
// which may use more memory as resources are reallocated when resumed.
// Note: The Scheduler will sleep if there are no jobs currently running.
func (s *Scheduler) Pause() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()
	if s.isRunning {
		log.Println("Scheduler Paused")
		if s.ticker != nil {
			s.ticker.Stop()
			s.ticker = nil
		}
		s.isRunning = false
		s.isSleeping = false
	}
}
func (s *Scheduler) UnPause() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()

	if !s.isRunning {
		log.Println("Scheduler Unpaused")
		s.isRunning = true
		if s.ticker != nil {
			s.ticker.Reset(time.Second)
		} else {
			s.ticker = time.NewTicker(time.Second)
		}
		s.isSleeping = false
	}
}
func (s *Scheduler) sleep() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()
	if !s.isSleeping {
		log.Println("Scheduler sleeps")
		if s.ticker != nil {
			s.ticker.Stop()
		}
		s.isRunning = false
		s.isSleeping = true
	}
}
func (s *Scheduler) awake() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()
	if s.isSleeping {
		log.Println("Scheduler awakes")
		s.ticker.Reset(time.Second)
		s.isRunning = true
		s.isSleeping = false
	}

}
func initScheduler() {
	schedulerInstance = &Scheduler{
		chanel:     make(chan job),
		jobs:       &sync.Map{},
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
				s.stateMux.Lock()
				s.wg.Done()
				s.stateMux.Unlock()
				return
			}
			s.stateMux.RLock()
			isSleeping := s.isSleeping
			s.stateMux.RUnlock()
			if isSleeping {
				s.awake()
			}
		}
	}
}
func (s *Scheduler) getTickerChannel() <-chan time.Time {
	s.stateMux.RLock()
	defer s.stateMux.RUnlock()
	if s.ticker != nil {
		return s.ticker.C
	}
	return make(<-chan time.Time)
}
func (s *Scheduler) processJobs() {
	s.stateMux.RLock()
	isRunning := s.isRunning
	isSleeping := s.isSleeping
	s.stateMux.RUnlock()

	// Check if we need to sleep (no jobs available)
	jobCount := 0
	s.jobs.Range(func(key, value interface{}) bool {
		jobCount++
		return false // Stop after first job found
	})

	if jobCount == 0 && isRunning && !isSleeping {
		s.sleep()
		return
	}

	// Collect jobs to execute without holding any locks
	var jobsToExecute []jobFunction
	var jobsToDelete []uint64
	var intervalJobsToUpdate []struct {
		id      uint64
		nextRun time.Time
	}

	now := time.Now()
	s.jobs.Range(func(key, value interface{}) bool {
		k := key.(uint64)
		v := value.(*job)

		if v.jobStatus.state == 0 { // paused
			return true // continue range
		}

		if v.mode == 0 && now.After(v.when) { // one-time job ready
			jobsToExecute = append(jobsToExecute, *v.jobFunction)
			jobsToDelete = append(jobsToDelete, k)
		} else if v.mode == 1 && now.After(v.nextRun) { // interval job ready
			jobsToExecute = append(jobsToExecute, *v.jobFunction)
			intervalJobsToUpdate = append(intervalJobsToUpdate, struct {
				id      uint64
				nextRun time.Time
			}{k, now.Add(v.interval)})
		}
		return true // continue range
	})

	// Execute jobs without holding any locks
	for _, jobFunc := range jobsToExecute {
		go jobFunc.executeJob()
	}

	// Update interval jobs next run time
	for _, update := range intervalJobsToUpdate {
		if jobInterface, exists := s.jobs.Load(update.id); exists {
			if job := jobInterface.(*job); job != nil {
				job.nextRun = update.nextRun
			}
		}
	}

	// Delete completed one-time jobs
	for _, k := range jobsToDelete {
		s.jobs.Delete(k)
	}
}
func (s *Scheduler) removeById(id uint64) error {
	if _, exists := s.jobs.LoadAndDelete(id); exists {
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
	if jobInterface, exists := s.jobs.Load(id); exists {
		if job := jobInterface.(*job); job != nil {
			job.jobStatus.state = 0
			return nil
		}
	}
	return fmt.Errorf("job with id %d not found", id)
}
func (s *Scheduler) unPauseById(id uint64) error {
	if jobInterface, exists := s.jobs.Load(id); exists {
		if job := jobInterface.(*job); job != nil {
			job.jobStatus.state = 1
			return nil
		}
	}
	return fmt.Errorf("job with id %d not found", id)
}
