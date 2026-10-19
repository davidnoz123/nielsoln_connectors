// caveats.go -- what this program can reach, declared in source.
//
// Checked by `repocheck`, which fails if an outward reach is not named here.
// Written WITH the module, which is now the habit rather than the exception.
//
// ⚠️ AND THE SUBJECT MATTER IS WHY IT MATTERS. A row patch carries somebody's
// conversation rows: titles, message bodies, timestamps. This program reads
// and writes those files. An auditor reading a link that names it is entitled
// to know it never sends one anywhere, and the honest answer is short.
package main

// Caveats is what an auditor should be told, keyed by the surface reached.
var Caveats = map[string]string{
	"[compress/gzip]": "reads and writes gzip streams, which is what a " +
		"patchbucket patch is. It decompresses into memory and compresses " +
		"from memory; it extracts nothing to disk beyond the single -out " +
		"path, and it never reads an archive the caller did not name.",

	"[os.Open, os.Create]": "reads ONE file, the patch named by -verify, and " +
		"writes ONE, when a rewrite is asked for. It creates no directories, " +
		"removes nothing, and touches no database. ⚠️ IT DOES NOT APPLY A " +
		"PATCH TO ANY STORE. That is deliberate and not an oversight: no Go " +
		"store exists yet, and a program that cannot write to a store cannot " +
		"corrupt one.",

	"[crypto/sha256]": "computes the patch integrity hash over the canonical " +
		"JSON form, excluding the patch_hash field itself. ⚠️ VERIFICATION " +
		"IS MANDATORY AND CANNOT BE SKIPPED BY A FLAG: patchbucket's own " +
		"comment records that validating the hash only 'when present' made " +
		"DELETING the field a cheaper forgery than recomputing it. There is " +
		"no -no-verify here and there must never be one.",
}

// CaveatsNotReached is the other half of the claim, and the more useful one.
//
// ⚠️ An auditor's question is rarely "what does it do" but "could it do X".
var CaveatsNotReached = []string{
	"⚠️ NO NETWORK OF ANY KIND. It neither fetches nor uploads a patch. " +
		"Moving patches between machines is rclone's job, invoked by the " +
		"Python side as a subprocess, and nothing in this program can reach " +
		"a remote -- so a patch it has read cannot be forwarded by it.",
	"no exec. It starts no process and shells out to nothing, rclone included.",
	"no sqlite, no database, and no store of any kind. It stops at files.",
	"no credentials, no tokens, no cookie jar. A patch carries conversation " +
		"rows, not secrets, and this program authenticates to nothing.",
	"no deletion, and no writing outside the single path a caller names.",
	"no syscall, no registry, no COM. Portable Go with no build tags.",
	"⚠️ no interpretation of the rows. It counts them and reports unmodelled " +
		"tables by name, but it reads no field value and transforms nothing: " +
		"the bytes written are the bytes read, re-canonicalised, which is the " +
		"property the hash check proves.",
}

// CaveatSurfaces are the imports and symbols the checker looks for.
//
// The ones that must stay ABSENT are named so their absence is asserted
// rather than nobody noticing they arrived.
var CaveatSurfaces = []string{
	"compress/gzip", "os.Open", "os.Create", "crypto/sha256",
	// must stay absent
	"net/http", "net.Dial", "os/exec", "syscall", "database/sql", "os.Remove",
}
