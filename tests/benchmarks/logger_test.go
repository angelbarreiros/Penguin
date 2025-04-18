package loggertests

import (
	"angelotero/commonBackend/logger"
	"testing"

	"github.com/fortytw2/leaktest"
)

func Init() {}

func BenchmarkLoggerColdStart(b *testing.B) {
	b.StopTimer()
	defer leaktest.Check(b)()

	b.StartTimer()

	logger.DefaultLogger()

}
