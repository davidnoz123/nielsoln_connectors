// manifest.go -- the list of places this program is allowed to go.
//
// THE WIRE CARRIES AN ID, NEVER A URL.
//
// This is the whole security model and it is worth being blunt about why. A
// link in a document is clicked by a browser, and any OTHER page open in
// that browser can also fetch 127.0.0.1 URLs. If the request could name a
// destination, any website could use this program to drive the user's
// browser anywhere and read what came back. Because the request names an id
// and the destinations come from a file on disk, the worst a hostile page
// can do is show the user a page they were already going to be shown.
//
// The token raises the bar further, but the vocabulary is what makes the
// program safe rather than merely inconvenient to misuse.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Task struct {
	// Page is what we open. For a source view this is the page whose HTML
	// is being inspected, NOT the tool that displays it: those stopped
	// being the same thing the moment a schema validator turned out to show
	// the markup better than view-source does.
	Page string `json:"page"`

	// Find is the exact string to highlight. Precision matters more than it
	// looks: on the first site this was used against, "http://amazeme.co.nz"
	// matched four places and only the first was the one being changed.
	Find string `json:"find"`

	// Source asks for view-source: rather than the rendered page. True for
	// the changes that leave no mark on what a visitor sees, which is most
	// of what this program exists to show.
	Source bool `json:"source"`

	// An address the human can OPEN if the driven Chrome is not where they
	// expected. Needed because `page` is often view-source:..., and Chrome
	// refuses view-source: opened from a click or a paste from another
	// program: printing it as the fallback hands somebody a string that
	// cannot work, on exactly the rows where they are most lost.
	Fallback string `json:"fallback"`

	// When the Find string should be present: "now", "after" or "browser".
	//
	// "after" means the anchor describes the state we are MOVING TO, so
	// absent is the correct answer today and the search should give up
	// quickly and say so. Ten of fifteen rows are "after", and before this
	// each of them cost a full budget to report a known absence.
	Expect string `json:"expect"`

	// What is shown to the human. One line.
	What string `json:"what"`
}

type Manifest struct {
	Tasks map[string]Task `json:"tasks"`
}

func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("the manifest is not valid JSON: %w", err)
	}
	if len(m.Tasks) == 0 {
		return nil, fmt.Errorf("%s lists no tasks, so every request would be "+
			"refused. Refusing to start rather than serving nothing and "+
			"looking healthy", path)
	}
	// Checked here rather than at request time: a manifest with a blank page
	// is a mistake somebody made while editing, and the moment to say so is
	// at startup, not halfway through a call.
	var bad []string
	for id, t := range m.Tasks {
		if strings.TrimSpace(t.Page) == "" {
			bad = append(bad, id)
		}
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		return nil, fmt.Errorf("these tasks have no page: %s",
			strings.Join(bad, ", "))
	}
	return &m, nil
}

// Lookup is case-insensitive in BOTH directions.
//
// ⚠️ The first version upper-cased the id before looking it up, which was
// invisible while the keys were task ids like "E11" and broke every one of
// them the moment the keys became row ids like "tl-e3-02". The server
// cheerfully listed fifteen tasks and then answered "unknown task" for all
// fifteen, which is the shape of bug that survives a startup log looking
// perfectly healthy.
func (m *Manifest) Lookup(id string) (Task, bool) {
	want := strings.ToLower(strings.TrimSpace(id))
	for key, task := range m.Tasks {
		if strings.ToLower(strings.TrimSpace(key)) == want {
			return task, true
		}
	}
	return Task{}, false
}

func (m *Manifest) IDs() []string {
	out := make([]string, 0, len(m.Tasks))
	for id := range m.Tasks {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
