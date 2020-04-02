# Go parameters
GO111MODULES=on
GOCMD=go
GOBUILD=$(GOCMD) build

obj_files:= cron-shell cron-me

all: $(obj_files)

.PHONY:	test
test:
	$(GOCMD) test .

run: build
	./$(BINARY_NAME) fortune
	./$(BINARY_NAME) false

.PHONY:	$(obj_files)
$(obj_files):
	$(GOBUILD) -o $@ cmd/$@/main.go

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
docker-build:
	docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_UNIX)" -v
