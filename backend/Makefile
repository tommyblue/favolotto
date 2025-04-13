BINARY=favolotto
SRC_DIR=./cmd/favolotto
GOARCH=arm
GOARM=6  # Use GOARM=6 for Pi 1 and Zero, GOARM=8 for 64 bit Pi 3/4
TARGET_OS=linux


# CGO cross-compilation toolchain
CC_arm=arm-linux-gnueabihf-gcc
CC_arm64=aarch64-linux-gnu-gcc

# ISet the correct compiler based on the architecture
ifeq ($(GOARCH),arm)
    CC=$(CC_arm)
else ifeq ($(GOARCH),arm64)
    CC=$(CC_arm64)
endif

# CGO settings
CGO_ENABLED=1
CGO_CFLAGS= \
	-I/home/asterix/src/rpi/rootfs/nfc_build/include \
	-I/home/asterix/src/rpi/rootfs/usr/local/include \
	#-I/home/asterix/src/rpi/rootfs/include \
	#

#-lnfc_nci_linux -lpthread -lusb -lpcsclite\

CGO_LDFLAGS= \
  -lnfc_nci_linux -lpthread -lnfc \
	-L/home/asterix/src/rpi/rootfs/nfc_build/lib \
	-L/home/asterix/src/rpi/rootfs/usr/local/lib \
  #-L/home/asterix/src/rpi/rootfs/lib/arm-linux-gnueabihf \
	#-Wl,-rpath=/home/asterix/src/rpi/rootfs/lib/  \
	#

# Compilation flags
BUILD_FLAGS=-ldflags="-s -w"

# Build tags
BUILD_TAGS=--tags "pn532 pn7150"

# Output folder
BUILD_DIR=bin

all: build

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the app with correct environment for the rpi
	@mkdir -p $(BUILD_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
	CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) \
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	go build $(BUILD_FLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(BINARY) $(SRC_DIR)

clean: ## Clean the build directory
	rm -rf $(BUILD_DIR)

.PHONY: all build clean help

