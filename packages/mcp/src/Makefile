CFLAGS = -Wall -Werror
TARGETS = claim.cgi puzzler.cgi puzzles.cgi
TARGETS += pointscli arc4 tea

all: build

build: $(TARGETS)

arc4: CFLAGS += -DARC4_MAIN
arc4: arc4.o

pointscli: pointscli.o common.o

puzzles.cgi: puzzles.cgi.o common.o
claim.cgi: claim.cgi.o common.o
puzzler.cgi: puzzler.cgi.o common.o

clean:
	rm -f $(TARGETS) *.o

