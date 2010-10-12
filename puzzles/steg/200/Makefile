.DELETE_ON_ERROR:

$(DESTDIR)/file.gz: key text.txt
	./encode < key 3< text.txt > $@

clean:
	rm -f $(DESTDIR)/file.gz
