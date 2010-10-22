install: file.mp3
	cp $< $(DESTDIR)/$$(md5sum $< | awk '{print $$1;}')

file.mp3: src.mp3 text.txt
	dd if=$< bs=4 count=1 > $@
	dd if=/dev/zero bs=4 count=3 >> $@
	m4 KEY="$(shell cat key)" text.txt | zip - - >> $@
	cat $< >> $@

clean:
	rm -f file.mp3
