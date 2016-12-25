.DEFAULT: punchabunch
.PHONY: clean deps

GOVENDOR = $(GOPATH)/bin/govendor

punchabunch: deps punchabunch.go lib/*.go
	go build

deps:
	go get github.com/kardianos/govendor
	$(GOPATH)/bin/govendor sync
clean:
	-rm -f punchabunch
