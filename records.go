// records
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"time"
)

var (
	records = make(map[string]*trecord)
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

func readRecords() {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fn := path.Join(user.HomeDir, "timewatch-hours.txt")
	f, err := os.Open(fn)
	if err != nil {
		fmt.Printf("couldn't open %s\n", fn)
		return
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), " ")
		if len(parts) > 2 {
			log.Fatal("todo")

		}
	}
	log.Fatal("todo")
}

func beginProject(prj string) {

	log.Fatal("todo")
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
