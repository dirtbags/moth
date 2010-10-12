PUZZLES += basemath bletchley codebreaking compaq crypto
PUZZLES += forensics hackme net-re sequence skynet webapp
PUZZLES += steg

PUZZLES_SUBMAKEFILES = $(wildcard puzzles/*/*/Makefile)
PUZZLES_SUBCLEANS = $(patsubst %/Makefile, %/clean, $(PUZZLES_SUBMAKEFILES))

install: $(patsubst %, puzzles/%-install, $(PUZZLES))
puzzles/%-install:
	mkdir -p build/$*
	puzzles/mkpuzzles puzzles/$* build/$*
	touch $@
%.pkg: puzzles/%-install
	mksquashfs build/$* $*.pkg -all-root -noappend

clean: puzzles-clean
clean: $(patsubst %, puzzles/%-clean, $(PUZZLES))
puzzles-clean: $(PUZZLES_SUBCLEANS) $(patsubst %, puzzles/%-clean, $(PUZZLES))
puzzles/%/clean:
	$(MAKE) -C $(@D) clean
puzzles/%-clean: $(PUZZLES_SUBCLEANS)
	rm -rf build/$* puzzles/$*-install $*.pkg

packages: $(addsuffix .pkg, $(PUZZLES))
