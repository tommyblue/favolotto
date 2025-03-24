# Nome del binario generato
BINARY=favolotto

# Percorso della cartella dei sorgenti
SRC_DIR=.

# Architettura target (può essere cambiata in arm, arm64, etc.)
GOARCH=arm
GOARM=6  # Per Raspberry Pi 2, 3. Usa GOARM=6 per Pi 1 e Zero, GOARM=8 per Pi 3/4 a 64 bit
TARGET_OS=linux


# Toolchain di cross-compilazione per CGO
CC_arm=arm-linux-gnueabihf-gcc
CC_arm64=aarch64-linux-gnu-gcc

# Imposta il compilatore C corretto in base all'architettura
ifeq ($(GOARCH),arm)
    CC=$(CC_arm)
else ifeq ($(GOARCH),arm64)
    CC=$(CC_arm64)
endif

# Impostazioni CGO
CGO_ENABLED=1
CGO_CFLAGS=-I/home/asterix/src/rpi/rootfs/usr/local/include \
					 #

CGO_LDFLAGS=-lnfc_nci_linux -lpthread \
						-L/home/asterix/src/rpi/rootfs/usr/local/lib \
						-Wl,-rpath=$(LIB_DIR) \
						#

# Flag di compilazione
BUILD_FLAGS=-ldflags="-s -w"

# Directory di output
BUILD_DIR=build

all: build

build:
	@mkdir -p $(BUILD_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
	CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) \
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY) $(SRC_DIR)
clean:
	rm -rf $(BUILD_DIR)

.PHONY: all build clean

