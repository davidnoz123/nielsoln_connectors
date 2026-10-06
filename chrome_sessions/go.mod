// The import path, so `go run <this>@<sha>` works. It must match where the
// code actually lives on GitHub: go refuses with "module declares its path
// as" before it downloads anything, and that failure happens on the user's
// machine rather than ours.
//
// Its own module for now, matching the three connectors beside it. Whether
// this repo should instead be ONE module with cmd/ subdirectories is an open
// decision recorded in SHIMP.md, and it is not settled here: one SHA pinning
// a connector AND its shared code is the property the audit story wants, but
// changing the layout of three working connectors is not this increment's
// job.
module github.com/davidnoz123/nielsoln_connectors/chrome_sessions

go 1.27
