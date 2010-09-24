CFLAGS = -Wall -Werror
TARGETS = in.tokend pointscli claim.cgi puzzler.cgi puzzles.cgi

all: build

build: $(TARGETS)

in.tokend: in.tokend.o xxtea.o common.o
pointscli: pointscli.o common.o

puzzles.cgi: puzzles.cgi.o common.o
claim.cgi: claim.cgi.o common.o
puzzler.cgi: puzzler.cgi.o common.o

clean:
	rm -f $(TARGETS) *.o

