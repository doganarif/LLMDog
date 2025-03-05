# LLMDog 🐶

**LLMDog** is a developer-friendly command-line tool that simplifies sharing your code with Large Language Models (LLMs) like Claude, ChatGPT, and others. It allows you to intelligently select files and directories from your project, automatically formats them with proper Markdown, and copies the output directly to your clipboard—ready to paste into any LLM chat interface.

<a href="https://www.producthunt.com/posts/llmdog?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-llmdog" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=938141&theme=light&t=1741167251516" alt="LLMDog - Your&#0032;code&#0039;s&#0032;best&#0032;friend&#0032;for&#0032;seamless&#0032;LLM&#0032;conversations | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

## Why Use LLMDog?

- **Streamlined Code Discussions**: Quickly share relevant files with LLMs for code reviews, debugging help, or implementation guidance.
- **Smart File Selection**: Choose only the files that matter for your question, without the noise of your entire codebase.
- **Gitignore Integration**: Automatically excludes files like `node_modules`, build artifacts, and other files in your `.gitignore`.
- **Consistent Formatting**: Ensures your code is properly formatted in Markdown, with syntax highlighting based on file extensions.
- **Context Preservation**: Includes a directory tree to help LLMs understand your project structure.
- **Zero Configuration**: Works out of the box with no setup required.

## Example Use Cases

- Get help fixing a bug by sharing only the relevant files
- Ask for code reviews on specific components of your application
- Request implementation suggestions for new features
- Seek assistance with refactoring specific parts of your codebase
- Generate documentation for your project or specific modules

[![asciicast](https://asciinema.org/a/lq2kdE5H1efWxz8296EfZVfHk.svg)](https://asciinema.org/a/lq2kdE5H1efWxz8296EfZVfHk)

## Features

- **Interactive TUI:** Browse and navigate your files and directories with an intuitive interface.
- **Recursive File & Directory Selection:** Easily select whole directories while automatically handling nested files and skipping Gitignored paths.
- **Gitignore Support:** Automatically respects your `.gitignore` file to exclude irrelevant files.
- **Markdown Output:** Generates a well-formatted Markdown report, complete with a file tree and file contents.
- **Clipboard Integration:** The output is copied directly to your clipboard for quick sharing and use.
- **Cross-Platform:** Built with Go, LLMDog works on macOS, Linux, and Windows.

## Installation

### From Source

Ensure you have [Go](https://golang.org/) installed (version 1.16 or higher is recommended). Then, clone the repository and build the application:

```bash
git clone https://github.com/doganarif/LLMDog.git
cd LLMDog
go build -o LLMDog ./cmd/LLMDog
```

### Homebrew (macOS)

For macOS users, you can install **LLMDog** using Homebrew with my custom tap:

1. Tap the repository:
   ```bash
   brew tap doganarif/LLMDog
   ```

2. Install LLMDog:
   ```bash
   brew install LLMDog
   ```

## Usage

Run **LLMDog** from your terminal:

```bash
./LLMDog [options]
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

## Workflow Example

1. Navigate to your project directory
2. Run `LLMDog`
3. Use the arrow keys to navigate and Tab to select relevant files
4. Press Enter to generate the Markdown output and copy it to your clipboard
5. Paste directly into your favorite LLM chat interface
6. Ask your question about the code you've shared

## Contributing

Contributions are welcome! If you find bugs or have ideas for improvements, please open an issue or submit a pull request. For major changes, please open an issue first to discuss what you would like to change.

## Development

To run and develop **LLMDog**, ensure you have the latest version of Go installed. Clone the repository, make your changes, and submit pull requests. Your contributions help improve the tool for everyone.

## Acknowledgements

- **llmcat**: For the initial inspiration behind this project
- **Bubble Tea & Lip Gloss**: For providing the powerful TUI libraries that make this project possible

## License

This project is licensed under the MIT License. See the LICENSE file for details.
