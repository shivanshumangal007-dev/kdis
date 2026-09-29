run:
	go run ./cmd/main 
build: 
	go build -o output/kdis.exe ./cmd/main 

run-bench-set:
	go run ./bench-set .

run-bench-get:
	go run ./bench-get .

run_bench:
	@$(MAKE) run > /tmp/myserver.log 2>&1 & \
	SERVER_PID=$$!; \
	trap 'echo "Stopping server..."; kill $$SERVER_PID 2>/dev/null || true' EXIT INT TERM; \
	echo "Server started (PID $$SERVER_PID)"; \
	sleep 2; \
	$(MAKE) run-bench-set && \
	$(MAKE) run-bench-get