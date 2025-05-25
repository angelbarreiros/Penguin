package benchmarks

import (
	"log"
	"testing"
	"time"

	"github.com/angelbarreiros/Penguin/scheduler"
)

type functionSum struct {
	a int
	b int
}

func (f functionSum) Execute() []any {
	var xd = f.a + f.b
	return []any{xd}
}

func BenchmarkCronEngineJobs(b *testing.B) {
	var sch *scheduler.Scheduler = scheduler.StartScheduler()
	var functionSum = functionSum{a: 1, b: 2}
	for b.Loop() {
		var job, ch = scheduler.JobFunction(functionSum).WithReturnChannel()
		var _, err = sch.ScheduleJob(time.Now().Add(1*time.Second), job)
		if err != nil {
			b.Error(err)

		}
		var result = <-ch
		log.Println(result)

	}
	defer sch.Stop()

}
