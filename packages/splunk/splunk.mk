splunk-source:
splunk-build:

splunk-install: packages/splunk/tokens.txt
	mkdir -p $(TARGET)/splunk/
	cp $< $(TARGET)/splunk/

PACKAGES += splunk
