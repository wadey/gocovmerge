// gocovmerge takes the results from multiple `go test -coverprofile` runs and
// merges them into one profile
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"golang.org/x/tools/cover"
)

func mergeProfileBlock(p *cover.Profile, pb cover.ProfileBlock) {
	var inserted, combined int
	i := sort.Search(len(p.Blocks), func(i int) bool {
		pi := p.Blocks[i]
		return pi.StartLine >= pb.StartLine && (pi.StartLine != pb.StartLine || pi.StartCol >= pb.StartCol)
	})
	if i < len(p.Blocks) && p.Blocks[i].StartLine == pb.StartLine && p.Blocks[i].StartCol == pb.StartCol {
		// TODO validate ends match
		p.Blocks[i].Count += pb.Count
		combined++
	} else {
		// TODO validate doesn't overlap with the adjacent blocks
		p.Blocks = append(p.Blocks, cover.ProfileBlock{})
		copy(p.Blocks[i+1:], p.Blocks[i:])
		p.Blocks[i] = pb
		inserted++
	}
}

func mergeProfile(profiles []*cover.Profile, p *cover.Profile) []*cover.Profile {
	i := sort.Search(len(profiles), func(i int) bool { return profiles[i].FileName >= p.FileName })
	if i < len(profiles) && profiles[i].FileName == p.FileName {
		if profiles[i].Mode != p.Mode {
			log.Fatalf("cannot merge profiles with different modes")
		}
		for _, b := range p.Blocks {
			// TODO we can be smarter if we keep track of where we are while merging
			// that way we don't have to re-do the binary search from the start for each entry
			mergeProfileBlock(profiles[i], b)
		}
		// log.Printf("merged: %v", p.FileName)
	} else {
		profiles = append(profiles, nil)
		copy(profiles[i+1:], profiles[i:])
		profiles[i] = p
	}
	return profiles
}

func dumpProfiles(profiles []*cover.Profile, out io.Writer) {
	if len(profiles) == 0 {
		return
	}
	fmt.Fprintf(out, "mode: %s\n", profiles[0].Mode)
	for _, p := range profiles {
		for _, b := range p.Blocks {
			fmt.Fprintf(out, "%s:%d.%d,%d.%d %d %d\n", p.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt, b.Count)
		}
	}
}

func main() {
	flag.Parse()

	var merged []*cover.Profile

	for _, file := range flag.Args() {
		profiles, err := cover.ParseProfiles(file)
		if err != nil {
			log.Fatalf("failed to parse profiles: %v", err)
		}
		for _, p := range profiles {
			merged = mergeProfile(merged, p)
		}
	}

	dumpProfiles(merged, os.Stdout)
}
