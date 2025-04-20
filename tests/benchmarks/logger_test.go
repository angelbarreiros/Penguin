package benchmarks

import (
	"angelotero/commonBackend/logger"
	"testing"

	"github.com/fortytw2/leaktest"
)

func BenchmarkLoggerColdStart(b *testing.B) {
	b.StopTimer()
	defer leaktest.Check(b)()

	b.StartTimer()

	logger.DefaultLogger()

}
