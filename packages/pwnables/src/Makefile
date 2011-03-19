CFLAGS = -Wall -Werror
TARGETS = gimmie ltraceme straceme killme

all: build

build: $(TARGETS)

gimmie: gimmie.o token.o arc4.o
octopus: octopus.o token.o arc4.o
ltraceme: ltraceme.o token.o arc4.o
straceme: straceme.o token.o arc4.o
killme: killme.o token.o arc4.o

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)