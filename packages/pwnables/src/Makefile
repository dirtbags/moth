CFLAGS = -Wall -Werror
TARGETS = gimmie ltraceme straceme killme

all: build

build: $(TARGETS)

gimmie: gimmie.o token.o
octopus: octopus.o token.o
ltraceme: ltraceme.o token.o
straceme: straceme.o token.o
killme: killme.o token.o

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)