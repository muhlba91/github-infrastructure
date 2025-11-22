TESTPARALLELISM := 8

WORKING_DIR := $(shell pwd)

.PHONY: lint
lint::
	golangci-lint run -c .golangci.yml
	go vet ./...

.PHONY: fix
fix::
	golangci-lint fmt -c .golangci.yml

.PHONY: test
test::
	go test -v -tags=all -parallel ${TESTPARALLELISM} -timeout 2h -covermode atomic -coverprofile=covprofile github.com/muhlba91/muehlbachler-shared-services/pkg/...

.PHONY: coverage
coverage::
	go tool cover -html=covprofile -o coverage.html
	open coverage.html
