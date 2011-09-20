CFLAGS = -Wall -Werror
LDFLAGS = -static
TARGETS = logger

all: build
build: $(TARGETS)

logger: logger.o arc4.o token.o

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)
