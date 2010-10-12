IMG = http://www.myconfinedspace.com/wp-content/uploads/2006/11/bill-stickers.jpg

.DELETE_ON_ERROR:

install: file.png
	cp $< $(DESTDIR)/$$(md5sum $< | awk '{print $$1;}')

src.jpg:
	wget -O $@ $(IMG)

%.pbm: %.jpg
	jpegtopnm $< > $@

alpha.pgm: src.pbm text.txt
	echo "P5" > $@
	awk '(i++ == 1) { print; }' $< >> $@
	echo "255" >> $@
	m4 -DKEY="$(shell cat key)" text.txt >> $@	
	awk 'BEGIN{for(i=0;i<500000;i++)printf("\377");}' >> $@

file.png: alpha.pgm src.pbm
	pnmtopng -alpha alpha.pgm src.pbm > $@

clean:
	rm -f file.png alpha.pgm src.pbm
