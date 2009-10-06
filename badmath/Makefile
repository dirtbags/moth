FAKE = fakeroot -s fake -i fake
INSTALL = $(FAKE) install

all: badmath.tce

badmath.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - --exclude=placeholder --exclude=*~ .' > $@


target:
	$(INSTALL) -d target/usr/lib/ctf/badmath
	$(INSTALL) Gyopi.py badmath.py target/usr/lib/ctf/badmath

	$(INSTALL) -d target/var/lib/badmath

	$(INSTALL) -d target/var/service/badmath
	$(INSTALL) run target/var/service/badmath/run

	$(INSTALL) -d target/var/service/badmath/log
	$(INSTALL) log.run target/var/service/badmath/log/run

clean:
	rm -rf target badmath.tce fake

%.pyc: %.py
	python3 -c 'import $*'
