IMG = http://mfrost.typepad.com/cute_overload/images/2008/02/11/bunny.jpg

install: file.jpg
	cp $< $(DESTDIR)/$$(md5sum $< | awk '{print $$1;}')

src.jpg:
	wget -O src.jpg $(IMG)

file.jpg: src.jpg
	cat src.jpg text.txt > $@

clean:
	rm -f file.jpg
