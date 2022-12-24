build:
	go build -o bin/sigChecker

run: build
	./bin/sigChecker