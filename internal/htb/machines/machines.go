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

// Language is one entry in the walkthrough-language enum.
type Language struct {
	ID        int64  `json:"id"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

// MachineRef is a minimal machine reference (id + display fields).
type MachineRef struct {
	ID     int64  `json:"id"`
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

// MatrixScores is the five-axis radar block.
type MatrixScores struct {
	CTF    float64 `json:"ctf"`
	Custom float64 `json:"custom"`
	CVE    float64 `json:"cve"`
	Enum   float64 `json:"enum"`
	Real   float64 `json:"real"`
}

// GraphMatrix is the per-machine difficulty radar triple.
type GraphMatrix struct {
	Aggregate MatrixScores `json:"aggregate"`
	Maker     MatrixScores `json:"maker"`
	User      MatrixScores `json:"user"`
}

// KindRef is HTB's small {id, text} discriminator used in tasks and adventure.
type KindRef struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// MachineTask is one guided-mode task.
// SENSITIVITY: the Flag field contains the actual flag value once
// the caller has completed the task; callers must treat it accordingly.
type MachineTask struct {
	ID             int64                    `json:"id"`
	Title          string                   `json:"title,omitempty"`
	Description    string                   `json:"description,omitempty"`
	Hint           *string                  `json:"hint,omitempty"`
	Flag           string                   `json:"flag,omitempty"`
	MaskedFlag     string                   `json:"masked_flag,omitempty"`
	Options        []map[string]interface{} `json:"options,omitempty"`
	PrerequisiteID *int64                   `json:"prerequisite_id,omitempty"`
	Completed      bool                     `json:"completed"`
	TaskType       KindRef                  `json:"task_type"`
	Type           KindRef                  `json:"type"`
}

// AdventureStep is one adventure-mode step.
// Unlike MachineTask, the Flag field is a textual completion indicator
// ("User flag owned"), not the flag value.
type AdventureStep struct {
	ID          *int64  `json:"id,omitempty"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Hint        *string `json:"hint,omitempty"`
	Flag        string  `json:"flag,omitempty"`
	MaskedFlag  string  `json:"masked_flag,omitempty"`
	FlagRating  int     `json:"flag_rating,omitempty"`
	Completed   bool    `json:"completed"`
	Type        KindRef `json:"type"`
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

// ListWalkthroughLanguages returns the language enum used for walkthroughs.
func (m *Client) ListWalkthroughLanguages(ctx context.Context) ([]Language, error) {
	var env struct {
		Info []Language `json:"info"`
	}
	if err := m.get(ctx, "machine/walkthroughs/language/list", nil, &env); err != nil {
		return nil, err
	}
	return env.Info, nil
}

// GetRandomWalkthroughMachine returns a random machine that has community walkthroughs.
func (m *Client) GetRandomWalkthroughMachine(ctx context.Context) (*MachineRef, error) {
	var env struct {
		Message MachineRef `json:"message"`
	}
	if err := m.get(ctx, "machine/walkthrough/random", nil, &env); err != nil {
		return nil, err
	}
	return &env.Message, nil
}

// GetGraphMatrix returns the difficulty radar matrix for a machine.
func (m *Client) GetGraphMatrix(ctx context.Context, machineID int64) (*GraphMatrix, error) {
	if machineID <= 0 {
		return nil, fmt.Errorf("machines: GetGraphMatrix: machineID must be positive")
	}
	var env struct {
		Info GraphMatrix `json:"info"`
	}
	path := "machine/graph/matrix/" + strconv.FormatInt(machineID, 10)
	if err := m.get(ctx, path, nil, &env); err != nil {
		return nil, err
	}
	return &env.Info, nil
}

// ListTasks returns the guided-mode tasks for a machine. Uses the plural-
// prefix route /machines/{id}/tasks (yes, the API mixes prefixes).
func (m *Client) ListTasks(ctx context.Context, machineID int64) ([]MachineTask, error) {
	if machineID <= 0 {
		return nil, fmt.Errorf("machines: ListTasks: machineID must be positive")
	}
	var env struct {
		Data []MachineTask `json:"data"`
	}
	path := "machines/" + strconv.FormatInt(machineID, 10) + "/tasks"
	if err := m.get(ctx, path, nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// ListAdventureSteps returns the adventure-mode steps for a machine.
func (m *Client) ListAdventureSteps(ctx context.Context, machineID int64) ([]AdventureStep, error) {
	if machineID <= 0 {
		return nil, fmt.Errorf("machines: ListAdventureSteps: machineID must be positive")
	}
	var env struct {
		Data []AdventureStep `json:"data"`
	}
	path := "machines/" + strconv.FormatInt(machineID, 10) + "/adventure"
	if err := m.get(ctx, path, nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// DownloadOfficialWriteupPDF downloads the official PDF writeup for a
// machine and writes it to w. Returns the number of bytes copied.
// Unlike the JSON-returning methods, this one streams the body directly
// — the response is application/pdf, not JSON.
func (m *Client) DownloadOfficialWriteupPDF(ctx context.Context, machineID int64, w io.Writer) (int64, error) {
	if machineID <= 0 {
		return 0, fmt.Errorf("machines: DownloadOfficialWriteupPDF: machineID must be positive")
	}
	if w == nil {
		return 0, fmt.Errorf("machines: DownloadOfficialWriteupPDF: writer is nil")
	}
	path := "machine/writeup/" + strconv.FormatInt(machineID, 10)
	req, err := m.c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}
	// Override Accept so a content-negotiating proxy doesn't try to give us JSON.
	req.Header.Set("Accept", "application/pdf, */*")
	resp, err := m.c.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("machines: writeup %d: not found", machineID)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return 0, fmt.Errorf("machines: writeup %d: auth failed (%d)", machineID, resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return 0, fmt.Errorf("machines: writeup %d: status %d: %s", machineID, resp.StatusCode, string(body))
	}
	return io.Copy(w, resp.Body)
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
