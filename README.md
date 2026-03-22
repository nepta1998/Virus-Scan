# Virus Scan (Go)

Go project to scan files (and URLs) using the VirusTotal API. It includes a simple TUI that loads the API key from environment variables, lets you choose what to scan, and fetches the results.

## Requirements

- Go **1.26.1** (see `go.mod`)
- A VirusTotal API key

## Installation

```bash
go mod tidy
```

## Configuration

Set the environment variable before running:

```bash
export VIRUSTOTAL_API_KEY=your_api_key_here
```

> Note: a `FileScanService` is planned and currently in progress (not yet implemented).

## Usage

Run the program:

```bash
go run ./cmd
```

In the TUI, select what you want to scan:

- **Scan file** — pick a local file to upload
- **Scan URL** — enter a URL to analyze

Progress and results are shown in the interface.

## Project structure

```
.
├─ cmd/
│  └─ main.go            # TUI entrypoint (scan file or URL)
├─ internal/
│  ├─ models/
│  │  └─ models.go       # Simple result structures
│  └─ service/
│     ├─ virustotal.go   # VirusTotal client implementation
│     └─ filescan.go     # FileScan service (in progress)
├─ go.mod
└─ go.sum
```
