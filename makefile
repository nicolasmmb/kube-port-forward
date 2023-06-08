BINARY_NAME=pf.so


build: 
	go build -o bin/${BINARY_NAME} -ldflags "-s -w" main.go