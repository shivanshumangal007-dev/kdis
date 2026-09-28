run:
	go run ./cmd/main 

run-bench-set:
	go run ./bench-set .

run-bench-get:
	go run ./bench-get .

run_bench:
	$(MAKE) run > /tmp/myserver.log 2>&1 & \
	sleep 2; \
	$(MAKE) run-bench-set; \
	$(MAKE) run-bench-get