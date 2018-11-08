package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Welcome!")
	for {
		var flag string
		fmt.Scanln(&flag)
		switch flag {
		case "out":
			fmt.Println("Outputing...")
		case "err":
			fmt.Println("error!")
		case "exit(0)":
			os.Exit(0)
		case "exit(1)":
			os.Exit(1)
		case "":
			os.Exit(-1)
		default:
			fmt.Println("{", flag, "}")
		}
	}
}
