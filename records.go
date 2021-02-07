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
)

const (
	fileName = "timewatch-hours.txt"
	scale_up = 16.0 / 15.0
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type trecord struct {
	started time.Time
	project string
	// already scaled
	worked float64
	// negative value, more work until next half hour can be billed
	remaining float64
	// bille
	billed float64
	// not saved
	previous *trecord
	year     int
	week     int
	yearDay  int
	weekDay  time.Weekday
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
			tr.worked, err = strconv.ParseFloat(parts[2], 32)
			if err != nil {
				fmt.Printf("worked part (1st integer) on line %d is wrong", line)
				os.Exit(2)
			}
		}
		if len(parts) >= 4 {
			tr.remaining, err = strconv.ParseFloat(parts[3], 32)
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
			tr.billed, err = strconv.ParseFloat(parts[4], 32)
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
		wbf.WriteString(fmt.Sprintf("%s %s %.0F %.0F %.1F\n", t, r.project, r.worked, r.remaining, r.billed))
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
	var previous *trecord
	tf := "2006-01-02_15:04:05 MST"
	// calculating the worked time in timely fashion
	fmt.Print("calculating times...")
	for _, record := range records {
		if previous != nil {
			d := record.started.Sub(previous.started)
			previous.worked = d.Seconds() * scale_up
			fmt.Printf("%s - %s = %.1F\n", record.started.Format(tf), previous.started.Format(tf), d.Seconds())
		}
		record.year, record.week = record.started.ISOWeek()
		record.yearDay = record.started.YearDay()
		record.weekDay = record.started.Weekday()
		previous = record
	}
	// recalculate projects
	fmt.Print("projects...")
	for _, rec := range records {
		if rec.previous != nil {
			rec.remaining = rec.previous.remaining + rec.worked
			if rec.previous.yearDay == rec.yearDay {
				rec.remaining += 1800 * rec.previous.billed

				rec.previous.billed = 0
			}
			if rec.remaining > 0 {
				billing := math.Ceil(rec.remaining / 1800.0)
				rec.billed = billing
				rec.remaining -= billing * 1800
			} else {
				rec.billed = 0
			}
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

	hours := regBilledHours / 2
	toWork := float64(regWorkDays*8) - hours

	sec2work := ((lastRec.billed+toWork)*3600.0 - lastRec.remaining) / scale_up
	fmt.Printf(" registered work days: %12d\n", regWorkDays)
	fmt.Printf("        worked so far: %12.1F\n", hours)
	fmt.Printf("      work more hours: %12.1F\n", toWork)
	fmt.Printf("      seconds to work: %12F\n", sec2work)

}
