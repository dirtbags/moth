.DELETE_ON_ERROR:

install: file.png
	cp $< $(DESTDIR)/$$(md5sum $< | awk '{print $$1;}')

%.pbm: %.png
	pngtopnm $< > $@

%.png: %.pnm
	pnmtopng $< > $@

file.pnm: lena.pbm text.txt
	m4 -DKEY="$(shell cat key)" text.txt | ./encode 3< $< > $@

clean:
	rm -f file.pnm file.png lena.pbm
