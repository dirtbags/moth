# Scratch directory for building extrenal sources
BUILD = build

# Root to install things before they're packaged
TARGET = target

# Downloaded source files go here
CACHE = cache

# The end result
BIN = bin


all: packages

clean: packages-clean
	rm -rf $(BUILD) $(TARGET) $(BIN)

scrub: clean
	rm -rf $(CACHE)

include packages/packages.mk
