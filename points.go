package main

import (
	"fmt"
	"log"
	"os"
)

type Award struct {
	when time.Time,
	team string,
	category string,
	points int,
	comment string
}

func ParseAward(s string) (*Award, error) {
	ret := Award{}
	
	parts := strings.SplitN(s, " ", 5)
	if len(parts) < 4 {
		return nil, Error("Malformed award string")
	}
	
	whenEpoch, err = strconv.Atoi(parts[0])
	if (err != nil) {
		return nil, Errorf("Malformed timestamp: %s", parts[0])
	}
	ret.when = time.Unix(whenEpoch, 0)
	
	ret.team = parts[1]
	ret.category = parts[2]
	
	points, err = strconv.Atoi(parts[3])
	if (err != nil) {
		return nil, Errorf("Malformed points: %s", parts[3])
	}
	
	if len(parts) == 5 {
		ret.comment = parts[4]
	}
	
	return &ret
}

func (a *Award) String() string {
	return fmt.Sprintf("%d %s %s %d %s", a.when.Unix(), a.team, a.category, a.points, a.comment)
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func collectPoints() {
	pointsLog = os.OpenFile(statePath("points.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer pointsLog.Close()
	
	for f := range allfiles(statePath("points.new")) {
		filename := statePath("points.new", f.Name())
		s := ioutil.ReadFile(filename)
		award, err := ParseAward(s)
		if (err != nil) {
			log.Printf("Can't parse award file %s: %s", filename, err)
			continue
		}
		fmt.Fprintf(pointsLog, "%s\n", award.String())
		log.Print(award.String())
		pointsLog.Sync()
		err := os.Remove(filename)
		if (err != nil) {
			log.Printf("Unable to remove %s: %s", filename, err)
		}
	}
}