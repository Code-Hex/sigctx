deps:
	go get github.com/mattn/goveralls

test:
	go test -v

cover: deps
	goveralls

.PHONY: test deps cover