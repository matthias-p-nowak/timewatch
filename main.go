// main.go
package main

import (
	"bufio"
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
	msgChan       = make(chan string, 256)
)

func printHelp() {
	fmt.Println("")
	fmt.Println(" use timewatch [<cmd> [<project>] | [<project]]")
	fmt.Println("   no command or project -> interactive mode")
	fmt.Println("   b{egin} <name>        -> starts the project <name>")
	fmt.Println("   d{elete}              -> deletes the current project")
	fmt.Println("   e{nd}                 -> ends the current project")
	fmt.Println("   h{elp}                -> prints this help")
	fmt.Println("   l{ist}                -> list the work for days and weeks")
	fmt.Println("   p{rojects}            -> list recorded projects")
	fmt.Println("   s{summary}            -> shows the current summary")
	fmt.Println("   <name>                -> starts the named project")
}

func getProject() {
	str := ""
	for true {
		ch, key, err := keyboard.GetSingleKey()
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
		" w - print week\n" +
		" q - quit\n")
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
		for len(msgChan) > 0 {
			fmt.Println(<-msgChan)
		}
		fmt.Printf("-->")
		char, key, err := keyboard.GetSingleKey()
		if err != nil {
			log.Fatal("can't get a single key")
		}
		if key == keyboard.KeyEsc || char == 'q' || char == 'Q' {
			return
		}
		if key == keyboard.KeyEnter {
			printIntHelp()
			continue
		}
		switch char {
		case 'b':
			printProjects()
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
		case 'n':
			printProjects()
			fmt.Printf("enter your new project:")
			scan := bufio.NewScanner(os.Stdin)
			if scan.Scan() {
				prj := scan.Text()
				beginProject(prj)
				recalculate()
				showSummary()
			}
		case 'p':
			printProjects()
		case 's':
			recalculate()
			showSummary()
		case 'w':
			recalculate()
			showWeek()
		default:
			fmt.Printf("option not recognized")
			printIntHelp()
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
	// readRecords()
	if len(os.Args) < 2 {
		if hasKeyboard {
			interact()
			recalculate()
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
		showSummary()
	case strings.HasPrefix("end", cmd):
		endProject()
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
	case strings.HasPrefix("week", cmd):
		recalculate()
		showWeek()
	default:
		beginProject(os.Args[1])
		recalculate()
		showSummary()
	}
	finishRecords()
	for len(msgChan) > 0 {
		fmt.Println(<-msgChan)
	}
}
