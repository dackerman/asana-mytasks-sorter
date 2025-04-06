# Asana Tasks Sorter

![Cute Mascot](assets/robot.png)

## 📋 Overview

Asana Tasks Sorter is a sleek, Go-powered CLI tool that pulls your Asana tasks and displays them in a clean, organized format. It's like having your Asana dashboard in your terminal - perfect for productivity-obsessed developers who want to check their tasks without leaving their command line.

Vibe-coded from scratch with ✨good vibes only✨, this tool connects to the Asana API to fetch your tasks, respecting their project sections and due dates - all in a minimalist display that sparks joy.

## ✨ Features

- **Due Date Categorization**: Tasks are automatically categorized by due date into "Overdue", "Due Today", "Due This Week", and "Due Later" sections
- **Automatic Task Organization**: Tasks are moved to the appropriate sections in Asana based on their due dates
- **Customizable Categories**: Customize category names through a simple JSON configuration file
- **Ignored Sections**: Specify sections to ignore so tasks in those sections don't get moved
- **Section Management**: Automatically creates required sections if they don't exist
- **Due Dates**: Shows due dates alongside each task
- **Workspace Selection**: Connects to your Asana workspace
- **Zero Dependencies**: Just Go standard library - no external dependencies required!
- **Test Coverage**: Includes snapshot testing for reliable, offline testing

## 🚀 Installation

1. Clone this repository:
```bash
git clone https://github.com/dackerman/asana-tasks-sorter.git
cd asana-tasks-sorter
```

2. Build the binary:
```bash
go build
```

3. Set up your Asana API token:
```bash
export ASANA_ACCESS_TOKEN="your_asana_personal_access_token"
```
   (Get your token from [Asana Developer Console](https://app.asana.com/0/developer-console))

## 🖥️ Usage

Run the tool:

```bash
./asana-tasks-sorter
```

Options:

```bash
# Use a custom section configuration file
./asana-tasks-sorter -config path/to/your/config.json

# Preview categorization without moving tasks in Asana
./asana-tasks-sorter -dry-run

# Specify a custom timeout for API operations (default is 30 seconds)
./asana-tasks-sorter -timeout 60s
```

### Configuration File

You can customize the section names by creating a JSON file with the following structure:

```json
{
  "overdue": "Overdue",
  "due_today": "Due today",
  "due_this_week": "Due within the next 7 days",
  "due_later": "Due later",
  "no_date": "Recently assigned",
  "ignored_sections": ["Doing Now", "Waiting For"]
}
```

The `ignored_sections` field is optional and allows you to specify sections that should not have their tasks moved or be moved to. Tasks in these sections will stay where they are.

Example output:

```
Logged in as: David Ackerman
Using workspace: Ackerman Household

Moving task 'Write out 2024 priorities + plan' to section: Overdue
Moving task 'Take out trash & recycling' to section: Due today
Moving task 'Turn on security cameras' to section: Due today
Moving task 'testing 123 task' to section: Due within the next 7 days
Moving task 'Quarterly spending & subscriptions review' to section: Due later

Moved 5 tasks to their appropriate sections

My Asana Tasks:
==============

## Overdue (1 tasks)
1. Write out 2024 priorities + plan (2024-04-06) - High level project task so we don't lose track...

## Due today (2 tasks)
1. Take out trash & recycling (2025-02-22)
2. Turn on security cameras (2025-02-22)

## Due within the next 7 days (1 tasks)
1. testing 123 task (2025-01-31)

## Due later (1 tasks)
1. [Async Review] Quarterly spending & subscriptions review (2024-03-05) - Look at expenses for the past quarter...

Found 5 tasks in 4 sections
```

## 🧰 Technical Details

This project uses:

- **Go Modules**: For package management
- **Interface-Based Design**: For testability and decoupling
- **Context Support**: For proper timeout and cancellation handling
- **HTTP Client Abstraction**: For clean API communication
- **Custom JSON Parsing**: For handling Asana's date formats
- **Snapshot Testing**: Record & replay HTTP interactions for testing without an API token
- **Proper Error Handling**: Using Go 1.13+ error wrapping

## 📊 Project Structure

```
.
├── go.mod              # Go module definition
├── main.go             # Main application entry point (CLI handling)
├── main_test.go        # Integration tests
├── task_moves_test.go  # Unit tests for task sorting logic
├── sections_config.json # Custom section names configuration
├── internal/           # Internal packages
│   ├── asana/          # Asana API client
│   │   ├── client.go   # API client implementation
│   │   └── interface.go # API interface definition
│   ├── config/         # Configuration handling
│   │   └── loader.go   # Configuration loading logic
│   ├── core/           # Core business logic
│   │   ├── config.go   # Domain configuration types
│   │   └── tasks.go    # Task categorization and management
│   ├── testing/        # Testing utilities
│   │   └── snapshot.go # HTTP snapshot recorder/player
│   └── ui/             # User interface components
│       └── display.go  # Task display formatting
└── snapshots/          # Recorded API interactions for tests
```

## 🧪 Testing

Run the tests (using recorded snapshots):

```bash
go test -v
```

Record new snapshots (requires valid ASANA_ACCESS_TOKEN):

```bash
RECORD=true go test -v
```

## 🔮 The Vibe-Coding Journey

This project was vibe-coded from scratch in a single coding session - flowing from idea to implementation with minimal friction. Instead of overthinking architecture or getting caught in analysis paralysis, the code emerged organically through iterative refinement.

The development process followed a "minimal viable slice" approach, starting with a simple script to fetch all tasks, then gradually adding section organization, date formatting, and proper error handling. Each feature was added when the vibe called for it, resulting in clean, focused code that does exactly what it needs to do.

## 📝 Todo

- Add filters for completed tasks
- Support multiple workspaces
- Add colorful output
- Implement interactive mode with task completion

## 🚢 Releasing

To create a new release:

1. Create and push a new tag:
   ```bash
   git tag -a v0.1.0 -m "First release"
   git push origin v0.1.0
   ```

2. GitHub Actions will automatically build and publish the release using GoReleaser.

## 📄 License

MIT

---

Made with ☕ and good vibes by David Ackerman.