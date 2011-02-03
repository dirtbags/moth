build: tokencli arc4

arc4: arc4.c
	$(CC) $(CFLAGS) $(LDFLAGS) -DARC4_MAIN -o $@ $<

tokencli: tokencli.o arc4.o

clean:
	rm -f *.o tokencli arc4
