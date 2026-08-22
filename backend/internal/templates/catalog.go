package templates

import (
	"fmt"
	"sort"
	"strings"
)

// TitleType is the media category of a catalog entry.
type TitleType string

// Media categories.
const (
	TypeAnime  TitleType = "anime"
	TypeMovie  TitleType = "kino"
	TypeSeries TitleType = "serial"
)

// Title is one searchable anime/movie entry of a bot catalog.
type Title struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Type        TitleType `json:"type"`
	Genres      []string  `json:"genres"`
	Year        int       `json:"year"`
	Rating      float64   `json:"rating"`
	Episodes    int       `json:"episodes,omitempty"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags,omitempty"`
}

// TypeLabel returns the human readable category name.
func (t Title) TypeLabel() string {
	switch t.Type {
	case TypeMovie:
		return "🎬 Kino"
	case TypeSeries:
		return "📺 Serial"
	default:
		return "🎌 Anime"
	}
}

// EpisodeLine renders the episode count line (empty for movies).
func (t Title) EpisodeLine() string {
	if t.Episodes > 0 {
		return fmt.Sprintf("📦 Bo'limlar: %d\n", t.Episodes)
	}
	return ""
}

// normalizeCode uppercases and trims a catalog code.
func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// FindByCode returns the title matching the given code (case-insensitive).
func FindByCode(catalog []Title, code string) (Title, bool) {
	code = normalizeCode(code)
	for _, t := range catalog {
		if normalizeCode(t.Code) == code {
			return t, true
		}
	}
	return Title{}, false
}

// SearchCatalog returns entries matching the query, best rated first.
func SearchCatalog(catalog []Title, query string) []Title {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	var hits []Title
	for _, t := range catalog {
		if matches(t, query) {
			hits = append(hits, t)
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		return hits[i].Rating > hits[j].Rating
	})
	return hits
}

func matches(t Title, query string) bool {
	haystack := strings.ToLower(t.Title + " " + t.Code + " " + t.Description + " " + strings.Join(t.Genres, " ") + " " + strings.Join(t.Tags, " "))
	return strings.Contains(haystack, query)
}

// TopRated returns the highest rated entries of the catalog.
func TopRated(catalog []Title, limit int) []Title {
	sorted := make([]Title, len(catalog))
	copy(sorted, catalog)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Rating > sorted[j].Rating
	})
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}
	return sorted
}

// AllTemplates returns every built-in bot template in registration order.
func AllTemplates() []Template {
	return []Template{
		&AniTez{},
		&AniXUltra{},
		&AIAssistant{},
		&EcommerceShop{},
		&FeedbackSupport{},
		&ChannelManager{},
		&CustomBuilder{},
	}
}
