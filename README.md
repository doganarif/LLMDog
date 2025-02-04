# llmdog

**llmdog** is a command-line tool designed to help you prepare files for LLM consumption. With an interactive terminal UI built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss), llmdog lets you navigate your file system, select files and directories (with support for Gitignore rules and recursive selection), and generate a Markdown-formatted output of your directory structure and file contents. The final output is automatically copied to your clipboard, streamlining your workflow for LLM-based projects.

> **Inspiration:** This project was inspired by [llmcat](https://github.com/azer/llmcat).

## Features

- **Interactive TUI:** Browse and navigate your files and directories with an intuitive interface.
- **Recursive File & Directory Selection:** Easily select whole directories while automatically handling nested files and skipping Gitignored paths.
- **Gitignore Support:** Automatically respects your `.gitignore` file to exclude irrelevant files.
- **Markdown Output:** Generates a well-formatted Markdown report, complete with a file tree and file contents.
- **Clipboard Integration:** The output is copied directly to your clipboard for quick sharing and use.
- **Cross-Platform:** Built with Go, llmdog works on macOS, Linux, and Windows.

## Installation

### From Source

Ensure you have [Go](https://golang.org/) installed (version 1.16 or higher is recommended). Then, clone the repository and build the application:

```bash
git clone https://github.com/doganarif/llmdog.git
cd llmdog
go build -o llmdog ./cmd/llmdog
```

### Homebrew (macOS)

For macOS users, you can install **llmdog** using Homebrew with my custom tap:

1. Tap the repository:
   ```bash
   brew tap doganarif/llmdog
   ```

2. Install llmdog:
   ```bash
   brew install llmdog
   ```

## Usage

Run **llmdog** from your terminal:

```bash
./llmdog [options]
```

### Command-Line Options

- `-h, --help`: Show the help message
- `-v, --version`: Display the application version

### Interactive TUI Keys

- **↑/↓**: Navigate through list items
- **Space**: Expand or collapse folders
- **Tab**: Select or unselect an item
- **/**: Filter items
- **ctrl+/**: Toggle the preview pane
- **Enter**: Confirm selection and generate the Markdown output (which is also copied to your clipboard)
- **q**: Quit the application

## Contributing

Contributions are welcome! If you find bugs or have ideas for improvements, please open an issue or submit a pull request. For major changes, please open an issue first to discuss what you would like to change.

## Development

To run and develop **llmdog**, ensure you have the latest version of Go installed. Clone the repository, make your changes, and submit pull requests. Your contributions help improve the tool for everyone.

## Acknowledgements

- **llmcat**: For the initial inspiration behind this project
- **Bubble Tea & Lip Gloss**: For providing the powerful TUI libraries that make this project possible

## License

This project is licensed under the MIT License. See the LICENSE file for details.

