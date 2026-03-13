package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runDiff(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	// Get remote (deployed) definition from active version
	var remote *ApplicationDefinition
	if appDetail.ActiveVersion != nil {
		verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
		verDetail, err := verOp.Read(ctx, v1.ApplicationVersionNumber(*appDetail.ActiveVersion))
		if err != nil {
			return fmt.Errorf("failed to read active version: %w", err)
		}
		remote = versionDetailToDefinition(verDetail)
		remote.Cluster = c.app.Cluster
		remote.Name = c.app.Name
	} else {
		// No active version; treat remote as empty
		remote = &ApplicationDefinition{
			Cluster: c.app.Cluster,
			Name:    c.app.Name,
		}
	}

	remoteJSON, err := marshalForDiff(remote)
	if err != nil {
		return fmt.Errorf("failed to marshal remote definition: %w", err)
	}
	localJSON, err := marshalForDiff(c.app)
	if err != nil {
		return fmt.Errorf("failed to marshal local definition: %w", err)
	}

	found := printColoredDiff(os.Stdout, remoteJSON, localJSON, "remote", "local")
	if !found {
		fmt.Fprintln(os.Stderr, "No differences found.")
	}
	return nil
}

func marshalForDiff(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

var (
	colorHeader = color.New(color.Bold)
	colorHunk   = color.New(color.FgCyan)
	colorAdd    = color.New(color.FgGreen)
	colorDel    = color.New(color.FgRed)
)

// printColoredDiff prints a colored unified diff to w. Returns true if differences were found.
func printColoredDiff(w io.Writer, a, b, labelA, labelB string) bool {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")

	edits := myersDiff(aLines, bLines)

	if !hasChanges(edits) {
		return false
	}

	colorHeader.Fprintf(w, "--- %s\n", labelA)
	colorHeader.Fprintf(w, "+++ %s\n", labelB)

	hunks := buildHunks(edits, 3)
	for _, h := range hunks {
		colorHunk.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", h.aStart+1, h.aCount, h.bStart+1, h.bCount)
		for _, line := range h.lines {
			switch {
			case strings.HasPrefix(line, "+"):
				colorAdd.Fprintln(w, line)
			case strings.HasPrefix(line, "-"):
				colorDel.Fprintln(w, line)
			default:
				fmt.Fprintln(w, line)
			}
		}
	}
	return true
}

type editKind int

const (
	editEqual  editKind = iota
	editDelete          // in a only
	editInsert          // in b only
)

type edit struct {
	kind editKind
	line string
}

func myersDiff(a, b []string) []edit {
	n, m := len(a), len(b)
	max := n + m
	if max == 0 {
		return nil
	}

	// v[k] = x of furthest reaching path in diagonal k
	// offset by max so index is always non-negative
	v := make([]int, 2*max+1)
	type snap struct {
		v    []int
		d, k int
	}
	var trace [][]int

	for d := 0; d <= max; d++ {
		vc := make([]int, len(v))
		copy(vc, v)
		trace = append(trace, vc)

		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[k-1+max] < v[k+1+max]) {
				x = v[k+1+max]
			} else {
				x = v[k-1+max] + 1
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[k+max] = x
			if x >= n && y >= m {
				return backtrack(trace, a, b, max)
			}
		}
	}
	return backtrack(trace, a, b, max)
}

func backtrack(trace [][]int, a, b []string, max int) []edit {
	x, y := len(a), len(b)
	var edits []edit

	for d := len(trace) - 1; d >= 0; d-- {
		v := trace[d]
		k := x - y
		var prevK int
		if k == -d || (k != d && v[k-1+max] < v[k+1+max]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := v[prevK+max]
		prevY := prevX - prevK

		// diagonal (equal)
		for x > prevX && y > prevY {
			x--
			y--
			edits = append(edits, edit{editEqual, a[x]})
		}

		if d > 0 {
			if x == prevX {
				// insert
				y--
				edits = append(edits, edit{editInsert, b[y]})
			} else {
				// delete
				x--
				edits = append(edits, edit{editDelete, a[x]})
			}
		}
	}

	// Reverse
	for i, j := 0, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}
	return edits
}

func hasChanges(edits []edit) bool {
	for _, e := range edits {
		if e.kind != editEqual {
			return true
		}
	}
	return false
}

type hunk struct {
	aStart, aCount int
	bStart, bCount int
	lines          []string
}

func buildHunks(edits []edit, contextLines int) []hunk {
	// Find change ranges, then expand with context
	type changeRange struct {
		start, end int // indices into edits
	}
	var ranges []changeRange
	i := 0
	for i < len(edits) {
		if edits[i].kind != editEqual {
			start := i
			for i < len(edits) && edits[i].kind != editEqual {
				i++
			}
			ranges = append(ranges, changeRange{start, i})
		} else {
			i++
		}
	}

	if len(ranges) == 0 {
		return nil
	}

	// Merge ranges that overlap when context is applied
	type contextRange struct {
		start, end int
	}
	var merged []contextRange
	for _, r := range ranges {
		cs := r.start - contextLines
		if cs < 0 {
			cs = 0
		}
		ce := r.end + contextLines
		if ce > len(edits) {
			ce = len(edits)
		}
		if len(merged) > 0 && cs <= merged[len(merged)-1].end {
			merged[len(merged)-1].end = ce
		} else {
			merged = append(merged, contextRange{cs, ce})
		}
	}

	var hunks []hunk
	for _, cr := range merged {
		h := hunk{}
		aLine, bLine := 0, 0
		// Count lines before this range
		for j := 0; j < cr.start; j++ {
			switch edits[j].kind {
			case editEqual:
				aLine++
				bLine++
			case editDelete:
				aLine++
			case editInsert:
				bLine++
			}
		}
		h.aStart = aLine
		h.bStart = bLine

		for j := cr.start; j < cr.end; j++ {
			e := edits[j]
			switch e.kind {
			case editEqual:
				h.lines = append(h.lines, " "+e.line)
				h.aCount++
				h.bCount++
			case editDelete:
				h.lines = append(h.lines, "-"+e.line)
				h.aCount++
			case editInsert:
				h.lines = append(h.lines, "+"+e.line)
				h.bCount++
			}
		}
		hunks = append(hunks, h)
	}
	return hunks
}
