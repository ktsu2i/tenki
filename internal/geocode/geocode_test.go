package geocode

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSearchReturnsFirstResult(t *testing.T) {
	httpClient := &fakeHTTPClient{
		statusCode: http.StatusOK,
		body:       `{"results":[{"name":"Tokyo","country":"Japan","latitude":35.6895,"longitude":139.6917,"timezone":"Asia/Tokyo"}]}`,
	}

	client := NewClient(httpClient)
	client.Endpoint = "https://example.test/search"

	location, err := client.Search(context.Background(), "tokyo")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if got := httpClient.request.URL.Query().Get("name"); got != "tokyo" {
		t.Fatalf("name query = %q, want tokyo", got)
	}
	if got := httpClient.request.URL.Query().Get("count"); got != "1" {
		t.Fatalf("count query = %q, want 1", got)
	}
	if location.Name != "Tokyo" || location.Country != "Japan" || location.Timezone != "Asia/Tokyo" {
		t.Fatalf("location = %+v, want Tokyo, Japan", location)
	}
}

func TestSearchRejectsNoResults(t *testing.T) {
	client := NewClient(&fakeHTTPClient{
		statusCode: http.StatusOK,
		body:       `{"results":[]}`,
	})
	client.Endpoint = "https://example.test/search"

	if _, err := client.Search(context.Background(), "missing"); err == nil {
		t.Fatal("Search returned nil error, want not found error")
	}
}

type fakeHTTPClient struct {
	statusCode int
	body       string
	request    *http.Request
}

func (c *fakeHTTPClient) Do(request *http.Request) (*http.Response, error) {
	c.request = request
	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader(c.body)),
		Header:     make(http.Header),
		Request:    request,
	}, nil
}
