// main.go
package main

import (
	"fmt"
	"os"

	"github.com/eiannone/keyboard"
)

func printHelp() {
	fmt.Println("help")
}

func interactive() {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		fmt.Printf("You pressed: rune %q, key %X\r\n", char, key)

		if char == 'Q' {
			break
		}
	}
}

func main() {
	fmt.Println("github.com/matthias-p-nowak/timewatch (2021) started")
	defer fmt.Println("all done")
	readRecords()
	if len(os.Args) < 2 {
		interactive()
		return
	}
	switch os.Args[1] {
	case "b":
		if len(os.Args) < 3 {
			printHelp()
			return
		}
		beginProject(os.Args[2])
	case "d":
		deleteCurrent()
	case "e":
		endProject()
	case "h":
		printHelp()
	case "l":
		listWork()
	default:
		beginProject(os.Args[1])
		recalculate()
		saveRecords()
	}

}
