// records
package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
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
	bills      []*bill
	loc        *time.Location
)

const (
	fileName = "timewatch-hours.txt"
	scale_up = 16.0 / 15.0
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err error
	loc, err = time.LoadLocation("Local")
	if err != nil {
		log.Fatal(err)
	}
}

type trecord struct {
	started time.Time
	project string
	// already scaled
	worked float64
	// aggregated: previous worked - previous billed*3600 + this.worked
	remaining float64
	// billed
	billed float64
	// not saved
	previous *trecord
	year     int
	week     int
	yearDay  int
	weekDay  time.Weekday
}

func readRecord(parts []string, line int) (r *trecord) {
	r = new(trecord)
	timeStr := parts[0]
	for _, f := range rdFmts {
		timeVal, err := time.ParseInLocation(f, timeStr, loc)
		if err == nil {
			r.started = timeVal
			break
		}
	}
	r.project = parts[1]
	if len(parts) >= 3 {
		rem, err := strconv.ParseFloat(parts[2], 32)
		r.remaining = rem
		if err != nil {
			fmt.Printf("remaining part (1rst integer) on line %d is wrong", line)
			os.Exit(2)
		}
	}
	if len(parts) >= 4 {
		billed, err := strconv.ParseFloat(parts[3], 32)
		if err != nil {
			fmt.Printf("remaining part (2nd number) on line %d is wrong", line)
			os.Exit(2)
		}
		r.billed = billed
	}
	return
}

func (r *trecord) save(wbf *bufio.Writer) {
	t := r.started.Format(rdFmts[0])
	wbf.WriteString(fmt.Sprintf("%s %s %.0F %.1F\n", t, r.project, r.remaining, r.billed))
}

type bill struct {
	year   int
	week   int
	billed [7][]*trecord
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
	var prev *trecord
	for sc.Scan() {
		line += 1
		parts := strings.Split(sc.Text(), " ")
		if len(parts) < 2 {
			continue
		}
		rec := readRecord(parts, line)
		if prev != nil && rec.project == prev.project {
			continue
		}
		addRecord(rec)
		prev = rec
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
		r.save(wbf)
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
	tf := "2006-01-02_15:04:05"
	fmt.Printf("started project: %10s at %s\n", prj, tr.started.Format(tf))
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
	// tf := "2006-01-02_15:04:05 MST"
	// calculating the worked time in timely fashion
	fmt.Print("calculating times...")
	recs := len(records)
	ended := time.Now()
	for i := recs - 1; i >= 0; i-- {
		rec := records[i]
		d := ended.Sub(rec.started)
		rec.worked = math.Ceil(d.Seconds() * scale_up)
		rec.year, rec.week = rec.started.ISOWeek()
		rec.yearDay = rec.started.YearDay()
		rec.weekDay = rec.started.Weekday()
		ended = rec.started
	}
	// recalculate projects
	ended = time.Now()
	fmt.Print("projects...")
	for _, rec := range records {
		if rec.previous != nil {
			// previous == previous of the same project
			previous := rec.previous
			if previous.yearDay == rec.yearDay {
				previous.billed = 0
			}
			rec.remaining = previous.remaining - previous.billed*3600.0 + rec.worked
		} else {
			// no previous record
			if ended.Sub(rec.started).Hours() < 168 {
				// the last record
				rec.remaining = rec.worked
			}
		}
		// calculating billed
		if rec.remaining > 0 {
			billedHH := math.Ceil(rec.remaining / 1800.0)
			rec.billed = billedHH / 2.0
		} else {
			rec.billed = 0
		}
	}
	fmt.Println("weeks")
	bills = nil
	var lastBill *bill
	for _, rec := range records {
		if rec.billed <= 0 {
			continue
		}
		if lastBill == nil || lastBill.week != rec.week || lastBill.year != rec.year {
			lastBill = new(bill)
			lastBill.year = rec.year
			lastBill.week = rec.week
			bills = append(bills, lastBill)
		}
		var wd int
		wd = int(rec.weekDay)
		lastBill.billed[wd] = append(lastBill.billed[wd], rec)
	}
}

func showSummary() {
	if len(records) == 0 {
		return
	}
	lastRec := records[len(records)-1]
	fmt.Printf("      ----- Summary -----\n")
	fmt.Printf("    current year/week: %9d/%02d\n", lastRec.year, lastRec.week)
	if len(lastRec.project) > 0 {
		fmt.Printf("      current project: %12s\n", lastRec.project)
	}
	if len(bills) == 0 {
		return
	}
	lastBill := bills[len(bills)-1]
	regWorkDays := 0
	regBilledHours := 0.0
	for i := 0; i < 7; i++ {
		if len(lastBill.billed[i]) > 0 {
			regWorkDays += 1
		}
		for _, rec := range lastBill.billed[i] {
			regBilledHours += rec.billed
		}
	}

	hours := regBilledHours
	toWork := float64(regWorkDays*8) - hours
	// is toWork - (lastRec.remaining - lastRec.billed*3600)
	sec2work := (toWork+lastRec.billed)*3600.0 - lastRec.remaining
	fmt.Printf(" registered work days: %12d\n", regWorkDays)
	fmt.Printf("        worked so far: %12.1F\n", hours)
	fmt.Printf("      work more hours: %12.1F\n", toWork)
	fmt.Printf("      seconds to work: %12.0F\n", sec2work)
	now := time.Now()
	workUntil := now.Add(time.Duration(sec2work/scale_up) * time.Second)
	tf := "15:04:05"
	fmt.Printf("           work until: %12s\n", workUntil.Format(tf))
}
