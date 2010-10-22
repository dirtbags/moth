.DELETE_ON_ERROR:

DESTDIR ?= .

JPEGS += ww0207-43.jpg  ww0870-11.jpg  ww1645-44.jpg  ww1646-66.jpg  ww1646-78.jpg
JPEGS += ww0207-45.jpg  ww0870-14.jpg  ww1645-52.jpg  ww1646-67.jpg  ww1647-37.jpg
JPEGS += ww0207-90.jpg  ww1645-22.jpg  ww1645-53.jpg  ww1646-69.jpg  ww1647-75.jpg
JPEGS += ww0870-09.jpg  ww1645-43.jpg  ww1645-55.jpg  ww1646-70.jpg  ww1647-85.jpg

PRISTINE = $(addprefix pristine/, $(JPEGS))
NORMALIZED = $(addprefix normalized/, $(JPEGS))

install: $(DESTDIR)/file.zip

.PRECIOUS: pristine/%.jpg
pristine/%.jpg:
	@mkdir -p $(@D)
	wget -O $@ http://digital.library.northwestern.edu/wwii-posters/img/$*.jpg

normalized/%.jpg: pristine/%.jpg image.cmds
	@mkdir -p $(@D)
	cp $< $@
	exiv2 -m image.cmds $@

key.png: key
	pbmtext < key | pnmtopng -compression 0 > $@

$(DESTDIR)/file.zip: key.png $(NORMALIZED)
	./encode $(NORMALIZED) < key.png > $@
	@echo $(NORMALIZED)

clean:
	rm -rf key.png $(DESTDIR)/file.zip
	rm -rf normalized/