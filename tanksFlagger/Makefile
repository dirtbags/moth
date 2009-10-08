FAKE = fakeroot -s fake -i fake
INSTALL = $(FAKE) install -o 100

all: tanksFlagger.tce

push: tanksFlagger.tce
	netcat -l -q 0 -p 3333 < tanksFlagger.tce	

tanksFlagger.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - .' > $@

target:
	$(INSTALL) -d target/var/service/tanksFlagger
	$(INSTALL) run report_score.py target/var/service/tanksFlagger/

	$(INSTALL) -d target/var/service/tanksFlagger/log/
	$(INSTALL) log.run target/var/service/tanksFlagger/log/run

clean:
	rm -rf target tanksFlagger.tce fake
