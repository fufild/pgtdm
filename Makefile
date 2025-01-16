BINARY_NAME = pgtdm
COVER_OUT = cover.out
COVER_HTML = cover.html
all: build
build:
	go build -o $(BINARY_NAME) ./cmd/main.go
	chmod +x $(BINARY_NAME)
test:
	go test ./... -coverprofile=$(COVER_OUT)
coverage: test
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
open: coverage
	open $(COVER_HTML)
clean:
	rm -f $(BINARY_NAME) $(COVER_OUT) $(COVER_HTML)
.PHONY: all build test coverage open clean
