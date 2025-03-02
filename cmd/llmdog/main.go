package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/doganarif/llmdog/internal/model"
	"github.com/doganarif/llmdog/internal/ui"
)

const (
	version = "2.0.0"
	banner  = `
 _      _      __  __ ____             
| |    | |    |  \/  |  _ \            
| |    | |    | \  / | | | | ___   __ _ 
| |    | |    | |\/| | | | |/ _ \ / _  |
| |____| |____| |  | | |_| | (_) | (_| |
|______|______|_|  |_|____/ \___/ \__, |
                                   __/ |
                                  |___/ 
`
)

func main() {
	// Parse command-line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("llmdog version %s\n", version)
			os.Exit(0)

		case "-h", "--help":
			fmt.Print(getHelpText())
			os.Exit(0)

		case "--about":
			fmt.Print(getAboutText())
			os.Exit(0)
		}
	}

	// Initialize the application
	p := tea.NewProgram(model.New(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal("Error running program:", err)
	}
}

func getHelpText() string {
	helpText := []string{
		ui.EmphasisStyle.Render(banner),
		"llmdog - Prepare files for LLM consumption",
		"",
		ui.EmphasisStyle.Render("USAGE:"),
		"  llmdog [options]",
		"",
		ui.EmphasisStyle.Render("OPTIONS:"),
		"  -h, --help      Show this help message",
		"  -v, --version   Show version",
		"  --about         About llmdog",
		"",
		ui.EmphasisStyle.Render("KEYS:"),
		"  ↑/↓             Navigate items",
		"  Space           Expand/collapse folder",
		"  Tab             Select/unselect item",
		"  /               Filter items",
		"  Ctrl+A          Select all visible items",
		"  Ctrl+D          Deselect all items",
		"  Ctrl+S          Toggle content search mode",
		"  Ctrl+/          Toggle preview pane",
		"  Enter           Confirm selection",
		"  Esc             Clear filter/errors",
		"  q               Quit",
	}

	return strings.Join(helpText, "\n") + "\n"
}

func getAboutText() string {
	aboutText := []string{
		ui.EmphasisStyle.Render(banner),
		fmt.Sprintf("LLMDog v%s", version),
		"",
		"LLMDog is a developer-friendly command-line tool that simplifies sharing your code",
		"with Large Language Models (LLMs) like Claude, ChatGPT, and others.",
		"",
		"It allows you to intelligently select files and directories from your project,",
		"automatically formats them with proper Markdown, and copies the output directly",
		"to your clipboard—ready to paste into any LLM chat interface.",
		"",
		ui.EmphasisStyle.Render("Features:"),
		"• Interactive TUI with file navigation and selection",
		"• Gitignore support to exclude irrelevant files",
		"• Content search for finding relevant code",
		"• Smart truncation for large files",
		"• Markdown-formatted output with syntax highlighting",
		"",
		"Repository: https://github.com/doganarif/llmdog",
		"License: MIT",
	}

	return strings.Join(aboutText, "\n") + "\n"
}
