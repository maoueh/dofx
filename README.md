## Desjardins OFX

A (hackish, quick) tool to print stats and clean up
desjardins QFX exported file that are in the OFX format
but it seems without tag ending?

I'm far from having the propre knowledge to create a full
blown reader/writer for OFX/QFX file format.

However, I add to transform my QFX exported file from
Desjardins website. 

I wanted to:

- Print stats about the file (min transaction date, max transaction date, duplicated `FITID`).
- Clean up duplicated `FITID` to ensure they are all unique to please Quicken import which assumes it's the same.
- Update the `FITID` for a given date range so I could re-import deleted transactions from Quicken.

The [main.go](./main.go) is a quick text processing tool that
does all this. 

The compiled binary permits `stats` and `clean` commands. For
the update `FITID` for range, you will have to edit the source
code and link the `update_fitid_in_range` function somewhere
and edit the date range parameters accordingly.