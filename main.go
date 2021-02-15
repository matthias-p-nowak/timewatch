// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

var (
	// has a tty attached
	hasKeyboard bool
	// keyboard is Open
	isInteractive bool
)

func OpenKeyboard() {
	if isInteractive {
		return
	}
	if !hasKeyboard {
		log.Fatal("shouldn't open keyboard")
	}
	err := keyboard.Open()
	if err != nil {
		log.Fatal("couldn't open the keyboard for single key presses")
	}
	isInteractive = true
}

func closeKeyboard() {
	if !isInteractive {
		return
	}
	if !hasKeyboard {
		log.Fatal("shouldn't close a non-existing keyboard")
	}
	err := keyboard.Close()
	if err != nil {
		log.Fatal("couldn't close the keyboard")
	}
}

func printHelp() {
	fmt.Println("github.com/matthias-p-nowak/timewatch (2021)")
}

func getProject() {
	str := ""
	for true {
		ch, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("couldn't get the key")
			return
		}
		if key == keyboard.KeyEsc {
			return
		}
		if key == keyboard.KeyEnter {
			return
		}
		str = str + string(ch)
		var prjs []string
		for p, _ := range projectMap {
			if strings.Contains(p, str) {
				prjs = append(prjs, p)
			}
		}
		fmt.Printf("%c\n", ch)
		if len(prjs) > 1 {
			for _, p := range prjs {
				fmt.Printf(" %s,", p)
			}
			fmt.Printf("\nhaving %s -->", str)
			continue
		}
		if len(prjs) == 1 {
			beginProject(prjs[0])
			return
		}
		str = ""
		fmt.Print("nothing found, start anew -->")
	}
}

func printIntHelp() {
	fmt.Print("\n\n     ---- available options -----\n" +
		"following options:\n" +
		" b - begin a project\n" +
		" d - delete current project\n" +
		" e - end current project\n" +
		" l - list weekly bills\n" +
		" n - new project\n" +
		" p - list projects\n" +
		" s - print summary\n" +
		" q - quit\n   -->")
}

func printProjects() {
	i := 0
	fmt.Println("--- list of projects ---")
	for p, _ := range projectMap {
		fmt.Printf("%15s ", p)
		i++
		if i%3 == 0 {
			fmt.Println("")
		}
	}
	fmt.Println("")

}

func interact() {
	printIntHelp()
	for {
		char, key, end := getKey()
		fmt.Printf("%c\n", char)
		if end {
			return
		}
		if key == keyboard.KeyEnter {
			printIntHelp()
			continue
		}
		switch char {
		case 'b':
			getProject()
			recalculate()
			showSummary()
		case 'd':
			deleteCurrent()
		case 'e':
			endProject()
		case 'l':
			recalculate()
			listWork()
		case 'p':
			printProjects()
		case 's':
			recalculate()
			showSummary()
		default:
			fmt.Prinf("option not recognized")
			printHelp()

		}
	}
}

func main() {
	fmt.Println("github.com/matthias-p-nowak/timewatch (2021)")
	err := keyboard.Open()
	if err != nil {
		hasKeyboard = false
	} else {
		hasKeyboard = true
		keyboard.Close()
	}
	readRecords()
	if len(os.Args) < 2 {
		if hasKeyboard {
			interact()
			recalculate()
			saveRecords()
		} else {
			printHelp()
		}
		return
	}
	cmd := strings.ToLower(os.Args[1])
	switch {
	case strings.HasPrefix("begin", cmd):
		if len(os.Args) < 3 {
			printHelp()
			return
		}
		beginProject(os.Args[2])
	case strings.HasPrefix("delete", cmd):
		deleteCurrent()
		recalculate()
		saveRecords()
		showSummary()
	case strings.HasPrefix("end", cmd):
		endProject()
		saveRecords()
		recalculate()
		showSummary()
	case strings.HasPrefix("help", cmd):
		printHelp()
	case strings.HasPrefix("list", cmd):
		recalculate()
		listWork()
	case strings.HasPrefix("projects", cmd):
		recalculate()
		printProjects()
	case strings.HasPrefix("summary", cmd):
		recalculate()
		showSummary()
	default:
		beginProject(os.Args[1])
		recalculate()
		saveRecords()
		showSummary()
	}

}
