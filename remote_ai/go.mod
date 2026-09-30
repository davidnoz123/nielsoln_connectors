// The import path, so `go install <this>@latest` works. It was
// `nielsoln-bridge`, which builds fine locally and cannot be fetched:
// go refuses with "module declares its path as" before it downloads
// anything. Nothing here imports it, so the name was free to change.
module github.com/davidnoz123/nielsoln_connectors/remote_ai

go 1.27
