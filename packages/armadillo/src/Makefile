CFLAGS = -Wall -Werror
TARGETS = gimmie straceme killme dillo

all: build

build: $(TARGETS)

gimmie: gimmie.o token.o arc4.o
straceme: straceme.o token.o arc4.o
killme: killme.o token.o arc4.o
dillo: dillo.o token.o arc4.o

install: $(TARGETS)
	install -m 0755 $(TARGETS) $(DESTDIR)/bin

clean:
	rm -f *.o $(TARGETS)
