CFLAGS = -Wall -Werror
TARGETS = multicaster

all: build
build: $(TARGETS)

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)
