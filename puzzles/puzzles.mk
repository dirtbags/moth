PUZZLES += basemath bletchley codebreaking compaq crypto
PUZZLES += forensics hackme net-re sequence skynet webapp

-include puzzles/*/*.mk

puzzles/%-install:
	mkdir -p build/$*
	puzzles/mkpuzzles puzzles/$* build/$*
	touch $@

puzzles/%-clean:
	rm -rf build/$* puzzles/$*-install

%.pkg: puzzles/%-install
	mksquashfs build/$* $*.pkg -all-root -noappend

packages: $(addsuffix .pkg, $(PUZZLES))
install: $(patsubst %, puzzles/%-install, $(PUZZLES))
clean: $(patsubst %, puzzles/%-clean, $(PUZZLES))
