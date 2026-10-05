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

	// Click asks for one click on the element the highlight landed in, and
	// it is the only thing in this program that changes a page rather than
	// looking at one.
	//
	// WHY IT EXISTS. An Elementor edit drive opened the right screen and
	// highlighted the right words, and the panel on the left still read
	// "Elements": the text was visible and not editable, because Elementor
	// opens a widget's settings only when the widget is clicked. The drive
	// was arriving one click short of the thing it existed for.
	//
	// WHY IT IS OPT-IN. A click can DO something. On a settings screen it
	// might toggle a control and on a list it might follow a link, so a tool
	// that clicked whatever it found would eventually press something nobody
	// asked it to. Declared per task, in the repo that knows what the screen
	// is, and absent everywhere it has not been thought about.
	//
	// WHAT IT MAY CLICK. The element the mark sits in, or the nearest
	// ancestor an editor would recognise as a widget. Never a selector off
	// the wire, never a coordinate, and never more than once.
	Click bool `json:"click"`

	// An address the human can OPEN if the driven Chrome is not where they
	// expected. Needed because `page` is often view-source:..., and Chrome
	// refuses view-source: opened from a click or a paste from another
	// program: printing it as the fallback hands somebody a string that
	// cannot work, on exactly the rows where they are most lost.
	Fallback string `json:"fallback"`

	// LocatorRE matches the source LINE holding the deepest HTML tag that
	// encloses the change: `<title[\s>]` for a title, the og:locale meta
	// for a locale. The FALLBACK for when the exact string is not found,
	// because a tag is on its line whatever its content says, and the
	// content is the part we are about to change.
	//
	// A pattern and not a line number. The number was tried and cannot be
	// computed off-browser: the manifest is built from an anonymous fetch
	// and this Chrome is signed into WordPress, so it receives a different
	// page. Every locator came out one line short. A pattern is matched
	// against whatever source the browser actually holds.
	LocatorRE string `json:"locator_re"`

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
