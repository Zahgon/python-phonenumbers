GO ?= go
GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*')
LOCALE_SOURCE := tools/locale/DumpLocale.java
LOCALE_CLASS := tools/locale/DumpLocale.class
LOCALE_OUTPUT := phonenumbers/geodata/locale_generated.go

.PHONY: all build test vet fmt cover locale upstream clean

all: build

build:
	$(GO) build ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	@unformatted="$$(gofmt -l $(GOFILES))"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following Go files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

cover:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

$(LOCALE_CLASS): $(LOCALE_SOURCE)
	javac $(LOCALE_SOURCE)

locale: $(LOCALE_CLASS)
	java -cp tools/locale DumpLocale > $(LOCALE_OUTPUT).tmp
	mv $(LOCALE_OUTPUT).tmp $(LOCALE_OUTPUT)

upstream:
	$(GO) run ./tools/checkupstreamversion

clean:
	$(GO) clean ./...
	rm -f coverage.out $(LOCALE_CLASS) $(LOCALE_OUTPUT).tmp
