// records
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	projectMap = make(map[string]*trecord)
	records    []*trecord
	rdFmts     = []string{"2006-01-02_15:04:05", "20O6:01:02:15:04:05"}
)

const (
	fileName = "timewatch-hours.txt"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type trecord struct {
	started   time.Time
	project   string
	worked    int
	remaining int
	billed    int
	previous  *trecord
}

func addRecord(tr *trecord) {
	lp := projectMap[tr.project]
	tr.previous = lp
	projectMap[tr.project] = tr
	records = append(records, tr)
}

func readRecords() {
	projectMap = make(map[string]*trecord)
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fn := path.Join(user.HomeDir, fileName)
	f, err := os.Open(fn)
	if err != nil {
		fmt.Printf("couldn't open %s\n", fn)
		return
	}
	sc := bufio.NewScanner(f)
	line := 0
	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Fatal(err)
	}
	for sc.Scan() {
		line += 1
		parts := strings.Split(sc.Text(), " ")
		if len(parts) < 2 {
			continue
		}
		tr := new(trecord)
		timeStr := parts[0]
		for _, f := range rdFmts {
			timeVal, err := time.ParseInLocation(f, timeStr, loc)
			if err == nil {
				tr.started = timeVal
				break
			}
		}
		tr.project = parts[1]
		if len(parts) >= 3 {
			tr.worked, err = strconv.Atoi(parts[2])
			if err != nil {
				fmt.Printf("worked part (1st integer) on line %d is wrong", line)
				os.Exit(2)
			}
		}
		if len(parts) >= 4 {
			tr.remaining, err = strconv.Atoi(parts[3])
			if err != nil {
				fmt.Printf("remaining part (2nd integer) on line %d is wrong", line)
				os.Exit(2)
				if err != nil {
					fmt.Printf("billed part (3rd integer) on line %d is wrong", line)
					os.Exit(2)
				}
			}
		}
		if len(parts) >= 5 {
			tr.billed, err = strconv.Atoi(parts[4])
		}
		addRecord(tr)
	}

}

func saveRecords() {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fn := path.Join(user.HomeDir, fileName)
	wf, err := os.OpenFile(fn+".new", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("writing to file failed: %s\n", err)
	}
	wbf := bufio.NewWriter(wf)
	for _, r := range records {
		t := r.started.Format(rdFmts[0])
		wbf.WriteString(fmt.Sprintf("%s %s %d %d %d\n", t, r.project, r.worked, r.remaining, r.billed))
	}
	wbf.Flush()
	wf.Close()
	os.Rename(fn+".new", fn)
}

func beginProject(prj string) {
	tr := new(trecord)
	tr.project = prj
	tr.started = time.Now()
	addRecord(tr)
}

func deleteCurrent() {
	log.Fatal("todo")
}

func endProject() {
	log.Fatal("todo")
}

func listWork() {
	log.Fatal("todo")
}

func recalculate() {
	var previous *trecord
	tf := "2006-01-02_15:04:05 MST"
	for _, record := range records {
		if previous != nil {
			d := record.started.Sub(previous.started)
			previous.worked = int(d.Seconds())
			fmt.Printf("%s - %s = %f\n", record.started.Format(tf), previous.started.Format(tf), d.Seconds())
		}
		previous = record
	}
}
