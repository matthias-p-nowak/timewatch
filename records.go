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

	"github.com/eiannone/keyboard"
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
	if len(parts) < 2 {
		r.project = ""
	} else {
		r.project = strings.Trim(parts[1], " \t")
	}
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
	records = append(records, tr)
	if tr.project != "" {
		projectMap[tr.project] = tr
	}
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
	defer f.Close()
	sc := bufio.NewScanner(f)
	line := 0
	var prev *trecord
	for sc.Scan() {
		line += 1
		txt := strings.Trim(sc.Text(), " \t")
		if strings.HasPrefix(txt, "#") {
			continue
		}
		parts := strings.Split(txt, " ")
		if len(parts) < 1 {
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
	var prev *trecord
	for _, r := range records {
		if prev != nil && prev.week != r.week {
			wbf.WriteString("\n")
		}
		r.save(wbf)
		prev = r
	}
	wbf.Flush()
	wf.Close()
	err = os.Rename(fn+".new", fn)
	if err != nil {
		fmt.Println("trying the hard way by removing the old file")
		err = os.Remove(fn)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Rename(fn+".new", fn)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func beginProject(prj string) {
	tr := new(trecord)
	tr.project = prj
	tr.started = time.Now()
	addRecord(tr)
	// tf := "2006-01-02_15:04:05"
	fmt.Printf("      started project: %12s\n", prj)
}

func deleteCurrent() {
	tf := "2006-01-02_15:04:05"
	recs := len(records)
	if recs == 0 {
		return
	}
	rec := records[recs-1]
	if hasKeyboard {
		fmt.Printf("Delete record project='%s' started at %s, \n  (yes/No) -->", rec.project, rec.started.Format(tf))
		ch, _, _ := keyboard.GetSingleKey()
		if ch == 'y' || ch == 'Y' {
			records = records[:recs-1]
			fmt.Println("record deleted")
		}
	} else {
		records = records[:recs-1]
	}
}

func endProject() {
	r := new(trecord)
	r.project = ""
	r.started = time.Now()
	addRecord(r)
	fmt.Println("empty record written")
}

func listWork() {
	tf := "2006-01-02"
	cntBills := len(bills)
	for i := cntBills - 1; i >= 0; i-- {
		bill := bills[i]
		fmt.Printf("Week %d/%02d\n", bill.year, bill.week)
		for d := 0; d < 7; d++ {
			wd := (d + 1) % 7
			recs := bill.billed[wd]
			if len(recs) > 0 {
				rec := recs[0]
				fmt.Printf("   %s %s\n", rec.weekDay.String(), rec.started.Format(tf))
				for _, rec = range recs {
					fmt.Printf("%15s: %5.1F\n", rec.project, rec.billed)
				}
			}
		}
		if hasKeyboard && i > 0 {
			char, key, err := keyboard.GetSingleKey()
			if err != nil {
				log.Fatal("failed to get single key")
			}
			if key == keyboard.KeyEsc || char == 'q' || char == 'Q' {
				fmt.Println("back to previous menu")
				return
			}
		}
	}
}

func showWeek() {
	l := len(bills)
	if l < 1 {
		return
	}
	lastBill := bills[l-1]
	m := make(map[string][7]float64)
	for _, d := range lastBill.billed {
		for _, e := range d {
			pl := m[e.project]
			wd := int(e.weekDay)
			wd = (wd + 6) % 7
			pl[wd] = e.billed
			m[e.project] = pl
		}
	}
	fmt.Println("                 Mon  Tue  Wed  Thu  Fri  Sat  Sun")
	for name, val := range m {
		fmt.Printf("%15s", name)
		for _, d := range val {
			if d > 0 {
				fmt.Printf("%5.1f", d)
			} else {
				fmt.Print("     ")
			}
		}
		fmt.Println("")
	}
}

func recalculate() {
	// tf := "2006-01-02_15:04:05 MST"
	// calculating the worked time in timely fashion
	fmt.Print("\ncalc times...\r")
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
	fmt.Print("calc projects...\r")
	for _, rec := range records {
		if len(rec.project) == 0 {
			rec.billed = 0
			rec.remaining = 0
			continue
		}
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
	fmt.Print("calc weeks...   \r")
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
	fmt.Print("                          \r")
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
	wdToday := -1
	if len(lastRec.project) > 0 {
		wdToday = int(lastRec.weekDay)
	}
	for i := 0; i < 7; i++ {
		if len(lastBill.billed[i]) > 0 {
			regWorkDays += 1
		} else if i == wdToday {
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
	// fmt.Printf("      seconds to work: %12.0F\n", sec2work)
	now := time.Now()
	workUntil := now.Add(time.Duration(sec2work/scale_up) * time.Second)
	tf := "15:04:05"
	fmt.Printf("           work until: %12s\n", workUntil.Format(tf))
}
