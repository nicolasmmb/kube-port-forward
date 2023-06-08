BINARY_NAME=pf.so


build: 
	go build -o bin/${BINARY_NAME} -tags=release -ldflags "-s -w" main.go 