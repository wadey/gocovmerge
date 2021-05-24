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

func mergeProfiles(p *cover.Profile, merge *cover.Profile) {
	if p.Mode != merge.Mode {
		log.Fatalf("cannot merge profiles with different modes")
	}
	// Since the blocks are sorted, we can keep track of where the last block
	// was inserted and only look at the blocks after that as targets for merge
	startIndex := 0
	for _, b := range merge.Blocks {
		startIndex = mergeProfileBlock(p, b, startIndex)
	}
}

func mergeProfileBlock(p *cover.Profile, pb cover.ProfileBlock, startIndex int) int {
	sortFunc := func(i int) bool {
		pi := p.Blocks[i+startIndex]
		return pi.StartLine >= pb.StartLine && (pi.StartLine != pb.StartLine || pi.StartCol >= pb.StartCol)
	}

	i := 0
	if sortFunc(i) != true {
		i = sort.Search(len(p.Blocks)-startIndex, sortFunc)
	}
	i += startIndex
	if i < len(p.Blocks) && p.Blocks[i].StartLine == pb.StartLine && p.Blocks[i].StartCol == pb.StartCol {
		if p.Blocks[i].EndLine != pb.EndLine || p.Blocks[i].EndCol != pb.EndCol {
			log.Fatalf("OVERLAP MERGE: %v %v %v", p.FileName, p.Blocks[i], pb)
		}
		switch p.Mode {
		case "set":
			p.Blocks[i].Count |= pb.Count
		case "count", "atomic":
			p.Blocks[i].Count += pb.Count
		default:
			log.Fatalf("unsupported covermode: '%s'", p.Mode)
		}
	} else {
		if i > 0 {
			pa := p.Blocks[i-1]
			if pa.EndLine >= pb.EndLine && (pa.EndLine != pb.EndLine || pa.EndCol > pb.EndCol) {
				log.Fatalf("OVERLAP BEFORE: %v %v %v", p.FileName, pa, pb)
			}
		}
		if i < len(p.Blocks)-1 {
			pa := p.Blocks[i+1]
			if pa.StartLine <= pb.StartLine && (pa.StartLine != pb.StartLine || pa.StartCol < pb.StartCol) {
				log.Fatalf("OVERLAP AFTER: %v %v %v", p.FileName, pa, pb)
			}
		}
		p.Blocks = append(p.Blocks, cover.ProfileBlock{})
		copy(p.Blocks[i+1:], p.Blocks[i:])
		p.Blocks[i] = pb
	}
	return i + 1
}

func addProfile(profiles []*cover.Profile, p *cover.Profile) []*cover.Profile {
	i := sort.Search(len(profiles), func(i int) bool { return profiles[i].FileName >= p.FileName })
	if i < len(profiles) && profiles[i].FileName == p.FileName {
		mergeProfiles(profiles[i], p)
	} else {
		profiles = append(profiles, nil)
		copy(profiles[i+1:], profiles[i:])
		profiles[i] = p
	}
	return profiles
}

func dumpProfiles(profiles []*cover.Profile, out io.Writer) {
	if len(profiles) == 0 {
		fmt.Println("no lines to print!")
		return
	}

	f, err := os.Create("coverage.out")
	if err != nil {
		fmt.Println("error creating coverage.out: ", err.Error())
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("mode: %s\n", profiles[0].Mode))

	for _, p := range profiles {
		for _, b := range p.Blocks {
			f.WriteString(fmt.Sprintf("%s:%d.%d,%d.%d %d %d\n", p.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt, b.Count))
		}
	}
}

func printCoverage(profiles []*cover.Profile) {
	linesCovered := 0
	linesNotCovered := 0

	for _, p := range profiles {
		for _, b := range p.Blocks {
			if b.Count == 0 {
				linesNotCovered += b.NumStmt
			} else {
				linesCovered += b.NumStmt
			}
		}
	}

	totalLines := linesCovered + linesNotCovered
	coverage := 100.0 * float32(linesCovered) / float32(totalLines)

	fmt.Printf("coverage: %v%%\n", coverage)
}

func main() {
	flag.Parse()

	var merged []*cover.Profile

	if len(flag.Args()) == 0 {
		fmt.Println("usage: go run gocovmerge.go /path/to/coverage/file1 /path/to/coverage/file2 /you/can/have/as/many/files/as/you/want/file3")
		return
	}

	for _, file := range flag.Args() {
		profiles, err := cover.ParseProfiles(file)
		if err != nil {
			log.Fatalf("failed to parse profiles: %v", err)
		}
		for _, p := range profiles {
			merged = addProfile(merged, p)
		}
	}

	dumpProfiles(merged, os.Stdout)
	printCoverage(merged)
}
