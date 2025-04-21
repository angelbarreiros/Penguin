package benchmarks

import (
	cronengine "angelotero/commonBackend/scheduler"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
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
	var sch *cronengine.Scheduler = cronengine.StartScheduler()
	var functionSum = functionSum{a: 1, b: 2}
	for b.Loop() {
		var job = cronengine.JobFunction(functionSum)
		var _, err = sch.ScheduleJob(time.Now().Add(1*time.Second), job)

		if err != nil {
			b.Error(err)

		}

	}
	defer sch.Stop()
	defer leaktest.Check(b)()

}
