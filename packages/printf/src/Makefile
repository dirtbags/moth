CFLAGS = -Wall
LDFLAGS = -static
TARGETS = printf

all: build
build: $(TARGETS)

printf: printf.o

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)