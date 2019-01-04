PROTO_FILES=$(shell find $(CURDIR) -name '*.proto' -type f)
PROTO_GO_FILES=$(patsubst %.proto, %.pb.go, $(PROTO_FILES))

TESTFIX_FLAGS=-X 'github.com/umk/go-testutil.fix=fix'
COVERAGE_FILE=.testcover

gen: $(PROTO_GO_FILES)

test: gen
	go test ./...

testfix: gen
	go test -ldflags="$(TESTFIX_FLAGS)" ./...

cover:
	go test -coverprofile $(CURDIR)/$(COVERAGE_FILE) ./...
	go tool cover -html=$(CURDIR)/$(COVERAGE_FILE)

$(PROTO_GO_FILES): $(PROTO_FILES)
	protoc --proto_path=$(dir $<) --go_out=$(dir $<) $^

$(PROTO_TEST_GO_FILES): $(PROTO_TEST_FILES)
	protoc --proto_path=$(dir $<) --go_out=$(dir $<) $^

format:
	gofmt -w -s $(CURDIR)

clean:
	find $(CURDIR) -name '*.pb.go' -type f -exec rm '{}' \;

.PHONY: gen test testfix cover format clean
