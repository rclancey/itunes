package itunes

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func pidsToText(pids []PersistentID) string {
	lines := make([]string, len(pids))
	for i, pid := range pids {
		lines[i] = pid.String()
	}
	return strings.Join(lines, "\n")
}

func ThreeWayMerge(base, delta_one, delta_two []PersistentID) ([]PersistentID, bool) {
	base_s := pidsToText(base)
	delta_one_s := pidsToText(delta_one)
	delta_two_s := pidsToText(delta_two)
	dmp := diffmatchpatch.New()
	patches := dmp.PatchMake(base_s, delta_one_s)
	res, applied := dmp.PatchApply(patches, delta_two_s)
	for _, app := range applied {
		if !app {
			return delta_two, false
		}
	}
	lines := strings.Split(res, "\n")
	pids := make([]PersistentID, len(lines))
	for i, line := range lines {
		var pid PersistentID
		if (&pid).DecodeString(line) == nil {
			pids[i] = pid
		}
	}
	return pids, true
}

