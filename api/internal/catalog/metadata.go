package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"recsys/internal/store"
)

// Options controls metadata derivation behaviour.
type Options struct {
	// GenerateEmbedding computes a deterministic fallback embedding from metadata.
	GenerateEmbedding bool
}

// Result captures an ItemUpsert plus whether any fields changed.
type Result struct {
	Upsert  store.ItemUpsert
	Changed bool
}

// BuildUpsert derives metadata for the given catalog row and prepares an upsert payload.
func BuildUpsert(row store.CatalogItem, opts Options) (Result, error) {
	props := parseProps(row.Props)
	name := getString(props, "name")
	currency := getString(props, "currency")

	brand := firstNonEmpty(
		ptrToString(row.Brand),
		getString(props, "brand"),
		extractTag(row.Tags, "brand:"),
	)

	category := firstNonEmpty(
		ptrToString(row.Category),
		getString(props, "category"),
		extractTag(row.Tags, "category:"),
	)

	description := firstNonEmpty(
		ptrToString(row.Description),
		getString(props, "description"),
	)

	imageURL := firstNonEmpty(
		ptrToString(row.ImageURL),
		getString(props, "image_url"),
	)

	categoryPath := deriveCategoryPath(row.CategoryPath, props, category)
	metadataVersion := computeMetadataVersion(metadataVersionInput{
		ItemID:      row.ItemID,
		Name:        name,
		Brand:       brand,
		Category:    category,
		Description: description,
		ImageURL:    imageURL,
		Price:       row.Price,
		Currency:    currency,
		UpdatedAt:   row.UpdatedAt,
	})

	propsChanged := false
	if props == nil {
		props = map[string]any{}
	}
	if setStringProp(props, "brand", brand) {
		propsChanged = true
	}
	if setStringProp(props, "category", category) {
		propsChanged = true
	}
	if setStringProp(props, "description", description) {
		propsChanged = true
	}
	if setStringProp(props, "image_url", imageURL) {
		propsChanged = true
	}
	if setStringSliceProp(props, "category_path", categoryPath) {
		propsChanged = true
	}
	if setStringProp(props, "metadata_version", metadataVersion) {
		propsChanged = true
	}

	upsert := store.ItemUpsert{
		ItemID:    row.ItemID,
		Available: row.Available,
		Price:     row.Price,
		Tags:      append([]string(nil), row.Tags...),
		Props:     props,
	}

	changed := false

	if ptr, ok := maybeUpdateString(row.Brand, brand); ok {
		upsert.Brand = ptr
		changed = true
	}
	if ptr, ok := maybeUpdateString(row.Category, category); ok {
		upsert.Category = ptr
		changed = true
	}
	if ptr, ok := maybeUpdateSlice(row.CategoryPath, categoryPath); ok {
		upsert.CategoryPath = ptr
		changed = true
	}
	if ptr, ok := maybeUpdateString(row.Description, description); ok {
		upsert.Description = ptr
		changed = true
	}
	if ptr, ok := maybeUpdateString(row.ImageURL, imageURL); ok {
		upsert.ImageURL = ptr
		changed = true
	}
	if ptr, ok := maybeUpdateString(row.MetadataVersion, metadataVersion); ok {
		upsert.MetadataVersion = ptr
		changed = true
	}

	if opts.GenerateEmbedding {
		if emb, changedEmb := buildEmbedding(row.Embedding, brand, category, description); changedEmb {
			upsert.Embedding = &emb
			changed = true
		}
	}

	if propsChanged {
		changed = true
	}

	return Result{Upsert: upsert, Changed: changed}, nil
}

func parseProps(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func getString(props map[string]any, key string) string {
	if val, ok := props[key]; ok {
		switch v := val.(type) {
		case string:
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func getStringSlice(props map[string]any, key string) []string {
	val, ok := props[key]
	if !ok {
		return nil
	}
	rawSlice, ok := val.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(rawSlice))
	for _, entry := range rawSlice {
		switch s := entry.(type) {
		case string:
			str := strings.TrimSpace(s)
			if str != "" {
				result = append(result, str)
			}
		}
	}
	return result
}

func ptrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return strings.TrimSpace(*ptr)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func extractTag(tags []string, prefix string) string {
	prefixLower := strings.ToLower(prefix)
	for _, tag := range tags {
		if strings.HasPrefix(strings.ToLower(tag), prefixLower) {
			val := strings.TrimPrefix(tag, prefix)
			if strings.EqualFold(tag[:len(prefix)], prefix) {
				val = tag[len(prefix):]
			}
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func deriveCategoryPath(existing []string, props map[string]any, category string) []string {
	if len(existing) > 0 {
		cp := make([]string, len(existing))
		copy(cp, existing)
		return cp
	}
	if fromProps := getStringSlice(props, "category_path"); len(fromProps) > 0 {
		return fromProps
	}
	if category == "" {
		return nil
	}
	splitters := []string{">", "/", "|"}
	for _, splitter := range splitters {
		if strings.Contains(category, splitter) {
			parts := strings.Split(category, splitter)
			result := make([]string, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					result = append(result, part)
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}
	return []string{strings.TrimSpace(category)}
}

func setStringProp(props map[string]any, key, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if current, ok := props[key]; ok {
		switch v := current.(type) {
		case string:
			if strings.TrimSpace(v) == value {
				return false
			}
		default:
			if str := fmt.Sprint(v); strings.TrimSpace(str) == value {
				return false
			}
		}
	}
	props[key] = value
	return true
}

func setStringSliceProp(props map[string]any, key string, values []string) bool {
	if len(values) == 0 {
		return false
	}
	if current, ok := props[key]; ok {
		if sliceEqual(normalizeToStringSlice(current), values) {
			return false
		}
	}
	props[key] = append([]string(nil), values...)
	return true
}

func normalizeToStringSlice(val any) []string {
	switch v := val.(type) {
	case []string:
		out := make([]string, len(v))
		for i, s := range v {
			out[i] = strings.TrimSpace(s)
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, entry := range v {
			switch s := entry.(type) {
			case string:
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

func maybeUpdateString(existing *string, derived string) (*string, bool) {
	derived = strings.TrimSpace(derived)
	if derived == "" {
		return nil, false
	}
	if existing != nil && strings.TrimSpace(*existing) == derived {
		return nil, false
	}
	val := derived
	return &val, true
}

func maybeUpdateSlice(existing []string, derived []string) (*[]string, bool) {
	if len(derived) == 0 {
		return nil, false
	}
	if len(existing) == len(derived) {
		match := true
		for i := range derived {
			if strings.TrimSpace(existing[i]) != derived[i] {
				match = false
				break
			}
		}
		if match {
			return nil, false
		}
	}
	cp := append([]string(nil), derived...)
	return &cp, true
}

type metadataVersionInput struct {
	ItemID      string
	Name        string
	Brand       string
	Category    string
	Description string
	ImageURL    string
	Price       *float64
	Currency    string
	UpdatedAt   time.Time
}

func computeMetadataVersion(input metadataVersionInput) string {
	var price float64
	if input.Price != nil {
		price = *input.Price
	}
	var updatedAt *string
	if !input.UpdatedAt.IsZero() {
		val := input.UpdatedAt.UTC().Format(time.RFC3339Nano)
		updatedAt = &val
	}

	payload := struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Brand       string  `json:"brand"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		ImageURL    string  `json:"imageUrl"`
		Price       float64 `json:"price"`
		Currency    string  `json:"currency"`
		UpdatedAt   *string `json:"updatedAt,omitempty"`
	}{
		ID:          input.ItemID,
		Name:        input.Name,
		Brand:       input.Brand,
		Category:    input.Category,
		Description: input.Description,
		ImageURL:    input.ImageURL,
		Price:       price,
		Currency:    input.Currency,
		UpdatedAt:   updatedAt,
	}

	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:16]
}

func buildEmbedding(existing []float64, brand, category, description string) ([]float64, bool) {
	text := strings.TrimSpace(strings.Join([]string{brand, category, description}, " "))
	if text == "" {
		if len(existing) == 0 {
			return nil, false
		}
		return nil, true
	}
	vec := deterministicEmbedding(text)
	if embeddingEqual(existing, vec) {
		return nil, false
	}
	return vec, true
}

func deterministicEmbedding(text string) []float64 {
	seed := []byte(strings.ToLower(text))
	if len(seed) == 0 {
		return nil
	}
	dims := EmbeddingDims()
	vec := make([]float64, dims)
	block := sha256.Sum256(seed)
	for i := 0; i < dims; i++ {
		if i != 0 && i%len(block) == 0 {
			counter := []byte{byte(i / len(block))}
			extended := append(seed, counter...)
			block = sha256.Sum256(extended)
		}
		b := block[i%len(block)]
		vec[i] = (float64(int(b)) / 127.5) - 1.0
	}
	return vec
}

func embeddingEqual(existing, candidate []float64) bool {
	if len(existing) == 0 && len(candidate) == 0 {
		return true
	}
	if len(existing) != len(candidate) {
		return false
	}
	const epsilon = 1e-6
	for i := range existing {
		if diff := existing[i] - candidate[i]; diff > epsilon || diff < -epsilon {
			return false
		}
	}
	return true
}

// EmbeddingDims exposes the store embedding dimension constant for deterministic embedding generation.
func EmbeddingDims() int {
	return store.EmbeddingDims
}
