PUZZLES += basemath bletchley codebreaking compaq crypto
PUZZLES += forensics hackme net-re sequence skynet webapp

-include puzzles/*/*.mk

puzzles/%-package:
	mkdir -p build/$*
	puzzles/mkpuzzles puzzles/$* build/$*
	mksquashfs build/$* $*.pkg -all-root -noappend

puzzles/%-clean:
	rm -rf build/$*

PACKAGES += $(addprefix puzzles/, $(PUZZLES))