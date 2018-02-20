.DEFAULT: punchabunch
.PHONY: clean deps

GOVENDOR = $(GOPATH)/bin/govendor

punchabunch: deps punchabunch.go lib/*.go
	go build

deps:
	go get github.com/golang/dep/cmd/dep
	$(GOPATH)/bin/dep ensure

clean:
	-rm -f punchabunch
