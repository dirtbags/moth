CFLAGS = -Wall -Werror
TARGETS = in.tokend claim.cgi
TARGETS += puzzler.cgi puzzles.cgi
TARGETS += pointscli mktoken

all: build

build: $(TARGETS)

in.tokend: in.tokend.o arc4.o md5.o common.o
tokencli: tokencli.o arc4.o
pointscli: pointscli.o common.o
mktoken: mktoken.o common.o
arc4: arc4-main.o arc4.o

puzzles.cgi: puzzles.cgi.o common.o
claim.cgi: claim.cgi.o common.o
puzzler.cgi: puzzler.cgi.o common.o

clean:
	rm -f $(TARGETS) *.o

