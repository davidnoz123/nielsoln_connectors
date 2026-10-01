// The import path, so `go run <this>@<sha>` works without a checkout. Same
// reasoning as remote_ai's: a module whose declared path does not match where
// it is fetched from is refused before anything downloads.
module github.com/davidnoz123/nielsoln_connectors/pdf_view_server

// Deliberately LOW rather than matching the toolchain that built it. A higher
// floor makes `go run` fetch a whole toolchain on a machine that already has a
// working one, and nothing here needs anything newer.
go 1.21
