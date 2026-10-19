// store.go -- where captured conversations go, behind one narrow seam.
//
// # WHY A SEAM AT ALL
//
// The storage question is OPEN and was deferred deliberately. Measured 20 Oct
// 2026, the options and what each costs a recipient:
//
//	CPython's sqlite        FTS5 free, 0 MB marginal, googleish works as-is
//	modernc.org/sqlite      266.8 MB, 25 modules, 1.5M lines of Go
//	mattn/go-sqlite3        10.2 MB, but cgo -- a C compiler on every machine
//	bbolt + bleve           53.2 MB, pure Go, but no SQL so the search is new
//
// ⚠️ Nothing in this file picks one. It exists so the capture half can be
// finished and used while that decision waits, and so the decision does not
// require editing the capture.
//
// # ⚠️ THIS INTERFACE IS UNPROVEN, AND THAT IS NOT MODESTY
//
// It was drawn from what `runLive` and `run` ALREADY DO, so it fits the
// file-handoff store by construction. Whether it is the right seam will only
// be known when a second implementation is written against it, and none is.
//
// That warning is lifted almost verbatim from `patchbucket`'s SqliteStore,
// whose author wrote: "Whether it is the right seam will only be known when a
// second store is written against it, and none is planned. Do not treat it as
// a promise that other stores work." ⓘ The same caution applies here for the
// same reason, and pretending otherwise is how a one-implementation interface
// becomes a constraint nobody chose.
//
// # ⚠️ WHAT A "PYTHON SQLITE" BACKEND CANNOT BE
//
// It cannot shell out to Python. caveats.go declares "no exec" and names
// os/exec as a surface that must stay absent, AGENTS states the direction as
// "Python uses Go; Go does not use Python", and s55 requires every module in
// the closure to live in this repo -- an unpinned interpreter running an
// unpinned script satisfies none of that.
//
// So the Python backend is a FILE HANDOFF: Go writes, Python reads, and the
// two never call each other. That is what handoffStore is.
package main

import "encoding/json"

// Store is where a captured conversation goes.
//
// ⚠️ THREE METHODS, BECAUSE THAT IS WHAT THE CAPTURE USES. No Get, no Delete,
// no query: adding them would be guessing at a second implementation's needs,
// and every unused method is a promise this code cannot keep.
type Store interface {
	// Watermark is the newest update_time already captured, and HOW that was
	// determined. The string is returned because a caller that cannot tell
	// "nothing stored yet" from "this store cannot say" would silently do a
	// full sweep and call it incremental.
	Watermark() (float64, string)

	// Put stores one conversation EXACTLY AS RECEIVED.
	//
	// ⚠️ The raw bytes, uninterpreted. The whole design property of this
	// module is that it does not read fields, and a Store that parsed them
	// would move the interpretation rather than remove it.
	Put(raw json.RawMessage) error

	// Commit finishes the write and reports how many conversations landed.
	// ⚠️ Called even on a failed run, so a partial capture is not lost.
	Commit() (int, error)
}
