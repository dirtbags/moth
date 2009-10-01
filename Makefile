DESTDIR = target

CTFDIR = $(DESTDIR)/usr/lib/ctf
WWWDIR = $(DESTDIR)/usr/lib/www

FAKE = fakeroot -s fake -i fake
INSTALL = $(FAKE) install

PYC  = config.pyc points.pyc game.pyc teams.pyc
PYC += register.pyc scoreboard.pyc puzzler.pyc
PYC += flagd.pyc pointsd.pyc pointscli.pyc
PYC += roshambo.pyc histogram.pyc

all: ctf.tce

target: $(PYC)
	$(INSTALL) -d --mode=0755 --owner=100 $(DESTDIR)/var/lib/ctf
	$(INSTALL) -d $(DESTDIR)/var/lib/ctf/disabled
	touch $(DESTDIR)/var/lib/ctf/disabled/survey

	$(INSTALL) -d $(CTFDIR)
	$(INSTALL) $(PYC) $(CTFDIR)
	$(INSTALL) ctfd.py $(CTFDIR)

	$(INSTALL) -d $(WWWDIR)
	$(INSTALL) index.html intro.html ctf.css grunge.png $(WWWDIR)
	$(FAKE) ln -s /var/lib/ctf/histogram.png $(WWWDIR)
	$(INSTALL) register.cgi scoreboard.cgi puzzler.cgi $(WWWDIR)

	$(INSTALL) -d $(DESTDIR)/var/service/ctf
	$(INSTALL) run.ctfd $(DESTDIR)/var/service/ctf/run

	$(INSTALL) -d $(DESTDIR)/var/service/ctf/log
	$(INSTALL) run.log.ctfd $(DESTDIR)/var/service/ctf/log/run

	rm -rf $(WWWDIR)/puzzler
	$(INSTALL) -d $(WWWDIR)/puzzler
	./mkpuzzles.py --htmldir=$(WWWDIR)/puzzler --keyfile=$(CTFDIR)/puzzler.keys

ctf.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - --exclude=placeholder --exclude=*~ .' > $@

clean:
	rm -rf target
	rm -f fake ctf.tce $(PYC)

%.pyc: %.py
	python3 -c 'import $*'
