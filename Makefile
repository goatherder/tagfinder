.PHONY: build
build:
	docker build . -t tagfinder

.PHONY: clean
clean:
	rm tagfinder

