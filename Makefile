run:
	go run ./cmd/app/main.go

build-mac:
	env GOOS=darwin GOARCH=amd64 go build -o ./build/app ./cmd/app/main.go

build-win:
	go build -o ./build/app.exe ./cmd/app/main.go

build-linux:
	env GOOS=linux GOARCH=amd64 GOEXPERIMENT=greenteagc go build -ldflags="-s -w" -o ./build/app ./cmd/app/main.go

run-kitchen:
	go run ./cmd/kitchen_sink/main.go