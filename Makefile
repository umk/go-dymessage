PROTO_FILES=$(shell find $(CURDIR) -name '*.proto' -type f)
PROTO_GO_FILES=$(patsubst %.proto, %.pb.go, $(PROTO_FILES))

TESTFIX_FLAGS=-X 'github.com/umk/go-testutil.fix=fix'

gen: $(PROTO_GO_FILES)

test: gen
	go test ./...

testfix: gen
	go test -ldflags="$(TESTFIX_FLAGS)" ./...

$(PROTO_GO_FILES): $(PROTO_FILES)
	protoc --proto_path=$(dir $<) --go_out=$(dir $<) $^

$(PROTO_TEST_GO_FILES): $(PROTO_TEST_FILES)
	protoc --proto_path=$(dir $<) --go_out=$(dir $<) $^

format:
	gofmt -w -s $(CURDIR)

clean:
	find $(CURDIR) -name '*.pb.go' -type f -exec rm '{}' \;

.PHONY: gen test testfix format clean
