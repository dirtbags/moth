TARGETS = in.tokend register.cgi claim.cgi pointscli

all: $(TARGETS)

in.tokend: in.tokend.o xxtea.o

register.cgi: register.cgi.o cgi.o
claim.cgi: claim.cgi.o cgi.o common.o
pointscli: common.o pointscli.o

clean:
	rm -f $(TARGETS) *.o

