package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/cors"
)

// CORSOptions represents configuration for the CORS middleware.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

// CORS returns a CORS middleware configured using the provided options.
func CORS(opts CORSOptions) func(next http.Handler) http.Handler {
	var exacts []string
	var wildcards []*regexp.Regexp

	for _, origin := range opts.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if strings.Contains(origin, "*") {
			pat := regexp.QuoteMeta(origin)
			pat = strings.ReplaceAll(pat, `\*`, `[^/]+`)
			re := regexp.MustCompile(`^https?://` + pat + `(?::\d+)?$`)
			wildcards = append(wildcards, re)
			continue
		}
		exacts = append(exacts, origin)
	}

	allowAll := len(exacts) == 0 && len(wildcards) == 0

	allowFunc := func(r *http.Request, origin string) bool {
		if origin == "" {
			return false
		}
		if allowAll {
			return true
		}
		for _, e := range exacts {
			if strings.EqualFold(origin, e) {
				return true
			}
		}
		for _, re := range wildcards {
			if re.MatchString(origin) {
				return true
			}
		}
		return false
	}

	return cors.Handler(cors.Options{
		AllowOriginFunc:  allowFunc,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Org-ID", "X-Requested-With", "X-Request-Id"},
		ExposedHeaders:   []string{"Link", "X-Request-Id"},
		AllowCredentials: opts.AllowCredentials,
		MaxAge:           300,
	})
}
