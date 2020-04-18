default: test

all: test lint

test: dirs
	go test ./... -race -count=1 -coverprofile=bin/coverage.out

soaktest: dirs
	SOAK_CMD="make test" sh/soak.sh

dirs:
	mkdir -p bin

lint:
	golint ./...

clean:
	rm -rf bin

list: FORCE
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

FORCE:
