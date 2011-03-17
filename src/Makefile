TARGETS = bubblebabble arc4

all: $(TARGETS)

arc4: CFLAGS=-DARC4_MAIN

tea: tea.o xxtea.o

clean:
	rm -f *.o $(TARGETS)