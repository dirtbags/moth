FAKE = fakeroot -s fake -i fake
INSTALL = $(FAKE) install

all: kevin.tce

kevin.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - --exclude=placeholder --exclude=*~ .' > $@


target: kevin.py run log.run
	$(INSTALL) -D kevin.py target/usr/bin/kevin.py

	$(INSTALL) --owner=100 -d target/var/lib/kevin/tokens

	$(INSTALL) -d target/var/service/kevin
	$(INSTALL) run target/var/service/kevin/run

	$(INSTALL) -d target/var/service/kevin/log
	$(INSTALL) log.run target/var/service/kevin/log/run

clean:
	rm -rf target kevin.tce fake

%.pyc: %.py
	python3 -c 'import $*'
