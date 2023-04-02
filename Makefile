BUILD_DIR := ./cmd/...
ARCHS := amd64 arm64 armv7
GO_JUNIT_REPORT := go run github.com/jstemmer/go-junit-report/v2@latest

.PHONY: build test verify exporter log-battery

verify:
	go mod verify

build: verify
	go build -o dist/ $(BUILD_FLAGS) $(BUILD_DIR)

build-%:
	@_cmd="$(subst -$(lastword $(subst -, ,$*)),,$*)" ; \
	_arch="$(lastword $(subst -, ,$*))" ; \
	case "$${_arch}" in \
		arm64) GOARCH=arm64 ;; \
		armv7) GOARCH=arm; GOARM=7 ;; \
		amd64) GOARCH=amd64 ;; \
		*) echo "unsupported architecture $${_arch}"; exit 1 ;; \
	esac; \
	GOOS=linux GOARCH="$${GOARCH}" GOARM="$${GOARM}" \
	go build -o "dist/$${_cmd}-$${_arch}" $(BUILD_FLAGS) "./cmd/$${_cmd}"

exporter: $(addprefix build-exporter-, $(ARCHS))
log-battery: $(addprefix build-log-battery-, $(ARCHS))

release: verify exporter log-battery

test:
	go test -count 1 -v -race ./... 2>&1 | $(GO_JUNIT_REPORT) -iocopy -out test-results.xml -set-exit-code
