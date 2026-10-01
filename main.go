package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--test-bell" {
		if err := playBell(); err != nil {
			fmt.Fprintf(os.Stderr, "pomodoro: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pomodoro: %v\n", err)
		os.Exit(1)
	}
}
