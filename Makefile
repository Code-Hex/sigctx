deps:
	go get github.com/mattn/goveralls

test:
	go test -cover

cover: deps
	goveralls

.PHONY: test deps cover