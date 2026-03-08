.PHONY: dep
dep:
	go mod tidy
	go mod vendor

.PHONY: record-demo
record-demo: dep
	go build ./cmd/...
	vhs demo.tape