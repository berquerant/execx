GOMOD = go mod
GOBUILD = go build -trimpath -race -v
GOTEST = go test -v -cover -race

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: init
init:
	$(GOMOD) tidy

.PHONY: vuln
vuln:
	go run golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: generate
generate: clean-generated
	go generate ./...

.PHONY: clean-generated
clean-generated:
	find . -name "*_generated.go" -type f -delete
