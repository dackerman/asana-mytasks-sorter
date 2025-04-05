# Asana Tasks Sorter

![Asana Logo](https://assets.asana.biz/m/5d807fb5d4e2a5ff/original/asana-logo-freelogodesign.png)

## 📋 Overview

Asana Tasks Sorter is a sleek, Go-powered CLI tool that pulls your Asana tasks and displays them in a clean, organized format. It's like having your Asana dashboard in your terminal - perfect for productivity-obsessed developers who want to check their tasks without leaving their command line.

Vibe-coded from scratch with ✨good vibes only✨, this tool connects to the Asana API to fetch your tasks, respecting their project sections, due dates, and notes - all in a minimalist display that sparks joy.

## ✨ Features

- **Section-Based Organization**: Tasks are displayed in their Asana sections, making it easy to see your "Focus Tasks," "High Priority / Today," and other custom sections
- **Due Dates**: Shows due dates alongside each task
- **Detailed Notes**: Includes task notes with proper indentation
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

Example output:

```
Logged in as: David Ackerman
Using workspace: Ackerman Household

My Asana Tasks:
==============

## 📌 Focus tasks (1 tasks)
1. testing 123 task (2025-01-31)

## High Priority / Today (2 tasks)
1. Take out trash & recycling (2025-02-22)
2. Turn on security cameras (2025-02-22)

## Tasks (2 tasks)
1. Write out 2024 priorities + plan (2024-04-06)
     High level project task so we don't lose track of this. It needs to be broken down into concrete next steps.

2. [Async Review] Quarterly spending & subscriptions review (2024-03-05)
     Look at expenses for the past quarter and compare to prior quarter.
     
Found 5 tasks in 3 sections
```

## 🧰 Technical Details

This project uses:

- **Go Modules**: For package management
- **HTTP Client Abstraction**: For clean API communication
- **Custom JSON Parsing**: For handling Asana's date formats
- **Snapshot Testing**: Record & replay HTTP interactions for testing without an API token
- **Structured Logging**: For clean, configurable output

## 📊 Project Structure

```
.
├── go.mod              # Go module definition
├── main.go             # Main application entry point
├── main_test.go        # Snapshot tests
├── internal/           # Internal packages
│   ├── asana/          # Asana API client
│   │   └── client.go   # API client implementation
│   └── testing/        # Testing utilities
│       └── snapshot.go # HTTP snapshot recorder/player
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

## 📄 License

MIT

---

Made with ☕ and good vibes by David Ackerman.