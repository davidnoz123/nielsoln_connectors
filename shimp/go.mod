// The import path, so `go run <this>@<sha> install` works with no checkout.
// It must match where the code actually lives on GitHub: go refuses with
// "module declares its path as" BEFORE it downloads anything, and that
// failure lands on the user's machine rather than ours.
//
// Its own module, matching the four connectors beside it, which is s1: shimp
// is a connector folder like any other and needed no repo reorganisation.
//
// The `require` block is empty and stays empty. s46 chose a COPIED
// shimpipc.go over a shared module precisely so that every connector keeps
// this property, and the handler is the one place where it matters most: it
// is the trust root, and a dependency tree is the thing that stops a human
// being able to audit it by reading it.
module github.com/davidnoz123/nielsoln_connectors/shimp

go 1.27
