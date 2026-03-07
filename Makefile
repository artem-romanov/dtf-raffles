APP=app
BUILD_DIR=build
CMD_APP=./cmd/app/main.go
CMD_KITCHEN=./cmd/kitchen_sink/main.go

run:
	go run $(CMD_APP)

run-kitchen:
	go run $(CMD_KITCHEN)

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP) $(CMD_APP)

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP).exe $(CMD_APP)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP) $(CMD_APP)

compress:
	upx -8 $(BUILD_DIR)/$(APP)

release-linux: build-linux compress