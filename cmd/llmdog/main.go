package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/doganarif/llmdog/internal/model"
)

const version = "1.0.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("llmdog version %s\n", version)
			os.Exit(0)
		case "-h", "--help":
			fmt.Println(`llmdog - Prepare files for LLM consumption

Usage: llmdog [options]

Options:
    -h, --help     Show this help message
    -v, --version  Show version

Keys:
    ↑/↓           Navigate items
    Tab           Select/unselect item
    /             Filter items
    ctrl+/        Toggle preview
    Enter         Confirm selection
    q             Quit`)
			os.Exit(0)
		}
	}

	p := tea.NewProgram(model.New(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
