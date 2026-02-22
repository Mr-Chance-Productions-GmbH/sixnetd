BINARY    = sixnetd
BUILD_DIR = build
CMD       = ./cmd/sixnetd

.PHONY: build run clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

run: build
	sudo $(BUILD_DIR)/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)/
