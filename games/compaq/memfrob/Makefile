LDFLAGS = -static

all: frob main

main: CFLAGS = -Dfrobbed='"$(shell tr -d '\n' < key | ./frob)"'

clean:
	rm -rf main frob

