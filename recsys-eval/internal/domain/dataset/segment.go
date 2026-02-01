package dataset

import (
	"sort"
	"strings"
)

// SegmentKey builds a stable segment key for the provided slice keys.
func SegmentKey(ctx map[string]string, keys []string) string {
	if len(keys) == 0 {
		return "__all__"
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		val := ctx[key]
		parts = append(parts, key+"="+val)
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}
