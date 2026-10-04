// The import path, so `go run <this>@<sha>` works. It must match where the
// code actually lives on GitHub: go refuses with "module declares its path
// as" before it downloads anything, and that failure happens on the user's
// machine rather than ours.
module github.com/davidnoz123/nielsoln_connectors/remote_chrome_cdp

go 1.27
