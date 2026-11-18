.PHONY: build run test check mutate

build:
	@mkdir -p bin
	@set -eu; \
	work=$$(mktemp -d bin/build.XXXXXX); \
	trap 'rm -rf "$$work"' EXIT; \
	for arch in arm64 amd64; do \
		CGO_ENABLED=0 GOOS=darwin GOARCH=$$arch go build -trimpath -o "$$work/$$arch" ./cmd/go-todo; \
	done; \
	xcrun lipo -create "$$work/arm64" "$$work/amd64" -output "$$work/go-todo"; \
	archs=$$(xcrun lipo -archs "$$work/go-todo"); \
	case "$$archs" in "arm64 x86_64"|"x86_64 arm64") ;; \
		*) echo "Unexpected architectures: $$archs" >&2; exit 1 ;; esac; \
	mv "$$work/go-todo" bin/go-todo

run:
	@go run ./cmd/go-todo

test:
	go test -count=1 ./...

check:
	@set -eu; files=$$(gofmt -l cmd internal tests); \
	if [ -n "$$files" ]; then printf 'Run gofmt on:\n%s\n' "$$files"; exit 1; fi
	go vet ./...
	$(MAKE) test

mutate:
	@cd internal/cli && { \
		report=$$(go tool gremlins unleash --timeout-coefficient 30) || { printf '%s\n' "$$report"; exit 1; }; \
		printf '%s\n' "$$report"; \
		printf '%s\n' "$$report" | grep -Eq '^Killed: [1-9][0-9]*, Lived: 0, Not covered: 0$$' || \
			{ echo 'Expected at least one killed mutation and none surviving or uncovered.' >&2; exit 1; }; \
	}
