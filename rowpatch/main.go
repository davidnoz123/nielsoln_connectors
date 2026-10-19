// rowpatch -- read, verify and write patchbucket row patches from Go.
//
//	go run github.com/davidnoz123/nielsoln_connectors/rowpatch@<sha> \
//	    -verify <patch.json.gz>
//
// # WHY THIS EXISTS
//
// The goal is that someone handed a shimp:// link gets their own ChatGPT
// capture, their own store and the ability to exchange patches -- without
// needing a private Python repo. Patch exchange is the piece that lets a Go
// peer and a Python peer replicate to each other, and it turns out to need no
// sqlite at all: the wire format is gzip'd canonical JSON carrying rows.
//
// ⚠️ WHAT IT DOES NOT DO, SO NOBODY READS MORE INTO IT. It does not apply a
// patch to any store, because no Go store exists yet. It reads, verifies,
// describes and writes. Applying is the next slice and it needs somewhere to
// apply to.
//
// # THE PART THAT WAS NOT OBVIOUS
//
// patchbucket refuses any patch whose sha256 does not match, computed over
// `json.dumps(patch, ensure_ascii=False, sort_keys=True)`. Reproducing those
// bytes from Go is four rules deep and the obvious implementation scores ZERO.
// canon.go has the rules and probes/README.md has the measurements.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

func main() {
	verify := flag.String("verify", "",
		"a patch .json.gz to read, verify and describe")
	roundtrip := flag.Bool("roundtrip", false,
		"with -verify: also re-canonicalise and require byte-identical output")
	jsonOut := flag.Bool("json", false, "machine-readable output")
	flag.Parse()

	if *verify == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr,
			"\n-verify is required. Nothing here guesses at a default path: a "+
				"patch is named by the caller.")
		os.Exit(2)
	}

	p, err := ReadPatch(*verify)
	if err != nil {
		// ⚠️ The reason is the useful part and it is verbatim from the
		// refusal. "missing_patch_hash" and "invalid_patch_hash" mean very
		// different things: one is a patch from the OLD hand-rolled log, the
		// other is corruption or a canonicalisation bug.
		fmt.Fprintf(os.Stderr, "REFUSED %s\n  %v\n", *verify, err)
		os.Exit(1)
	}

	rows := p.Rows()
	tables := make([]string, 0, len(rows))
	for t := range rows {
		tables = append(tables, t)
	}
	sort.Strings(tables)

	if *jsonOut {
		out := map[string]any{
			"ok":             true,
			"path":           *verify,
			"machine_id":     p.Field("machine_id"),
			"seq":            p.Field("seq"),
			"exported_at":    p.Field("exported_at"),
			"description":    p.Field("description"),
			"patch_hash":     p.Field("patch_hash"),
			"rows":           rows,
			"total_rows":     p.TotalRows(),
			"unknown_tables": p.UnknownTables(),
		}
		if *roundtrip {
			same, _, err := p.Reproduce()
			out["byte_identical"] = same && err == nil
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}

	fmt.Printf("VERIFIED %s\n", *verify)
	fmt.Printf("  machine_id   %s\n", p.Field("machine_id"))
	fmt.Printf("  seq          %s\n", p.Field("seq"))
	fmt.Printf("  exported_at  %s\n", p.Field("exported_at"))
	if d := p.Field("description"); d != "" {
		fmt.Printf("  description  %s\n", d)
	}
	fmt.Printf("  patch_hash   %s\n", p.Field("patch_hash"))
	fmt.Printf("  rows         %d in total\n", p.TotalRows())
	for _, t := range tables {
		fmt.Printf("    %-18s %d\n", t, rows[t])
	}
	if u := p.UnknownTables(); len(u) > 0 {
		// ⚠️ Not an error. The hash covers these and they are written back
		// verbatim, but a peer replicating rows it cannot describe should
		// know it is doing so.
		fmt.Printf("  ⚠️ TABLES THIS BUILD DOES NOT MODEL: %v\n", u)
		fmt.Println("     The hash covers them and they survive a rewrite, " +
			"but this build cannot describe their rows.")
	}

	if *roundtrip {
		same, got, err := p.Reproduce()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  roundtrip FAILED: %v\n", err)
			os.Exit(1)
		}
		if !same {
			// ⚠️ A hash that verified while the bytes differ would mean the
			// hash was computed over something other than the file contents.
			fmt.Fprintf(os.Stderr,
				"  ⚠️ BYTES DIFFER despite the hash verifying.\n"+
					"     on disk: %d bytes\n     ours   : %d bytes\n"+
					"     This means canonical() and the hash disagree, which "+
					"is a bug here, not a bad patch.\n",
				len(p.Raw), len(got))
			os.Exit(1)
		}
		fmt.Println("  ✅ byte-identical: Go's canonical form equals the " +
			"bytes Python wrote")
	}
}
