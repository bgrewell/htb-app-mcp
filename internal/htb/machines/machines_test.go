package machines

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bgrewell/htb-app-mcp/internal/htb"
)

// fixtureBody reads the captured fixture and returns just the response body
// as a raw JSON byte slice — what the real server would have sent.
func fixtureBody(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "..", "scripts", "capture", "fixtures", "machines", name)
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var wrap struct {
		Response struct {
			Body json.RawMessage `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return wrap.Response.Body
}

// newFixtureClient returns a Client whose underlying HTTP server serves
// each path -> fixture mapping. The captured request expects to hit
// /api/v4/<path>, so we mount under that prefix.
func newFixtureClient(t *testing.T, fixtures map[string]string) (*Client, func()) {
	t.Helper()
	mux := http.NewServeMux()
	for path, fxName := range fixtures {
		body := fixtureBody(t, fxName)
		mux.HandleFunc("/api/v4/"+strings.TrimLeft(path, "/"), func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		})
	}
	srv := httptest.NewServer(mux)
	hc, err := htb.New(htb.Config{
		Token:          "test-token",
		BaseURL:        srv.URL + "/api/v4",
		HTTPClient:     srv.Client(),
		RequestTimeout: 2 * time.Second,
		MaxRetries:     1,
	})
	if err != nil {
		srv.Close()
		t.Fatalf("htb.New: %v", err)
	}
	return New(hc), srv.Close
}

func TestGetActiveSpawn(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/active": "list_active.json",
	})
	defer cleanup()

	s, err := c.GetActiveSpawn(context.Background())
	if err != nil {
		t.Fatalf("GetActiveSpawn: %v", err)
	}
	if s == nil {
		t.Fatal("expected a spawn, got nil")
	}
	if s.ID != 900 || s.Name != "Reactor" {
		t.Errorf("spawn = {ID:%d, Name:%q}, want {900, Reactor}", s.ID, s.Name)
	}
	if s.IP == "" || s.LabServer == "" {
		t.Errorf("expected IP and LabServer to be populated, got %+v", s)
	}
}

func TestGetActiveSpawn_None(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"info":null}`)
	}))
	defer srv.Close()
	hc, _ := htb.New(htb.Config{
		Token: "t", BaseURL: srv.URL + "/api/v4", HTTPClient: srv.Client(), RequestTimeout: time.Second, MaxRetries: 1,
	})

	s, err := New(hc).GetActiveSpawn(context.Background())
	if err != nil {
		t.Fatalf("GetActiveSpawn: %v", err)
	}
	if s != nil {
		t.Errorf("expected nil spawn, got %+v", s)
	}
}

func TestGetRecommended(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/recommended": "list_recommended.json",
	})
	defer cleanup()

	r, err := c.GetRecommended(context.Background())
	if err != nil {
		t.Fatalf("GetRecommended: %v", err)
	}
	if r.Card1.ID == 0 || r.Card2.ID == 0 {
		t.Errorf("recommended cards missing IDs: %+v", r)
	}
	if len(r.State) != 2 {
		t.Errorf("State = %v, want 2 entries", r.State)
	}
	if r.Card1.Name != "Reactor" {
		t.Errorf("Card1.Name = %q, want Reactor", r.Card1.Name)
	}
}

func TestGetInfo(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/profile/Cap": "get_info_by_name.json",
	})
	defer cleanup()

	info, err := c.GetInfo(context.Background(), "Cap")
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.ID != 351 || info.Name != "Cap" {
		t.Errorf("info = {ID:%d, Name:%q}, want {351, Cap}", info.ID, info.Name)
	}
	if info.OS != "Linux" {
		t.Errorf("OS = %q, want Linux", info.OS)
	}
	if !info.Retired || !info.Free {
		t.Errorf("expected Retired+Free, got %+v", *info)
	}
	if info.Maker == nil || info.Maker.ID == 0 {
		t.Errorf("expected Maker to be populated")
	}
	if info.FeedbackForChart == nil {
		t.Error("expected FeedbackForChart to be populated")
	}
	if info.PlayInfo == nil {
		t.Error("expected PlayInfo to be populated")
	}
}

func TestGetInfo_RejectsEmptyName(t *testing.T) {
	c := New(nil)
	if _, err := c.GetInfo(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetWalkthroughs(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/walkthroughs/351": "list_walkthroughs.json",
	})
	defer cleanup()

	w, err := c.GetWalkthroughs(context.Background(), 351)
	if err != nil {
		t.Fatalf("GetWalkthroughs: %v", err)
	}
	if w.Official == nil || w.Official.Filename != "Cap.pdf" {
		t.Errorf("Official = %+v, want filename Cap.pdf", w.Official)
	}
	if w.Video == nil || w.Video.YoutubeID == nil || *w.Video.YoutubeID != "O_z6o2xuvlw" {
		yid := ""
		if w.Video != nil && w.Video.YoutubeID != nil {
			yid = *w.Video.YoutubeID
		}
		t.Errorf("Video youtube_id = %q, want O_z6o2xuvlw", yid)
	}
	if len(w.Writeups) == 0 {
		t.Fatal("expected at least one community writeup")
	}
	first := w.Writeups[0]
	if first.URL == "" || first.UserName == "" {
		t.Errorf("first writeup missing URL or UserName: %+v", first)
	}
}

func TestGetWalkthroughs_RejectsBadID(t *testing.T) {
	c := New(nil)
	if _, err := c.GetWalkthroughs(context.Background(), 0); err == nil {
		t.Fatal("expected error for machineID=0")
	}
}

func TestListReviews(t *testing.T) {
	var lastQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastQuery = r.URL.RawQuery
		_, _ = w.Write(fixtureBody(t, "list_reviews.json"))
	}))
	defer srv.Close()
	hc, _ := htb.New(htb.Config{
		Token: "t", BaseURL: srv.URL + "/api/v4", HTTPClient: srv.Client(), RequestTimeout: time.Second, MaxRetries: 1,
	})
	c := New(hc)

	p, err := c.ListReviews(context.Background(), 351, ListReviewsOptions{
		PerPage:  15,
		SortType: "desc",
		SortBy:   []string{"created_at"},
	})
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	if !strings.Contains(lastQuery, "per_page=15") || !strings.Contains(lastQuery, "sort_type=desc") {
		t.Errorf("query missing expected params: %q", lastQuery)
	}
	if !strings.Contains(lastQuery, "sort_by%5B%5D=created_at") {
		t.Errorf("query missing sort_by[] encoding: %q", lastQuery)
	}
	if p.Meta.LastPage <= 0 || p.Meta.CurrentPage != 1 {
		t.Errorf("meta = %+v, want CurrentPage=1 and LastPage>0", p.Meta)
	}
	if p.Count == 0 || p.Average == 0 {
		t.Errorf("count/average not populated: count=%d avg=%v", p.Count, p.Average)
	}
}

func TestListReviews_RejectsBadID(t *testing.T) {
	c := New(nil)
	if _, err := c.ListReviews(context.Background(), 0, ListReviewsOptions{}); err == nil {
		t.Fatal("expected error for machineID=0")
	}
}

func TestListWalkthroughLanguages(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/walkthroughs/language/list": "list_walkthrough_languages.json",
	})
	defer cleanup()

	langs, err := c.ListWalkthroughLanguages(context.Background())
	if err != nil {
		t.Fatalf("ListWalkthroughLanguages: %v", err)
	}
	if len(langs) == 0 {
		t.Fatal("expected at least one language")
	}
	if langs[0].FullName == "" || langs[0].ShortName == "" {
		t.Errorf("first language missing fields: %+v", langs[0])
	}
}

func TestGetRandomWalkthroughMachine(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/walkthrough/random": "get_walkthrough_random.json",
	})
	defer cleanup()

	ref, err := c.GetRandomWalkthroughMachine(context.Background())
	if err != nil {
		t.Fatalf("GetRandomWalkthroughMachine: %v", err)
	}
	if ref.ID == 0 || ref.Name == "" {
		t.Errorf("ref = %+v, want non-zero id and name", ref)
	}
}

func TestGetGraphMatrix(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machine/graph/matrix/351": "get_graph_matrix.json",
	})
	defer cleanup()

	gm, err := c.GetGraphMatrix(context.Background(), 351)
	if err != nil {
		t.Fatalf("GetGraphMatrix: %v", err)
	}
	if gm.Aggregate.CTF == 0 && gm.Aggregate.Custom == 0 && gm.Aggregate.Real == 0 {
		t.Errorf("aggregate scores all zero: %+v", gm.Aggregate)
	}
}

func TestGetGraphMatrix_RejectsBadID(t *testing.T) {
	c := New(nil)
	if _, err := c.GetGraphMatrix(context.Background(), 0); err == nil {
		t.Fatal("expected error for machineID=0")
	}
}

func TestListTasks(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machines/351/tasks": "list_machine_tasks.json",
	})
	defer cleanup()

	tasks, err := c.ListTasks(context.Background(), 351)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) < 2 {
		t.Fatalf("expected several tasks, got %d", len(tasks))
	}
	if tasks[0].Title == "" || tasks[0].Description == "" || tasks[0].MaskedFlag == "" {
		t.Errorf("first task missing fields: %+v", tasks[0])
	}
	// Second task should chain off the first via prerequisite_id.
	if tasks[1].PrerequisiteID == nil || *tasks[1].PrerequisiteID != tasks[0].ID {
		t.Errorf("expected task[1].prerequisite_id == task[0].id, got %+v", tasks[1].PrerequisiteID)
	}
}

func TestListTasks_RejectsBadID(t *testing.T) {
	c := New(nil)
	if _, err := c.ListTasks(context.Background(), 0); err == nil {
		t.Fatal("expected error for machineID=0")
	}
}

func TestListAdventureSteps(t *testing.T) {
	c, cleanup := newFixtureClient(t, map[string]string{
		"machines/351/adventure": "get_machine_adventure.json",
	})
	defer cleanup()

	steps, err := c.ListAdventureSteps(context.Background(), 351)
	if err != nil {
		t.Fatalf("ListAdventureSteps: %v", err)
	}
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 adventure steps, got %d", len(steps))
	}
	if steps[0].Title != "Submit User Flag" || steps[0].Type.Text != "user" {
		t.Errorf("step[0] = %+v, want title=Submit User Flag type.text=user", steps[0])
	}
	if steps[1].Title != "Submit Root Flag" {
		t.Errorf("step[1].Title = %q, want Submit Root Flag", steps[1].Title)
	}
}

func TestDownloadOfficialWriteupPDF(t *testing.T) {
	pdfBody := []byte("%PDF-1.4\nfake pdf bytes for test\n%%EOF\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); !strings.Contains(got, "application/pdf") {
			t.Errorf("Accept = %q, want it to contain application/pdf", got)
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdfBody)
	}))
	defer srv.Close()
	hc, _ := htb.New(htb.Config{
		Token: "t", BaseURL: srv.URL + "/api/v4", HTTPClient: srv.Client(), RequestTimeout: time.Second, MaxRetries: 1,
	})

	var buf strings.Builder
	n, err := New(hc).DownloadOfficialWriteupPDF(context.Background(), 351, &buf)
	if err != nil {
		t.Fatalf("DownloadOfficialWriteupPDF: %v", err)
	}
	if int(n) != len(pdfBody) || buf.String() != string(pdfBody) {
		t.Errorf("got %d bytes, want %d; body match=%v", n, len(pdfBody), buf.String() == string(pdfBody))
	}
}

func TestDownloadOfficialWriteupPDF_Rejects(t *testing.T) {
	c := New(nil)
	if _, err := c.DownloadOfficialWriteupPDF(context.Background(), 0, &strings.Builder{}); err == nil {
		t.Fatal("expected error for machineID=0")
	}
	if _, err := c.DownloadOfficialWriteupPDF(context.Background(), 1, nil); err == nil {
		t.Fatal("expected error for nil writer")
	}
}

func TestGet_HandlesNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"Not found."}`)
	}))
	defer srv.Close()
	hc, _ := htb.New(htb.Config{
		Token: "t", BaseURL: srv.URL + "/api/v4", HTTPClient: srv.Client(), RequestTimeout: time.Second, MaxRetries: 1,
	})

	_, err := New(hc).GetInfo(context.Background(), "DoesNotExist")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}
