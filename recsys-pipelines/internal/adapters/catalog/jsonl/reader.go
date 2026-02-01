package jsonl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
)

type rawItem struct {
	ItemID    string    `json:"item_id"`
	Tags      []string  `json:"tags"`
	Price     *float64  `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	Namespace string    `json:"namespace"`
}

// Reader loads item tags from a JSONL file.
type Reader struct {
	Path string
}

func New(path string) *Reader {
	return &Reader{Path: path}
}

func (r *Reader) Read(ctx context.Context) ([]signals.ItemTag, error) {
	if r == nil || r.Path == "" {
		return nil, fmt.Errorf("jsonl path is required")
	}
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	items := make([]signals.ItemTag, 0, 1024)
	for scanner.Scan() {
		if ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw rawItem
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, err
		}
		if raw.ItemID == "" {
			continue
		}
		items = append(items, signals.ItemTag{
			ItemID:    raw.ItemID,
			Tags:      raw.Tags,
			Price:     raw.Price,
			CreatedAt: raw.CreatedAt,
			Namespace: raw.Namespace,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

var _ catalog.Reader = (*Reader)(nil)
