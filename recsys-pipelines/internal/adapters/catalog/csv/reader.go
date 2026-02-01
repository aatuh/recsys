package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
)

// Reader loads item tags from a CSV file.
type Reader struct {
	Path string
}

func New(path string) *Reader {
	return &Reader{Path: path}
}

func (r *Reader) Read(ctx context.Context) ([]signals.ItemTag, error) {
	if r == nil || r.Path == "" {
		return nil, fmt.Errorf("csv path is required")
	}
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	idx := indexHeaders(headers)
	if idx.itemID < 0 {
		return nil, fmt.Errorf("csv missing item_id column")
	}

	items := make([]signals.ItemTag, 0, 1024)
	for {
		if ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		itemID := getField(rec, idx.itemID)
		if itemID == "" {
			continue
		}
		it := signals.ItemTag{ItemID: itemID}
		if idx.tags >= 0 {
			it.Tags = splitTags(getField(rec, idx.tags))
		}
		if idx.price >= 0 {
			if v := strings.TrimSpace(getField(rec, idx.price)); v != "" {
				if f, err := parseFloat(v); err == nil {
					it.Price = &f
				}
			}
		}
		if idx.createdAt >= 0 {
			if v := strings.TrimSpace(getField(rec, idx.createdAt)); v != "" {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					it.CreatedAt = t
				}
			}
		}
		if idx.namespace >= 0 {
			it.Namespace = strings.TrimSpace(getField(rec, idx.namespace))
		}
		items = append(items, it)
	}
	return items, nil
}

var _ catalog.Reader = (*Reader)(nil)

type headerIndex struct {
	itemID    int
	tags      int
	price     int
	createdAt int
	namespace int
}

func indexHeaders(headers []string) headerIndex {
	idx := headerIndex{itemID: -1, tags: -1, price: -1, createdAt: -1, namespace: -1}
	for i, h := range headers {
		key := strings.ToLower(strings.TrimSpace(h))
		switch key {
		case "item_id", "itemid", "id":
			idx.itemID = i
		case "tags", "tag_list":
			idx.tags = i
		case "price":
			idx.price = i
		case "created_at", "createdat":
			idx.createdAt = i
		case "namespace", "surface":
			idx.namespace = i
		}
	}
	return idx
}

func getField(rec []string, idx int) string {
	if idx < 0 || idx >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[idx])
}

func splitTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '|'
	})
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func parseFloat(raw string) (float64, error) {
	return strconv.ParseFloat(raw, 64)
}
