TARGETS = in.tokend register.cgi claim.cgi

all: $(TARGETS)

in.tokend: in.tokend.o xxtea.o

register.cgi: register.cgi.o cgi.o
claim.cgi: claim.cgi.o cgi.o pointscli.o

clean:
	rm -f $(TARGETS) *.o