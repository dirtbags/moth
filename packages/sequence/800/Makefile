all: index.mdwn

.DELETE_ON_ERROR:

index.mdwn: key convert.py
	./convert.py | hd > $@

clean:
	rm -f index.mdwn
