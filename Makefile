# Go source files
SRCS := $(shell find . -type d -name 'vendor' -prune -o -name '*.go' -print)

.PHONY: install
install: glide #download dependencies (including test deps) for the package
	glide install

.PHONY: update
update: glide #updates dependencies used by the package and installs them
	glide update

.PHONY: glide
glide: # ensure the glide package tool is installed
	which glide || go get github.com/Masterminds/glide

.PHONY: lint
lint: #lints the package for common code smells
	@for file in $(SRCS); do \
		gofmt -d -s $${file}; \
		if [ -n "$$(gofmt -d -s $${file})" ]; then \
			exit 1; \
		fi; \
	done
	which golint || go get -u golang.org/x/lint/golint
	which shadow || go get golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	golint -set_exit_status $(shell go list ./...)
	go vet -all ./...
	go vet -vettool=$(which shadow) ./...

.PHONY: test
test: # runs all tests against the package with race detection and coverage percentage
	go test -race -cover ./...

.PHONY: quick
quick: # runs all tests without coverage or the race detector
	go test ./...

.PHONY: cover
cover: # runs all tests against the package, generating a coverage report and opening it in the default browser
	go test -race -covermode=atomic -coverprofile=cover.out ./...
	go tool cover -html cover.out -o cover.html
	which open && open cover.html

