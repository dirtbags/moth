TARGETS = in.tokend claim.cgi puzzler.cgi pointscli puzzles

all: $(TARGETS)

in.tokend: in.tokend.o xxtea.o
puzzles: puzzles.o common.o
pointscli: pointscli.o common.o

claim.cgi: claim.cgi.o common.o
puzzler.cgi: puzzler.cgi.o common.o

clean:
	rm -f $(TARGETS) *.o

