.PHONY: check fmt vet

.DEFALT_GOAL:= check

check:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
