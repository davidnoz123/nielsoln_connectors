# remote_ai — the connector

A small program you run on your own computer so that an AI running somewhere
else can read files **in one folder you choose** — without those files being
uploaded anywhere.

It is the participant's half of a tool built for library workshops, where
people want to ask an AI about their own documents and would rather not hand
those documents to a website.

    AI in a browser --> a server --> this program --> one folder of yours
                                     (your laptop)

## What it does, exactly

When you start it, it asks which folder to share. From then on it waits, and
answers questions about that folder when the server asks. That is all it does.

It can, **within the folder you chose and nowhere else**:

| | |
|---|---|
| `list_directory` | say what files are there |
| `read_file` | return the contents of one, in pieces if it is large |
| `write_file` | save a file the AI produced |
| `get_file_info` | size and modification date |
| `ping` | confirm it is still alive |

## What it cannot do

**It never volunteers anything.** There is no upload, no sync, no scan. It
answers one question at a time and sends nothing otherwise.

**It cannot leave the folder you chose.** A path that resolves outside it is
refused. So is any shortcut, symlink or Windows junction *inside* it, even one
pointing somewhere harmless — because a link is the one way a path inside the
folder can name a file outside it, and refusing is a guarantee where following
correctly on every filesystem is only a hope.

**It never listens for connections.** It dials out to the server and nothing
can dial in. That is also why it does not trip Windows Firewall's "Allow
access?" prompt.

**It stops completely when you close the window.** Nothing is installed,
nothing is left running, and there is no background service.

## What it sends

Only what was asked for. File contents leave your machine only when the AI
asks to read a specific file, and only that file's contents.

**Note, while this is still in development: the connection is not yet
encrypted.** Until it is, use it on a folder of files you would not mind
crossing the internet in the clear. The server's own documentation tracks
that as the next thing to fix.

## The code

`main.go` is the whole program — one file, the Go standard library, and no
dependencies. `links_windows.go` and `links_other.go` are the check that keeps
a path inside your folder.

It is commented at some length, deliberately: roughly a fifth of the source
explains *why* rather than what, because most of it exists to stop a
plausible-looking change reintroducing a bug somebody already found.

## Building it

    BRIDGE_HOST=your.server goreleaser build --snapshot --clean

`BRIDGE_HOST` is stamped in at build time — a connector that does not know
where to dial would reach your own machine and find nothing, which looks like
a broken download rather than a wrong address.

Deliberately **not** built with `-s -w`. Stripping the symbol table makes the
binary smaller and changes nothing about how it runs, but antivirus
heuristics read a stripped, unsigned, freshly compiled binary as suspicious,
and four downloads were quarantined as a false positive before those flags
came off. Comments in the source say so at the point somebody would add them
back.

## Where the rest is

The server half — the web page, the bridge that speaks to the AI, the
protocol specification and the fixtures both implementations are checked
against — lives in a separate, private repository. This repository holds only
the program you are asked to run, which is the part you should be able to read
before running it.

This code moved here from that repository on 19 September 2026; its earlier
commit history is there.
