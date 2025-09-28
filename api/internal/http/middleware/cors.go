package middleware

import (
	"net/http"
	"recsys/shared/util"
	"regexp"
	"strings"

	"github.com/go-chi/cors"
)

// CORS returns a CORS middleware. It supports:
//   - CORS_ALLOWED_ORIGINS: CSV of origins, e.g.
//     "https://your-ui.vercel.app,https://example.com,*.vercel.app"
//     Exact matches work out-of-the-box. Patterns containing a "*" are
//     treated as wildcards (suffix or full-domain match).
//   - CORS_ALLOW_CREDENTIALS: "true" to allow credentials.
func CORS() func(next http.Handler) http.Handler {
	raw := strings.TrimSpace(util.MustGetEnv("CORS_ALLOWED_ORIGINS"))
	cred := strings.EqualFold(strings.TrimSpace(util.MustGetEnv("CORS_ALLOW_CREDENTIALS")), "true")

	// Split CSV and trim.
	var entries []string
	if raw != "" {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				entries = append(entries, p)
			}
		}
	}

	// Separate exacts and wildcard patterns.
	var exacts []string
	var wildcards []*regexp.Regexp
	for _, e := range entries {
		if strings.Contains(e, "*") {
			// Convert "*.vercel.app" to ^https?://([^.]+\.)*vercel\.app(?::\d+)?$
			pat := regexp.QuoteMeta(e)
			pat = strings.ReplaceAll(pat, `\*`, `[^/]+`)
			re := regexp.MustCompile(`^https?://` + pat + `(?::\d+)?$`)
			wildcards = append(wildcards, re)
		} else {
			exacts = append(exacts, e)
		}
	}

	allowAll := len(exacts) == 0 && len(wildcards) == 0

	allowFunc := func(r *http.Request, origin string) bool {
		if origin == "" {
			// No Origin header => not a CORS request.
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
		AllowCredentials: cred,
		MaxAge:           300,
	})
}
