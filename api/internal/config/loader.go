package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FieldError struct {
	Key string
	Err error
}

func (f FieldError) Error() string {
	return fmt.Sprintf("%s: %v", f.Key, f.Err)
}

// ValidationError aggregates configuration field errors.
type ValidationError struct {
	fields []FieldError
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	msgs := make([]string, 0, len(v.fields))
	for _, fe := range v.fields {
		msgs = append(msgs, fe.Error())
	}
	return "config validation failed: " + strings.Join(msgs, "; ")
}

// Fields returns the underlying field errors.
func (v ValidationError) Fields() []FieldError {
	return append([]FieldError(nil), v.fields...)
}

type loader struct {
	src   Source
	errs  []FieldError
	trims bool
}

func newLoader(src Source) *loader {
	if src == nil {
		src = EnvSource{}
	}
	return &loader{src: src, trims: true}
}

func (l *loader) appendErr(key string, err error) {
	l.errs = append(l.errs, FieldError{Key: key, Err: err})
}

func (l *loader) lookup(key string) (string, bool) {
	v, ok := l.src.Lookup(key)
	if !ok {
		return "", false
	}
	if l.trims {
		v = strings.TrimSpace(v)
	}
	return v, true
}

func (l *loader) requiredString(key string) string {
	if v, ok := l.lookup(key); ok && v != "" {
		return v
	}
	l.appendErr(key, fmt.Errorf("must be set"))
	return ""
}

func (l *loader) optionalString(key, def string) string {
	if v, ok := l.lookup(key); ok {
		if v == "" {
			return def
		}
		return v
	}
	return def
}

func (l *loader) stringSlice(key string, sep rune, allowEmpty bool) []string {
	raw, ok := l.lookup(key)
	if !ok || raw == "" {
		if allowEmpty {
			return nil
		}
		l.appendErr(key, fmt.Errorf("must include at least one entry"))
		return nil
	}
	parts := strings.Split(raw, string(sep))
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{})
	for _, part := range parts {
		val := strings.TrimSpace(part)
		if val == "" {
			continue
		}
		lower := strings.ToLower(val)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		out = append(out, lower)
	}
	if len(out) == 0 && !allowEmpty {
		l.appendErr(key, fmt.Errorf("must include at least one entry"))
	}
	return out
}

func (l *loader) bool(key string, def bool) bool {
	if v, ok := l.lookup(key); ok {
		switch strings.ToLower(v) {
		case "1", "t", "true", "y", "yes":
			return true
		case "0", "f", "false", "n", "no":
			return false
		case "":
			return def
		default:
			l.appendErr(key, fmt.Errorf("invalid boolean value: %q", v))
		}
	}
	return def
}

func (l *loader) positiveFloat(key string) float64 {
	v := l.requiredString(key)
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 {
		l.appendErr(key, fmt.Errorf("must be a positive number"))
		return 0
	}
	return f
}

func (l *loader) nonNegativeFloat(key string) float64 {
	v := l.requiredString(key)
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f < 0 {
		l.appendErr(key, fmt.Errorf("must be a non-negative number"))
		return 0
	}
	return f
}

func (l *loader) floatBetween(key string, min, max float64) float64 {
	v := l.requiredString(key)
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f < min || f > max {
		l.appendErr(key, fmt.Errorf("must be between %.2f and %.2f", min, max))
		return 0
	}
	return f
}

func (l *loader) intGreaterThan(key string, min int) int {
	v := l.requiredString(key)
	if v == "" {
		return 0
	}
	i, err := strconv.Atoi(v)
	if err != nil || i <= min {
		cmp := ">"
		if min >= 0 {
			cmp = fmt.Sprintf("> %d", min)
		}
		l.appendErr(key, fmt.Errorf("must be an integer %s", cmp))
		return 0
	}
	return i
}

func (l *loader) intNonNegative(key string) int {
	v := l.requiredString(key)
	if v == "" {
		return 0
	}
	i, err := strconv.Atoi(v)
	if err != nil || i < 0 {
		l.appendErr(key, fmt.Errorf("must be a non-negative integer"))
		return 0
	}
	return i
}

func (l *loader) optionalIntGreaterThan(key string, min int, def int) int {
	v, ok := l.lookup(key)
	if !ok || v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil || i <= min {
		l.appendErr(key, fmt.Errorf("must be an integer > %d", min))
		return def
	}
	return i
}

func (l *loader) optionalPositiveFloat(key string, def float64) float64 {
	v, ok := l.lookup(key)
	if !ok || v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 {
		l.appendErr(key, fmt.Errorf("must be a positive number"))
		return def
	}
	return f
}

func (l *loader) optionalDuration(key string, def time.Duration) time.Duration {
	v, ok := l.lookup(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		l.appendErr(key, fmt.Errorf("must be a positive duration"))
		return def
	}
	return d
}

func (l *loader) requiredDuration(key string) time.Duration {
	v := l.requiredString(key)
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		l.appendErr(key, fmt.Errorf("must be a positive duration"))
		return 0
	}
	return d
}

func (l *loader) optionalIntSlice(key string) []int16 {
	raw, ok := l.lookup(key)
	if !ok || raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int16, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.Atoi(part)
		if err != nil || val < -32768 || val > 32767 {
			l.appendErr(key, fmt.Errorf("invalid int16 value %q", part))
			continue
		}
		out = append(out, int16(val))
	}
	return out
}

func (l *loader) optionalStringMap(key string) map[string]float64 {
	raw, ok := l.lookup(key)
	if !ok || raw == "" || raw == "-" {
		return nil
	}
	entries := strings.Split(raw, ",")
	out := make(map[string]float64, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			l.appendErr(key, fmt.Errorf("invalid entry %q (expected namespace=rate)", entry))
			continue
		}
		ns := strings.TrimSpace(kv[0])
		if ns == "" {
			l.appendErr(key, fmt.Errorf("namespace missing in %q", entry))
			continue
		}
		rateStr := strings.TrimSpace(kv[1])
		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil || rate < 0 || rate > 1 {
			l.appendErr(key, fmt.Errorf("invalid rate %q for namespace %s", rateStr, ns))
			continue
		}
		out[ns] = rate
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (l *loader) err() error {
	if len(l.errs) == 0 {
		return nil
	}
	return ValidationError{fields: l.errs}
}
