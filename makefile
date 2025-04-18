benchmarkDir := ./tests/benchmarks
benchmarkFunctions := .
timestamp := $(shell date +%Y-%m-%d_%H-%M-%S)
outputDir := ./tests/output/benchmarks/
tempDir := ./tests/output
benchmarkOutFile := $(outputDir)$(timestamp)_benchmark_results.csv
tempFile := $(tempDir)/benchmark_raw.tmp
defaultBenchmark := go test -benchmem -bench=. ./tests/benchmarks 
allocsDir := ./services
logDir := ./logs/
mainFile := ./main.go
$(shell mkdir -p $(outputDir) )

benchmark:
	@echo "Running benchmarks..."
	@echo "Saving formatted results to $(benchmarkOutFile)"
	@echo "----------------------------------------"
	
	@bash -c '\
	set -e; \
	trap "echo Cleaning up temp file...; rm -f $(tempFile)" EXIT; \
	go test -benchmem -bench=$(benchmarkFunctions) $(benchmarkDir) > $(tempFile); \
	echo "Benchmark,Iterations,Time/Op,Mem/Op,Allocs/Op" > $(benchmarkOutFile); \
	awk '\''$$1 ~ /^Benchmark/ { printf "%s,%s,%s,%s,%s\n", $$1, $$2, $$3$$4, $$5$$6, $$7 }'\'' $(tempFile) >> $(benchmarkOutFile); \
	'
	@echo "----------------------------------------"
	@echo "Formatted benchmark results saved to $(benchmarkOutFile)"

cleanOutputs:
	@echo "Cleaning up benchmark output files..."
	@rm -f $(outputDir)*
	@echo "Cleaned up benchmark output files."

cleanLogs:
	@echo "Cleaning up log files..."
	@rm -f $(logDir)*
	@echo "Cleaned up log files."

compilerInfo:
	@echo "Allocations Info:"
	@echo "----------------------------------------"
	go build -gcflags="all=-m -l -N" $(allocsDir)
	@echo "----------------------------------------"
	@echo "Allocations info displayed."

run: 
	@echo "Running application..."
	@echo "----------------------------------------"
	go run $(mainFile); 	
	@echo "----------------------------------------"
	@echo "Application run ended."

info:
	@echo " ----------------------------------------"
	@echo " Makefile Targets and Variables"
	@echo " ----------------------------------------"
	@echo " benchmark: Run benchmarks and save results"
	@echo "   Variables:"
	@echo "     benchmarkDir=$(benchmarkDir) - Directory containing benchmark tests"
	@echo "     benchmarkFunctions=$(benchmarkFunctions) - Specific benchmark functions to run (default: all)"
	@echo "     outputDir=$(outputDir) - Directory for benchmark output files"
	@echo "     benchmarkOutFile=$(benchmarkOutFile) - Output file for benchmark results"
	@echo ""
	@echo " cleanOutputs: Clean benchmark output files"
	@echo "   Variables:"
	@echo "     outputDir=$(outputDir) - Directory to clean"
	@echo ""
	@echo " cleanLogs: Clean log files"
	@echo "   Variables:"
	@echo "     logDir=$(logDir) - Directory containing log files"
	@echo ""
	@echo " compilerInfo: Show memory allocation information"
	@echo "   Variables:"
	@echo "     allocsDir=$(allocsDir) - Directory containing services to analyze"
	@echo ""
	@echo " run: Run the main application"
	@echo "   Variables:"
	@echo "     mainFile=$(mainFile) - Main Go file to execute"
	@echo ""
	@echo " Global Variables:"
	@echo "   timestamp=$(timestamp) - Timestamp used for output files"
	@echo "   tempDir=$(tempDir) - Temporary directory for intermediate files"
	@echo "   tempFile=$(tempFile) - Temporary file used during benchmarking"
	@echo " ----------------------------------------"