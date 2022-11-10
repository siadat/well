test:
	go list ./... | while read -r pkg; do \
		go test -failfast -count=1 -v "$$pkg" || exit 1; \
	done
	@echo "All tests passed"

all:
	go test -v -count=1 ./...

test-examples:
	@go run ./cmd/well/ run -f ./testdata/test1.well

todo:
	grep --color=always -Po 'TODO.*' -R .
