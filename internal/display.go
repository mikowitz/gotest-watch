package internal

import (
	"fmt"
	"strings"
)

func displayPrompt() {
	fmt.Print("> ")
}

func displayCommand(command []string) {
	fmt.Println(strings.Join(command, " "))
}
