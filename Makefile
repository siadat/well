test:
	go list ./... | while read -r pkg; do \
		go test -failfast -count=1 -v "$$pkg" || exit 1; \
	done
	@echo "All tests passed"

all:
	go test -v -count=1 ./...

examples:
	# cp _example1.go /tmp/example1.go
	# cp _example2.go /tmp/example2.go
	# cp _example3.go /tmp/example3.go
	# go run /tmp/example1.go
