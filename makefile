run:
	go run ./cmd/main 
build-mac: 
	go build -o bin/kdis_mac.exe ./cmd/main 
build-win:
	GOOS=windows GOARCH=amd64 go build -o bin/kdis_win.exe ./cmd/main 
build:
	${MAKE} build-mac
	${MAKE} build-win
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