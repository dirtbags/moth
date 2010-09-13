SUBDIRS = src

all: build

include $(addsuffix /*.mk, $(SUBDIRS))

test: build
	./test.sh

build: $(addsuffix -build, $(SUBDIRS))
clean: $(addsuffix -clean, $(SUBDIRS))

