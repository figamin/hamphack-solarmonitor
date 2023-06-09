BINARY_NAME=hamp-solar-monitor

build:
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin-x86 main.go
	GOARCH=arm64 GOOS=darwin go build -o ${BINARY_NAME}-darwin-arm main.go
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
	GOARCH=amd64 GOOS=windows go build -o ${BINARY_NAME}-windows main.go

run: build
	./${BINARY_NAME}-darwin-arm

clean:
	go clean
	rm ${BINARY_NAME}-darwin
	rm ${BINARY_NAME}-linux
	rm ${BINARY_NAME}-windows