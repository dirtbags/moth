mom.jpg: key.bz2 1GiB_nulls.bz2
	rm -f $@
	for i in $$(seq 300); do cat 1GiB_nulls.bz2 >> $@; done
	cat key.bz2 >> $@
	for i in $$(seq 200); do cat 1GiB_nulls.bz2 >> $@; done

key.bz2: key
	cat $< | bzip2 > $@

1GiB_nulls.bz2:
	dd if=/dev/zero bs=1G count=1 | bzip2 > $@
