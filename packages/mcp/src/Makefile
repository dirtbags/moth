CFLAGS = -Wall -Werror
TARGETS = claim.cgi puzzler.cgi puzzles.cgi
TARGETS += pointscli

all: build

build: $(TARGETS)

pointscli: pointscli.o common.o

puzzles.cgi: puzzles.cgi.o common.o
claim.cgi: claim.cgi.o common.o
puzzler.cgi: puzzler.cgi.o common.o

clean:
	rm -f $(TARGETS) *.o

