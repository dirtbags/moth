PUZZLES = sequence codebreaking

-include puzzles/*/*.mk

puzzles/%-package:
	mkdir -p build/$*
	puzzles/mkpuzzles puzzles/$* build/$*
	mksquashfs build/$* $*.pkg -all-root -noappend

puzzles/%-clean:
	rm -rf build/$*

PACKAGES += $(addprefix puzzles/, $(PUZZLES))