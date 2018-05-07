package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Award struct {
	when time.Time
	team string
	category string
	points int
}

func ParseAward(s string) (*Award, error) {
	ret := Award{}
	
	parts := strings.SplitN(s, " ", 5)
	if len(parts) < 4 {
		return nil, fmt.Errorf("Malformed award string")
	}
	
	whenEpoch, err := strconv.ParseInt(parts[0], 10, 64)
	if (err != nil) {
		return nil, fmt.Errorf("Malformed timestamp: %s", parts[0])
	}
	ret.when = time.Unix(whenEpoch, 0)
	
	ret.team = parts[1]
	ret.category = parts[2]
	
	points, err := strconv.Atoi(parts[3])
	if (err != nil) {
		return nil, fmt.Errorf("Malformed points: %s", parts[3])
	}
	ret.points = points

	return &ret, nil
}

func (a *Award) String() string {
	return fmt.Sprintf("%d %s %s %d", a.when.Unix(), a.team, a.category, a.points)
}

func pointsLog() []Award {
	var ret []Award

	fn := statePath("points.log")
	f, err := os.Open(fn)
	if err != nil {
		log.Printf("Unable to open %s: %s", fn, err)
		return ret
	}
	defer f.Close()
	
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		cur, err := ParseAward(line)
		if err != nil {
			log.Printf("Skipping malformed award line %s: %s", line, err)
			continue
		}
		ret = append(ret, *cur)
	}
	
	return ret
}

// awardPoints gives points points to team teamid in category category
func awardPoints(teamid string, category string, points int) error {
	fn := fmt.Sprintf("%s-%s-%d", teamid, category, points)
	tmpfn := statePath("points.tmp", fn)
	newfn := statePath("points.new", fn)
	
	contents := fmt.Sprintf("%d %s %s %d\n", time.Now().Unix(), teamid, points)
	
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0644); err != nil {
		return err
	}
	
	if err := os.Rename(tmpfn, newfn); err != nil {
		return err
	}
	
	return nil
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func collectPoints() {
	logf, err := os.OpenFile(statePath("points.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Can't append to points log: %s", err)
		return
	}
	defer logf.Close()
	
	files, err := ioutil.ReadDir(statePath("points.new"))
	if err != nil {
		log.Printf("Error reading packages: %s", err)
	}
	for _, f := range files {
		filename := statePath("points.new", f.Name())
		s, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("Can't read points file %s: %s", filename, err)
			continue
		}
		award, err := ParseAward(string(s))
		if err != nil {
			log.Printf("Can't parse award file %s: %s", filename, err)
			continue
		}
		fmt.Fprintf(logf, "%s\n", award.String())
		log.Print(award.String())
		logf.Sync()
		if err := os.Remove(filename); err != nil {
			log.Printf("Unable to remove %s: %s", filename, err)
		}
	}
}