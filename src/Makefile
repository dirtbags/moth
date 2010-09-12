TARGETS = in.tokend claim.cgi puzzler.cgi pointscli

all: $(TARGETS)

in.tokend: in.tokend.o xxtea.o

claim.cgi: claim.cgi.o cgi.o common.o
puzzler.cgi: puzzler.cgi.o cgi.o common.o
pointscli: common.o pointscli.o

clean:
	rm -f $(TARGETS) *.o

