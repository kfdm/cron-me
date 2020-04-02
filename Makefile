# Go parameters
GO111MODULE=on
GO ?= go
GIT ?= git


PLATFORMS=darwin linux
BINARIES:= cron-shell cron-me


.PHONY:	all $(BINARIES) clean

default: $(BINARIES)
all: ${PLATFORMS}
clean:
	${GIT} clean -dxf bin

$(BINARIES):
	$(GO) build -o $@ cmd/$@/main.go

${PLATFORMS}:
	@mkdir -p bin/$@
	GOOS=$@ $(GO) build -o bin/$@/cron-shell cmd/cron-shell/main.go
	GOOS=$@ $(GO) build -o bin/$@/cron-me cmd/cron-me/main.go
