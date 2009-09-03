LDFLAGS = -static

all: main

main: CFLAGS = -Dpasswd='"$(shell cat key)"'

clean:
	rm -rf main

