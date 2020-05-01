# Go parameters
GO111MODULE=on
GO ?= go
GIT ?= git

BINARIES:= cron-shell cron-me

.PHONY:	all $(BINARIES) clean test ship

default: $(BINARIES)
all: ${PLATFORMS}
clean:
	${GIT} clean -dxf dist

$(BINARIES):
	$(GO) build -o $@ cmd/$@/main.go

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist

release:
	goreleaser

test:
	$(GO) test -v ./...
