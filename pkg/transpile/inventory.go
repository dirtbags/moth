package transpile

import (
	"log"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

// Inventory maps category names to lists of point values.
type Inventory map[string][]int

// FsInventory returns a mapping of category names to puzzle point values.
func FsInventory(fs afero.Fs) (Inventory, error) {
	dirEnts, err := afero.ReadDir(fs, "")
	if err != nil {
		log.Print(err)
		return nil, err
	}

	inv := make(Inventory)
	for _, ent := range dirEnts {
		if strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		if ent.IsDir() {
			name := ent.Name()
			c := NewFsCategory(fs, name)
			puzzles, err := c.Inventory()
			if err != nil {
				log.Printf("Inventory: %s: %s", name, err)
				continue
			}
			sort.Ints(puzzles)
			inv[name] = puzzles
		}
	}

	return inv, nil
}
