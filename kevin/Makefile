FAKE = fakeroot -s fake -i fake
INSTALL = $(FAKE) install

all: kevin.tce

kevin.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - --exclude=placeholder --exclude=*~ .' > $@


target: kevin.py irc.pyc run log.run
	$(INSTALL) -d target/usr/lib/ctf
	$(INSTALL) kevin.py irc.py target/usr/lib/ctf

	$(INSTALL) --owner=100 -d target/var/lib/ctf/kevin/tokens

	$(INSTALL) -d target/var/service/kevin
	$(INSTALL) run target/var/service/kevin/run

	$(INSTALL) -d target/var/service/kevin/log
	$(INSTALL) log.run target/var/service/kevin/log/run

clean:
	rm -rf target kevin.tce fake

%.pyc: %.py
	python3 -c 'import $*'
