package scheduler

import (
	"container/heap"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/angelbarreiros/Penguin/logger"
)

type jobMode uint8

const (
	jobModeOneTime jobMode = iota
	jobModeInterval
	jobModeEveryXDaysAt
)

type jobStatus struct {
	state       int16 // "running" : 1 , "paused" 0: "failed" : -1
	nextRunTime time.Time
}

type Scheduler struct {
	stateMux sync.RWMutex

	jobs  map[uint64]*job
	queue jobPriorityQueue
	chMap sync.Map

	commandCh chan schedulerCommand
	done      chan struct{}
	wg        sync.WaitGroup

	isRunning  bool
	isSleeping bool
	stopped    bool
}

type JobFuncInterface interface {
	Execute() []any
}

type jobFunction struct {
	function      JobFuncInterface
	returnChannel chan []any
	closeOnce     sync.Once
}

type job struct {
	id           uint64
	mode         jobMode
	when         time.Time
	interval     time.Duration
	daysInterval int
	timeOfDay    time.Duration
	nextRun      time.Time
	heapIndex    int
	jobFunction  *jobFunction
	jobStatus    jobStatus
}

type schedulerCommandType uint8

const (
	cmdAddJob schedulerCommandType = iota
	cmdRemoveJob
	cmdPauseJob
	cmdUnpauseJob
	cmdPauseScheduler
	cmdUnpauseScheduler
	cmdStopScheduler
)

type schedulerCommand struct {
	typ schedulerCommandType
	id  uint64
	job *job

	respErr chan error
	respAck chan struct{}
}

type jobPriorityQueue []*job

func (q jobPriorityQueue) Len() int { return len(q) }

func (q jobPriorityQueue) Less(i, j int) bool {
	return q[i].nextRun.Before(q[j].nextRun)
}

func (q jobPriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].heapIndex = i
	q[j].heapIndex = j
}

func (q *jobPriorityQueue) Push(x any) {
	item := x.(*job)
	item.heapIndex = len(*q)
	*q = append(*q, item)
}

func (q *jobPriorityQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	item.heapIndex = -1
	*q = old[:n-1]
	return item
}

var schedulerInstance *Scheduler
var schedulerInstanceMux sync.Mutex
var jobCounter uint64

func StartScheduler() *Scheduler {
	schedulerInstanceMux.Lock()
	defer schedulerInstanceMux.Unlock()

	if schedulerInstance == nil || schedulerInstance.isStopped() {
		schedulerInstance = newScheduler()
	}

	return schedulerInstance
}

func JobFunction(f JobFuncInterface) *jobFunction {
	return &jobFunction{function: f}
}

func (j *jobFunction) WithReturnChannel() (*jobFunction, chan []any) {
	returnChannel := make(chan []any, 1)
	j.returnChannel = returnChannel
	j.closeOnce = sync.Once{}
	return j, j.returnChannel
}

func (s *Scheduler) ScheduleProgrammedOneTimeJob(when time.Time, jobFunction *jobFunction) (uint64, error) {
	if !when.After(time.Now()) {
		return 0, fmt.Errorf("job time must be in the future")
	}
	if jobFunction == nil || jobFunction.function == nil {
		return 0, fmt.Errorf("job function cannot be nil")
	}

	id := atomic.AddUint64(&jobCounter, 1)
	newJob := &job{
		id:          id,
		mode:        jobModeOneTime,
		when:        when,
		nextRun:     when,
		heapIndex:   -1,
		jobFunction: jobFunction,
		jobStatus:   jobStatus{state: 1, nextRunTime: when},
	}

	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdAddJob, job: newJob, respErr: resp}); err != nil {
		return 0, err
	}
	if err := <-resp; err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Scheduler) ScheduleOneTimeJob(delay time.Duration, jobFunction *jobFunction) (uint64, error) {
	if delay <= 0 {
		return 0, fmt.Errorf("delay must be greater than zero")
	}
	return s.ScheduleProgrammedOneTimeJob(time.Now().Add(delay), jobFunction)
}

func (s *Scheduler) ScheduleIntervalJob(interval time.Duration, jobFunction *jobFunction) (uint64, error) {
	if interval <= 0 {
		return 0, fmt.Errorf("interval must be greater than zero")
	}
	if jobFunction == nil || jobFunction.function == nil {
		return 0, fmt.Errorf("job function cannot be nil")
	}

	id := atomic.AddUint64(&jobCounter, 1)
	nextRun := time.Now().Add(interval)
	intervalJob := &job{
		id:          id,
		mode:        jobModeInterval,
		interval:    interval,
		nextRun:     nextRun,
		heapIndex:   -1,
		jobFunction: jobFunction,
		jobStatus:   jobStatus{state: 1, nextRunTime: nextRun},
	}

	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdAddJob, job: intervalJob, respErr: resp}); err != nil {
		return 0, err
	}
	if err := <-resp; err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Scheduler) ScheduleProgrammedIntervalJob(everyXDays int, at time.Time, jobFunction *jobFunction) (uint64, error) {
	if everyXDays <= 0 {
		return 0, fmt.Errorf("days must be greater than zero")
	}
	if jobFunction == nil || jobFunction.function == nil {
		return 0, fmt.Errorf("job function cannot be nil")
	}

	id := atomic.AddUint64(&jobCounter, 1)
	hour, minute, second := at.Clock()
	timeOfDay := time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second + time.Duration(at.Nanosecond())
	nextRun := nextRunEveryXDaysAt(time.Now(), everyXDays, timeOfDay)

	periodicAtJob := &job{
		id:           id,
		mode:         jobModeEveryXDaysAt,
		daysInterval: everyXDays,
		timeOfDay:    timeOfDay,
		nextRun:      nextRun,
		heapIndex:    -1,
		jobFunction:  jobFunction,
		jobStatus:    jobStatus{state: 1, nextRunTime: nextRun},
	}

	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdAddJob, job: periodicAtJob, respErr: resp}); err != nil {
		return 0, err
	}
	if err := <-resp; err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Scheduler) RemoveJob(id uint64) error {
	if id == 0 {
		return fmt.Errorf("job id cannot be empty")
	}

	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdRemoveJob, id: id, respErr: resp}); err != nil {
		return err
	}

	return <-resp
}

func (s *Scheduler) PauseJob(id uint64) error {
	if id == 0 {
		return fmt.Errorf("job id cannot be empty")
	}

	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdPauseJob, id: id, respErr: resp}); err != nil {
		return err
	}

	return <-resp
}

func (s *Scheduler) UnPauseJob(id uint64) error {
	if id == 0 {
		return fmt.Errorf("job id cannot be empty")
	}

	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdUnpauseJob, id: id, respErr: resp}); err != nil {
		return err
	}

	return <-resp
}

func (s *Scheduler) GetChannel(id uint64) chan []any {
	if ch, exists := s.chMap.Load(id); exists {
		if typed, ok := ch.(chan []any); ok {
			return typed
		}
	}

	return nil
}

func (s *Scheduler) Stop() {
	ack := make(chan struct{}, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdStopScheduler, respAck: ack}); err != nil {
		return
	}
	<-ack
	s.wg.Wait()

	schedulerInstanceMux.Lock()
	if schedulerInstance == s {
		schedulerInstance = nil
	}
	schedulerInstanceMux.Unlock()

	logger.GetConsoleLogger().Info("Scheduler stopped")
}

func (s *Scheduler) IsRunning() bool {
	s.stateMux.RLock()
	defer s.stateMux.RUnlock()
	return s.isRunning
}

func (s *Scheduler) Pause() {
	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdPauseScheduler, respErr: resp}); err != nil {
		return
	}
	_ = <-resp
}

func (s *Scheduler) UnPause() {
	resp := make(chan error, 1)
	if err := s.submitCommand(schedulerCommand{typ: cmdUnpauseScheduler, respErr: resp}); err != nil {
		return
	}
	_ = <-resp
}

func (s *Scheduler) sleepLocked() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()
	if !s.isSleeping {
		logger.GetConsoleLogger().Info("Scheduler Sleeps (no jobs)")
		s.isSleeping = true
	}
}

func (s *Scheduler) awakeLocked() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()
	if s.isSleeping {
		logger.GetConsoleLogger().Info("Scheduler awakes")
		s.isSleeping = false
	}
}

func newScheduler() *Scheduler {
	s := &Scheduler{
		jobs:       make(map[uint64]*job),
		queue:      make(jobPriorityQueue, 0),
		commandCh:  make(chan schedulerCommand),
		done:       make(chan struct{}),
		isRunning:  true,
		isSleeping: true,
		stopped:    false,
	}
	heap.Init(&s.queue)

	logger.GetConsoleLogger().Info("Scheduler initialized")
	s.wg.Add(1)
	go s.schedulerLoop()

	return s
}

func (s *Scheduler) submitCommand(cmd schedulerCommand) error {
	s.stateMux.RLock()
	stopped := s.stopped
	s.stateMux.RUnlock()
	if stopped {
		return fmt.Errorf("scheduler stopped")
	}

	select {
	case s.commandCh <- cmd:
		return nil
	case <-s.done:
		return fmt.Errorf("scheduler stopped")
	}
}

func (s *Scheduler) isStopped() bool {
	s.stateMux.RLock()
	defer s.stateMux.RUnlock()
	return s.stopped
}

func (s *Scheduler) schedulerLoop() {
	defer s.wg.Done()

	var timer *time.Timer
	for {
		s.stateMux.RLock()
		running := s.isRunning
		stopped := s.stopped
		s.stateMux.RUnlock()

		if stopped {
			if timer != nil {
				stopAndDrainTimer(timer)
			}
			return
		}

		if !running || len(s.queue) == 0 {
			if len(s.queue) == 0 {
				s.sleepLocked()
			}

			cmd := <-s.commandCh
			if s.handleCommand(cmd) {
				if timer != nil {
					stopAndDrainTimer(timer)
				}
				return
			}
			continue
		}

		s.awakeLocked()
		next := s.queue[0].nextRun
		wait := max(time.Until(next), 0)

		if timer == nil {
			timer = time.NewTimer(wait)
		} else {
			stopAndDrainTimer(timer)
			timer.Reset(wait)
		}

		select {
		case <-timer.C:
			s.executeDueJobs()
		case cmd := <-s.commandCh:
			if s.handleCommand(cmd) {
				if timer != nil {
					stopAndDrainTimer(timer)
				}
				return
			}
		}
	}
}

func (s *Scheduler) handleCommand(cmd schedulerCommand) bool {
	switch cmd.typ {
	case cmdAddJob:
		s.addJob(cmd.job)
		if cmd.respErr != nil {
			cmd.respErr <- nil
		}

	case cmdRemoveJob:
		cmd.respErr <- s.removeById(cmd.id)

	case cmdPauseJob:
		cmd.respErr <- s.pauseById(cmd.id)

	case cmdUnpauseJob:
		cmd.respErr <- s.unPauseById(cmd.id)

	case cmdPauseScheduler:
		s.stateMux.Lock()
		if s.isRunning {
			logger.GetConsoleLogger().Info("Scheduler Paused")
			s.isRunning = false
		}
		s.stateMux.Unlock()
		if cmd.respErr != nil {
			cmd.respErr <- nil
		}

	case cmdUnpauseScheduler:
		s.stateMux.Lock()
		if !s.isRunning && !s.stopped {
			logger.GetConsoleLogger().Info("Scheduler Unpaused")
			s.isRunning = true
		}
		s.stateMux.Unlock()
		if cmd.respErr != nil {
			cmd.respErr <- nil
		}

	case cmdStopScheduler:
		s.stateMux.Lock()
		s.stopped = true
		s.isRunning = false
		s.stateMux.Unlock()

		for _, j := range s.jobs {
			if j.jobFunction != nil {
				j.jobFunction.closeReturnChannel()
			}
			s.chMap.Delete(j.id)
		}
		s.jobs = make(map[uint64]*job)
		s.queue = s.queue[:0]

		select {
		case <-s.done:
		default:
			close(s.done)
		}

		if cmd.respAck != nil {
			cmd.respAck <- struct{}{}
		}

		return true
	}

	return false
}

func (s *Scheduler) addJob(j *job) {
	s.jobs[j.id] = j
	if j.jobFunction != nil && j.jobFunction.returnChannel != nil {
		s.chMap.Store(j.id, j.jobFunction.returnChannel)
	}
	if j.jobStatus.state != 0 {
		heap.Push(&s.queue, j)
	}
}

func (s *Scheduler) executeDueJobs() {
	now := time.Now()

	for len(s.queue) > 0 {
		nextJob := s.queue[0]
		if nextJob.nextRun.After(now) {
			break
		}

		heap.Pop(&s.queue)

		storedJob, exists := s.jobs[nextJob.id]
		if !exists || storedJob.jobStatus.state == 0 {
			continue
		}

		closeChannel := storedJob.mode == jobModeOneTime
		go storedJob.jobFunction.executeJob(closeChannel)

		switch storedJob.mode {
		case jobModeOneTime:
			s.chMap.Delete(storedJob.id)
			delete(s.jobs, storedJob.id)

		case jobModeInterval:
			storedJob.nextRun = now.Add(storedJob.interval)
			storedJob.jobStatus.nextRunTime = storedJob.nextRun
			heap.Push(&s.queue, storedJob)

		case jobModeEveryXDaysAt:
			storedJob.nextRun = nextRunAfterXDays(now, storedJob.daysInterval, storedJob.timeOfDay)
			storedJob.jobStatus.nextRunTime = storedJob.nextRun
			heap.Push(&s.queue, storedJob)
		}
	}
}

func (s *Scheduler) removeById(id uint64) error {
	j, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job with id %d not found", id)
	}

	if j.heapIndex >= 0 && j.heapIndex < len(s.queue) {
		heap.Remove(&s.queue, j.heapIndex)
	}

	delete(s.jobs, id)
	s.chMap.Delete(id)
	if j.jobFunction != nil {
		j.jobFunction.closeReturnChannel()
	}
	return nil
}

func (j *jobFunction) executeJob(closeAfter bool) {
	results := j.function.Execute()

	if j.returnChannel != nil {
		if len(results) > 0 {
			select {
			case j.returnChannel <- results:
			default:
			}
		}

		if closeAfter {
			j.closeReturnChannel()
		}
	}
}

func (j *jobFunction) closeReturnChannel() {
	if j.returnChannel == nil {
		return
	}

	j.closeOnce.Do(func() {
		close(j.returnChannel)
	})
}

func (s *Scheduler) pauseById(id uint64) error {
	j, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job with id %d not found", id)
	}

	if j.jobStatus.state == 0 {
		return nil
	}

	j.jobStatus.state = 0
	if j.heapIndex >= 0 && j.heapIndex < len(s.queue) {
		heap.Remove(&s.queue, j.heapIndex)
	}

	return nil
}

func (s *Scheduler) unPauseById(id uint64) error {
	j, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job with id %d not found", id)
	}

	if j.jobStatus.state != 0 {
		return nil
	}

	j.jobStatus.state = 1
	if j.nextRun.Before(time.Now()) {
		j.nextRun = time.Now()
	}
	j.jobStatus.nextRunTime = j.nextRun
	heap.Push(&s.queue, j)

	return nil
}

func nextRunEveryXDaysAt(from time.Time, days int, timeOfDay time.Duration) time.Time {
	dayStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	candidate := dayStart.Add(timeOfDay)
	if !candidate.After(from) {
		candidate = candidate.AddDate(0, 0, days)
	}
	return candidate
}

func nextRunAfterXDays(from time.Time, days int, timeOfDay time.Duration) time.Time {
	next := nextRunEveryXDaysAt(from, days, timeOfDay)
	for !next.After(from) {
		next = next.AddDate(0, 0, days)
	}
	return next
}

func stopAndDrainTimer(t *time.Timer) {
	if t == nil {
		return
	}

	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
