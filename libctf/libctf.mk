libctf: libctf/libctf.a
libctf/libctf.a: libctf/libctf.a(libctf/md5.o)
libctf/libctf.a: libctf/libctf.a(libctf/arc4.o)
libctf/libctf.a: libctf/libctf.a(libctf/rand.o)
libctf/libctf.a: libctf/libctf.a(libctf/token.o)

clean: libctf-clean
libctf-clean:
	rm -f libctf/*.o libctf/libctf.a
