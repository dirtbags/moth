package transpile

import (
	"github.com/spf13/afero"
)

var testMothYaml = []byte(`---
answers:
  - YAML answer
authors:
  - Arthur
  - Buster
  - DW
attachments:
  - moo.txt
---
YAML body
`)
var testMothRfc822 = []byte(`author: test
Author: Arthur
author: Fred Flintstone
answer: RFC822 answer

RFC822 body
`)
var testMothMarkdown = []byte(`---
answers:
  - answer
authors:
  - Fred
---

one | two
--- | ---
1 | 2

Term
:  definition of that term
`)

func newTestFs() afero.Fs {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "cat0/1/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/1/moo.txt", []byte("Moo."), 0644)
	afero.WriteFile(fs, "cat0/2/puzzle.md", testMothRfc822, 0644)
	afero.WriteFile(fs, "cat0/3/puzzle.moth", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/4/puzzle.md", testMothMarkdown, 0644)
	afero.WriteFile(fs, "cat0/5/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/10/puzzle.md", []byte(`---
Answers:
  - moo
Authors:
  - bad field
---
body
`), 0644)
	afero.WriteFile(fs, "cat0/20/puzzle.md", []byte("Answer: no\nBadField: yes\n\nbody\n"), 0644)
	afero.WriteFile(fs, "cat0/21/puzzle.md", []byte("Answer: broken\nSpooon\n"), 0644)
	afero.WriteFile(fs, "cat0/22/puzzle.md", []byte("---\nanswers:\n  - pencil\npre:\n unused-field: Spooon\n---\nSpoon?\n"), 0644)
	afero.WriteFile(fs, "cat1/93/puzzle.md", []byte("Answer: no\n\nbody"), 0644)
	afero.WriteFile(fs, "cat1/barney/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "unbroken/1/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "unbroken/1/moo.txt", []byte("Moo."), 0644)
	afero.WriteFile(fs, "unbroken/2/puzzle.md", testMothRfc822, 0644)
	return fs
}
