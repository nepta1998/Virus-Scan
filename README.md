# Virus Scan (Go)

Go project to scan files (and URLs) using the VirusTotal API. It includes a simple service that loads the API key from `.env`, uploads a file for analysis, and fetches the results.

## Requirements

- Go **1.26.1** (see `go.mod`)
- A VirusTotal API key

## Installation

```bash
go mod tidy
```

## Configuration

Create a `.env` file at the project root with:

```env
VIRUSTOTAL_API_KEY=your_api_key_here
```

> Note: there is also a `FileScanService` that expects `FILESCAN_API_KEY`, but it is a placeholder (not implemented).

## Usage

The current example lives in `cmd/main.go` and scans a local file. Update the file path before running:

```go
file, err := os.Open("/path/to/your/file")
```

Run the program:

```bash
go run ./cmd
```

Upload progress prints to the console and then the analysis is fetched.

## Project structure

```
.
├─ cmd/
│  └─ main.go            # Example usage of the VirusTotal service
├─ internal/
│  ├─ models/
│  │  └─ models.go       # Simple result structures
│  └─ service/
│     ├─ virustotal.go   # VirusTotal client implementation
│     └─ filescan.go     # Placeholder for another service (WIP)
├─ .env                  # Environment variables (do not commit)
├─ go.mod
└─ go.sum
```

## Notes

- This is a basic example. For production, consider:
  - Better error handling and retries.
  - Parameterizing the file path via CLI or flags.
  - Persisting and/or presenting analysis results in a friendly way.
