FAKE = fakeroot -s fake -i fake
INSTALL = $(FAKE) install -o 100

all: tanks.tce

tanks.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - .' > $@

target:
	$(INSTALL) -d target/var/lib/tanks/
	$(INSTALL) -d target/var/lib/tanks/results/
	$(INSTALL) -d target/var/lib/tanks/ai/easy
	$(INSTALL) -d target/var/lib/tanks/ai/medium
	$(INSTALL) -d target/var/lib/tanks/ai/hard
	$(INSTALL) -d target/var/lib/tanks/ai/players
	$(INSTALL) AI/easy/* target/var/lib/tanks/ai/easy/
	$(INSTALL) AI/medium/* target/var/lib/tanks/ai/medium/
	$(INSTALL) AI/hard/* target/var/lib/tanks/ai/hard/

	$(INSTALL) -d target/usr/lib/www/tanks/
	$(INSTALL) www/* target/usr/lib/www/tanks/

	$(INSTALL) -d target/usr/lib/python2.6/site-packages/tanks/
	$(INSTALL) lib/* target/usr/lib/python2.6/site-packages/tanks/

	$(INSTALL) -d target/var/service/tanks
	$(INSTALL) run run_tanks.py target/var/service/tanks/
	
	$(INSTALL) -d target/var/service/tanks/log/
	$(INSTALL) log.run target/var/service/tanks/log/run

clean:
	rm -rf target tanks.tce fake
