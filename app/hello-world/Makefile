BUILD_DATE?=$(shell date +'%Y-%m-%d_%H:%M:%S')
GITHUB_SHA?=unassigned


build-HelloWorldFunction:
	go test .
	CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/hello-world -trimpath -ldflags='-w -s -X main.VERSION=$(GITHUB_SHA) -X main.BUILD_DATE=$(BUILD_DATE)' *.go
