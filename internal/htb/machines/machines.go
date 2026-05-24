// Package machines is the HackTheBox API client for the machines domain.
//
// It implements one method per documented operation in
// openapi/openapi.yaml under the `machines` tag. The HTB API mixes two
// URL roots (`/machine/...` and `/machines/...`); this package follows
// whatever the spec documents for each operation. Where the API uses an
// envelope (`info`, `message`, ...) we unwrap it and expose the inner
// payload as the typed return value.
package machines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bgrewell/htb-app-mcp/internal/htb"
)

// Client is the machines-domain client. Construct with New.
type Client struct {
	c *htb.Client
}

// New wraps a shared HTB client.
func New(c *htb.Client) *Client {
	return &Client{c: c}
}

// ---------- types ----------

// SpawnInfo mirrors components.schemas.SpawnInfo in openapi.yaml.
type SpawnInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	LabServer   string `json:"lab_server,omitempty"`
	VPNServerID *int64 `json:"vpn_server_id,omitempty"`
	IP          string `json:"ip,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	IsSpawning  bool   `json:"isSpawning,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	TierID      *int   `json:"tier_id,omitempty"`
}

// Maker is a machine maker (community contributor).
type Maker struct {
	ID          int64  `json:"id"`
	Name        string `json:"name,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	IsRespected bool   `json:"isRespected,omitempty"`
	ProfileURL  string `json:"profile_url,omitempty"`
}

// UserRef is a minimal embedded user reference.
type UserRef struct {
	ID     int64  `json:"id"`
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

// BloodHolder represents first-blood metadata.
type BloodHolder struct {
	BloodDifference string  `json:"blood_difference,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	User            UserRef `json:"user"`
}

// FeedbackChart is the difficulty-rating histogram.
type FeedbackChart struct {
	CounterCake      int `json:"counterCake"`
	CounterVeryEasy  int `json:"counterVeryEasy"`
	CounterEasy      int `json:"counterEasy"`
	CounterTooEasy   int `json:"counterTooEasy"`
	CounterMedium    int `json:"counterMedium"`
	CounterBitHard   int `json:"counterBitHard"`
	CounterHard      int `json:"counterHard"`
	CounterTooHard   int `json:"counterTooHard"`
	CounterExHard    int `json:"counterExHard"`
	CounterBrainFuck int `json:"counterBrainFuck"`
}

// PlayInfo is per-caller play state.
type PlayInfo struct {
	IsSpawned         bool    `json:"isSpawned"`
	IsSpawning        bool    `json:"isSpawning"`
	IsActive          *bool   `json:"isActive,omitempty"`
	ActivePlayerCount *int    `json:"active_player_count,omitempty"`
	ExpiresAt         *string `json:"expires_at,omitempty"`
	LifeRemaining     *int    `json:"life_remaining,omitempty"`
}

// MachineInfo is the full single-machine info object returned by
// GET /machine/profile/{name}. Many fields are caller-relative.
type MachineInfo struct {
	ID                         int64           `json:"id"`
	Name                       string          `json:"name"`
	OS                         string          `json:"os,omitempty"`
	DifficultyText             string          `json:"difficultyText,omitempty"`
	Points                     int             `json:"points"`
	StaticPoints               int             `json:"static_points"`
	ExperiencePoints           int             `json:"experience_points"`
	Release                    string          `json:"release,omitempty"`
	Retired                    bool            `json:"retired"`
	Active                     bool            `json:"active"`
	Free                       bool            `json:"free"`
	Avatar                     string          `json:"avatar,omitempty"`
	Synopsis                   string          `json:"synopsis,omitempty"`
	IP                         string          `json:"ip,omitempty"`
	Stars                      float64         `json:"stars"`
	ReviewsCount               int             `json:"reviews_count"`
	UserOwnsCount              int             `json:"user_owns_count"`
	RootOwnsCount              int             `json:"root_owns_count"`
	UserBlood                  *BloodHolder    `json:"userBlood,omitempty"`
	RootBlood                  *BloodHolder    `json:"rootBlood,omitempty"`
	BotHasBlood                bool            `json:"botHasBlood"`
	Maker                      *Maker          `json:"maker,omitempty"`
	Maker2                     *Maker          `json:"maker2,omitempty"`
	PlayInfo                   *PlayInfo       `json:"playInfo,omitempty"`
	FeedbackForChart           *FeedbackChart  `json:"feedbackForChart,omitempty"`
	AcademyModules             json.RawMessage `json:"academy_modules,omitempty"`
	InfoStatus                 *string         `json:"info_status,omitempty"`
	MachinePwnedDate           *string         `json:"machinePwnedDate,omitempty"`
	MachineMode                *string         `json:"machine_mode,omitempty"`
	SeasonID                   *int            `json:"season_id,omitempty"`
	RequiredSubscription       *string         `json:"requiredSubscription,omitempty"`
	PriceTier                  int             `json:"priceTier"`
	StartMode                  string          `json:"start_mode,omitempty"`
	IsGuidedEnabled            bool            `json:"isGuidedEnabled"`
	IsSingleFlag               bool            `json:"isSingleFlag"`
	IsTodo                     bool            `json:"isTodo"`
	HasChangelog               bool            `json:"has_changelog"`
	CanAccessWalkthrough       bool            `json:"can_access_walkthrough"`
	ShowGoVIP                  bool            `json:"show_go_vip"`
	ShowGoVIPServer            bool            `json:"show_go_vip_server"`
	SpFlag                     int             `json:"sp_flag"`
	SwitchServerWarning        *string         `json:"switchServerWarning,omitempty"`
	Recommended                bool            `json:"recommended"`
	OwnRank                    int             `json:"ownRank"`
	AuthUserInUserOwns         bool            `json:"authUserInUserOwns"`
	AuthUserInRootOwns         bool            `json:"authUserInRootOwns"`
	AuthUserFirstUserTime      string          `json:"authUserFirstUserTime,omitempty"`
	AuthUserFirstRootTime      string          `json:"authUserFirstRootTime,omitempty"`
	AuthUserHasReviewed        bool            `json:"authUserHasReviewed"`
	AuthUserHasSubmittedMatrix bool            `json:"authUserHasSubmittedMatrix"`
	UserCanReview              bool            `json:"user_can_review"`
}

// RecommendationCard is one slot in the recommendations payload.
type RecommendationCard struct {
	ID                 int64          `json:"id"`
	Name               string         `json:"name"`
	OS                 string         `json:"os,omitempty"`
	DifficultyText     string         `json:"difficultyText,omitempty"`
	Avatar             string         `json:"avatar,omitempty"`
	Points             int            `json:"points"`
	Release            string         `json:"release,omitempty"`
	Retired            int            `json:"retired"`
	RetiredDate        *string        `json:"retired_date,omitempty"`
	TypeCard           string         `json:"typeCard"`
	IsTodo             bool           `json:"isTodo"`
	Stars              int            `json:"stars"`
	RootOwnsCount      int            `json:"root_owns_count"`
	UserOwnsCount      int            `json:"user_owns_count"`
	AuthUserInUserOwns bool           `json:"authUserInUserOwns"`
	AuthUserInRootOwns bool           `json:"authUserInRootOwns"`
	Maker              *Maker         `json:"maker,omitempty"`
	Maker2             *Maker         `json:"maker2,omitempty"`
	FeedbackForChart   *FeedbackChart `json:"feedbackForChart,omitempty"`
}

// Recommendations is the response of GET /machine/recommended.
type Recommendations struct {
	Card1 RecommendationCard `json:"card1"`
	Card2 RecommendationCard `json:"card2"`
	State []string           `json:"state"`
}

// OfficialWriteup is the metadata for the official PDF writeup.
type OfficialWriteup struct {
	Filename                   string `json:"filename"`
	SHA256                     string `json:"sha256"`
	Rating                     int    `json:"rating"`
	TotalRatings               int    `json:"total_ratings"`
	ThresholdForDisplayReached int    `json:"threshold_for_display_reached"`
	LikedByUser                *bool  `json:"likedByUser,omitempty"`
	DislikedByUser             *bool  `json:"dislikedByUser,omitempty"`
}

// OfficialVideo is the metadata for the official video walkthrough.
type OfficialVideo struct {
	CreatorID     *int    `json:"creator_id,omitempty"`
	CreatorName   *string `json:"creator_name,omitempty"`
	CreatorAvatar *string `json:"creator_avatar,omitempty"`
	YoutubeID     *string `json:"youtube_id,omitempty"`
}

// CommunityWriteup is a community-authored writeup link.
type CommunityWriteup struct {
	ID             int64   `json:"id"`
	URL            string  `json:"url"`
	Rating         int     `json:"rating"`
	TotalRatings   int     `json:"total_ratings"`
	CreatedAt      string  `json:"created_at"`
	LanguageCode   *string `json:"language_code,omitempty"`
	LanguageName   *string `json:"language_name,omitempty"`
	LikedByUser    *bool   `json:"liked_by_user,omitempty"`
	DislikedByUser *bool   `json:"disliked_by_user,omitempty"`
	UserID         int64   `json:"user_id"`
	UserName       string  `json:"user_name"`
	UserAvatar     *string `json:"user_avatar,omitempty"`
}

// WalkthroughsBundle is the composite payload of /machine/walkthroughs/{id}.
type WalkthroughsBundle struct {
	Official    *OfficialWriteup   `json:"official,omitempty"`
	Video       *OfficialVideo     `json:"video,omitempty"`
	Writeups    []CommunityWriteup `json:"writeups,omitempty"`
	UnderReview json.RawMessage    `json:"under_review,omitempty"`
}

// Review is one user review of a machine.
type Review struct {
	ID                       int64           `json:"id"`
	Stars                    int             `json:"stars,omitempty"`
	Title                    string          `json:"title,omitempty"`
	Headline                 string          `json:"headline,omitempty"`
	Message                  string          `json:"message,omitempty"`
	Review                   string          `json:"review,omitempty"`
	Difficulty               *string         `json:"difficulty,omitempty"`
	Featured                 int             `json:"featured,omitempty"`
	Released                 int             `json:"released,omitempty"`
	CreatedAt                string          `json:"created_at,omitempty"`
	HelpfulReviewsCount      int             `json:"helpful_reviews_count,omitempty"`
	HelpfulReviews           json.RawMessage `json:"helpful_reviews,omitempty"`
	AuthUserInHelpfulReviews bool            `json:"authUserInHelpfulReviews,omitempty"`
	UserID                   int64           `json:"user_id,omitempty"`
	User                     UserRef         `json:"user"`
}

// PaginatorLinks is the top-level links block on a Laravel paginator response.
type PaginatorLinks struct {
	First *string `json:"first,omitempty"`
	Last  *string `json:"last,omitempty"`
	Next  *string `json:"next,omitempty"`
	Prev  *string `json:"prev,omitempty"`
}

// PaginationMeta is the Laravel paginator meta block.
type PaginationMeta struct {
	CurrentPage int    `json:"current_page"`
	PerPage     int    `json:"per_page"`
	From        *int   `json:"from,omitempty"`
	To          *int   `json:"to,omitempty"`
	Total       int    `json:"total"`
	LastPage    int    `json:"last_page"`
	Path        string `json:"path,omitempty"`
}

// ReviewPage is the paginated review response.
type ReviewPage struct {
	Average float64        `json:"average"`
	Count   int            `json:"count"`
	Data    []Review       `json:"data"`
	Links   PaginatorLinks `json:"links"`
	Meta    PaginationMeta `json:"meta"`
}

// ListReviewsOptions controls pagination/sort for ListReviews.
type ListReviewsOptions struct {
	Page     int      // 1-based; zero means "use API default".
	PerPage  int      // zero means "use API default" (15).
	SortType string   // "asc" or "desc"; empty means "use API default".
	SortBy   []string // repeated `sort_by[]` values.
}

// ---------- methods ----------

// GetActiveSpawn returns the caller's currently-spawned machine, or nil
// when nothing is currently spawned.
func (m *Client) GetActiveSpawn(ctx context.Context) (*SpawnInfo, error) {
	var env struct {
		Info *SpawnInfo `json:"info"`
	}
	if err := m.get(ctx, "machine/active", nil, &env); err != nil {
		return nil, err
	}
	return env.Info, nil
}

// GetRecommended returns the caller's two recommendation cards.
func (m *Client) GetRecommended(ctx context.Context) (*Recommendations, error) {
	var out Recommendations
	if err := m.get(ctx, "machine/recommended", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetInfo looks up a machine by its case-sensitive display name.
func (m *Client) GetInfo(ctx context.Context, name string) (*MachineInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("machines: GetInfo: name is required")
	}
	var env struct {
		Info MachineInfo `json:"info"`
	}
	path := "machine/profile/" + url.PathEscape(name)
	if err := m.get(ctx, path, nil, &env); err != nil {
		return nil, err
	}
	return &env.Info, nil
}

// GetWalkthroughs returns the walkthroughs metadata bundle for a machine id.
func (m *Client) GetWalkthroughs(ctx context.Context, machineID int64) (*WalkthroughsBundle, error) {
	if machineID <= 0 {
		return nil, fmt.Errorf("machines: GetWalkthroughs: machineID must be positive")
	}
	var env struct {
		Message WalkthroughsBundle `json:"message"`
	}
	path := "machine/walkthroughs/" + strconv.FormatInt(machineID, 10)
	if err := m.get(ctx, path, nil, &env); err != nil {
		return nil, err
	}
	return &env.Message, nil
}

// ListReviews returns one page of reviews for a machine.
func (m *Client) ListReviews(ctx context.Context, machineID int64, opts ListReviewsOptions) (*ReviewPage, error) {
	if machineID <= 0 {
		return nil, fmt.Errorf("machines: ListReviews: machineID must be positive")
	}
	q := url.Values{}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.SortType != "" {
		q.Set("sort_type", opts.SortType)
	}
	for _, s := range opts.SortBy {
		q.Add("sort_by[]", s)
	}
	path := "review/machine/" + strconv.FormatInt(machineID, 10) + "/paginated"
	var out ReviewPage
	if err := m.get(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- internals ----------

func (m *Client) get(ctx context.Context, path string, q url.Values, out any) error {
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	req, err := m.c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	resp, err := m.c.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("machines: %s: auth failed (%d)", path, resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("machines: %s: not found", path)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("machines: %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("machines: %s: decoding response: %w", path, err)
	}
	return nil
}
