# LLMDog 🐶

**LLMDog** is a developer-friendly command-line tool that simplifies sharing your code with Large Language Models (LLMs) like Claude, ChatGPT, and others. It allows you to intelligently select files and directories from your project, automatically formats them with proper Markdown, and copies the output directly to your clipboard—ready to paste into any LLM chat interface.

[![asciicast](https://asciinema.org/a/lq2kdE5H1efWxz8296EfZVfHk.svg)](https://asciinema.org/a/lq2kdE5H1efWxz8296EfZVfHk)

## Why LLMDog?

### The Problem
Working with LLMs on code-related tasks often requires sharing multiple files or code snippets from your project. The traditional approach involves:
- Manually copying individual files
- Formatting each file with proper Markdown code blocks
- Ensuring context about file location and project structure
- Repeating this tedious process for each LLM conversation

This is time-consuming, error-prone, and interrupts your workflow.

### The Solution
LLMDog streamlines this entire process with an intuitive interface that lets you:
- Quickly navigate and select exactly the files you need
- Generate properly formatted Markdown output with file structure
- Have everything copied to your clipboard in one step
- Save common file selections as bookmarks for reuse

## When to Use LLMDog

### Code Reviews and Feedback
When seeking feedback on your code, LLMDog helps you present your code in a structured way that preserves context. LLMs can better understand your architecture and provide more valuable feedback when they see both the code and its organization within your project.

### Debugging Assistance
Stuck on a difficult bug? LLMDog makes it easy to share the relevant files without having to individually copy each one. Include unit tests, error logs, and implementation files in a single selection to give LLMs the complete picture they need to help.

### Learning New Codebases
When working with unfamiliar code, use LLMDog to select parts of the codebase you don't understand and ask LLMs for explanations. The tool preserves file relationships and project structure, helping LLMs provide more accurate explanations of how components interact.

### Architecture and Design Discussions
Share your current implementation with an LLM before discussing potential architecture changes. LLMDog helps you include all the relevant components so the LLM can suggest improvements with full context.

### Documentation Generation
Need documentation for your code? Select the files that need documentation and ask an LLM to generate it. The structured output from LLMDog gives the LLM everything it needs to create accurate and comprehensive documentation.

### Onboarding Team Members
Create bookmarks of critical code paths and share them with new team members. They can use these bookmarks with LLMDog to quickly get explanations of core functionality from LLMs.

## Key Features

### 🔖 Bookmarks System (New in 2.0)
Save time with reusable file selections:
- Store commonly used file combinations for different components
- Quickly switch between different aspects of your project
- Share bookmark configurations with team members for consistent LLM interactions
- Access bookmarks with `Ctrl+B`, create with `Ctrl+Shift+B`

### 🔍 Intelligent Search (Enhanced in 2.0)
Find exactly what you need:
- Content search mode finds code by its content, not just filename
- Visual indicators show which files contain your search terms
- Automatically expands relevant directories to show matches
- Search history for quick refinement of searches

### 📂 Smart File Selection
Work efficiently with your project structure:
- Recursive selection of directories with a single click
- Automatic exclusion of irrelevant files using your .gitignore
- Expand/collapse folders to navigate large projects with ease
- Select all visible files with `Ctrl+A` or deselect all with `Ctrl+D`

### 📋 Optimized for LLM Consumption
Generate output that maximizes LLM understanding:
- Includes directory structure for context
- Formats code with language-specific syntax highlighting
- Smart truncation of large files to fit within token limits
- Properly labeled file paths for clear references

## Getting Started

### Installation

#### From Source
```bash
git clone https://github.com/doganarif/LLMDog.git
cd LLMDog
go build -o LLMDog ./cmd/LLMDog
```

#### Homebrew (macOS)
```bash
brew tap doganarif/LLMDog
brew install LLMDog
```

### Basic Workflow

1. **Navigate** to your project directory
2. **Run** LLMDog: `LLMDog`
3. **Select files** using the interactive interface:
   - Navigate with arrow keys (↑/↓)
   - Expand/collapse folders with Space
   - Select files and directories with Tab
   - Filter with /
4. **Press Enter** to generate the output and copy to clipboard
5. **Paste** into your LLM chat
6. **Ask your question** with full code context

### Bookmarks Workflow (New in 2.0)

Bookmarks streamline repetitive tasks:

1. **Creating a bookmark**:
   - Select your files
   - Press `Ctrl+Shift+B`
   - Name your bookmark
   - Optionally add a description

2. **Using bookmarks**:
   - Press `Ctrl+B` to view your saved bookmarks
   - Select the bookmark you need
   - Press Enter to automatically select those files
   - Continue as usual

### Content Search (New in 2.0)

Finding relevant code is easier:

1. Press `/` to start searching
2. Toggle content search with `Ctrl+S`
3. Enter your search term
4. Navigate matches and select what you need

## Keyboard Shortcuts

| Key            | Action                          |
|----------------|----------------------------------|
| ↑/↓            | Navigate items                  |
| Space          | Expand/collapse folder          |
| Tab            | Select/unselect item            |
| /              | Filter items                    |
| Enter          | Confirm selection               |
| Ctrl+A         | Select all visible items        |
| Ctrl+D         | Deselect all items              |
| Ctrl+S         | Toggle content search mode      |
| Ctrl+B         | Open bookmarks menu             |
| Ctrl+Shift+B   | Save selection as bookmark      |
| Ctrl+/         | Toggle preview pane             |
| Esc            | Clear filter/close modal        |
| q              | Quit                            |

## Command-Line Options

```bash
LLMDog [options]
```

- `-h, --help`: Show help message
- `-v, --version`: Show version
- `--about`: Display information about LLMDog

## Configuration

LLMDog stores configuration in `~/.config/llmdog/`:
- `config.json`: General settings
- `bookmarks.json`: Saved bookmarks

## Contributing

Contributions are welcome! Whether it's bug reports, feature requests, or code contributions, please feel free to engage with the project.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
