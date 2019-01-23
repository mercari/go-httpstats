.PHONY: dep check fmt vet

.DEFALT_GOAL:= check

dep:
	@dep ensure

check: dep
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
