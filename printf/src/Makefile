CFLAGS = -Wall -Werror
LDFLAGS = -static
TARGETS = printf

all: build
build: $(TARGETS)

printf: printf.o token.o

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)