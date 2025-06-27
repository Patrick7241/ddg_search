package ddg_search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrRatelimit     = errors.New("rate limit exceeded")
	ErrTimeout       = errors.New("request timeout")
	ErrSearch        = errors.New("search error")
	ErrInvalidParams = errors.New("invalid parameters")
)

type SafeSearchLevel string

const (
	SafeSearchOn       SafeSearchLevel = "on"
	SafeSearchModerate SafeSearchLevel = "moderate"
	SafeSearchOff      SafeSearchLevel = "off"
)

type Backend string

const (
	BackendAuto Backend = "auto"
	BackendHTML Backend = "html"
	BackendLite Backend = "lite"
)

type Timelimit string

const (
	// TimelimitDay restricts results to the past day
	TimelimitDay Timelimit = "d"
	// TimelimitWeek restricts results to the past week
	TimelimitWeek Timelimit = "w"
	// TimelimitMonth restricts results to the past month
	TimelimitMonth Timelimit = "m"
	// TimelimitYear restricts results to the past year
	TimelimitYear Timelimit = "y"
	// TimelimitAll means no time restriction
	TimelimitAll Timelimit = ""
)

type resolution string

const (
	ResolutionHigh     resolution = "high"
	ResolutionStandard resolution = "standard"
	ResolutionAll      resolution = ""
)

type durationTime string

const (
	DurationShort  durationTime = "short"
	DurationMedium durationTime = "medium"
	DurationLong   durationTime = "long"
	DurationAll    durationTime = ""
)

type licenseVideos string

const (
	LicenseCreativeCommon licenseVideos = "creativeCommon"
	LicenseYouTube        licenseVideos = "youtube"
	LicenseAll            licenseVideos = ""
)

type DDGS struct {
	client         *http.Client
	headers        map[string]string
	proxy          string
	timeout        time.Duration
	sleepTimestamp time.Time
	sleepDuration  time.Duration
	mu             sync.Mutex
}

// NewDDGS creates a new DDGS instance with optional configuration
func NewDDGS(options ...func(*DDGS)) *DDGS {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	ddgs := &DDGS{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jar,
		},
		headers: map[string]string{
			"Referer": "https://duckduckgo.com/",
		},
		timeout:       10 * time.Second,
		sleepDuration: 1500 * time.Millisecond,
	}

	for _, option := range options {
		option(ddgs)
	}

	if ddgs.proxy == "" {
		ddgs.proxy = os.Getenv("DDGS_PROXY")
	}

	if ddgs.proxy != "" {
		ddgs.client.Transport = &http.Transport{
			Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: ddgs.proxy}),
		}
	}

	return ddgs
}

// WithHeaders sets custom headers for the DDGS client
func WithHeaders(headers map[string]string) func(*DDGS) {
	return func(d *DDGS) {
		for k, v := range headers {
			d.headers[k] = v
		}
	}
}

// WithProxy sets the proxy for the DDGS client
func WithProxy(proxy string) func(*DDGS) {
	return func(d *DDGS) {
		d.proxy = proxy
	}
}

// WithTimeout sets the request timeout for the DDGS client
func WithTimeout(timeout time.Duration) func(*DDGS) {
	return func(d *DDGS) {
		d.timeout = timeout
	}
}

// WithSleepDuration sets the sleep duration between requests for rate limiting
func WithSleepDuration(d time.Duration) func(*DDGS) {
	return func(ddgs *DDGS) {
		ddgs.sleepDuration = d
	}
}

// sleep implements rate limiting between requests
func (d *DDGS) sleep() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.sleepTimestamp.IsZero() {
		d.sleepTimestamp = time.Now()
		return
	}

	elapsed := time.Since(d.sleepTimestamp)
	if elapsed < 20*time.Second {
		time.Sleep(d.sleepDuration)
	}

	d.sleepTimestamp = time.Now()
}

// doRequest performs the HTTP request with rate limiting and timeout
func (d *DDGS) doRequest(req *http.Request) (*http.Response, error) {
	d.sleep()
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	req = req.WithContext(ctx)

	for k, v := range d.headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrSearch, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return resp, nil
	case http.StatusAccepted, http.StatusMovedPermanently, http.StatusForbidden,
		http.StatusBadRequest, http.StatusTooManyRequests, http.StatusTeapot:
		return nil, ErrRatelimit
	default:
		return nil, fmt.Errorf("%w: status %d", ErrSearch, resp.StatusCode)
	}
}

// getVQD retrieves the VQD token required for some DuckDuckGo requests
func (d *DDGS) getVQD(keywords string) (string, error) {
	req, _ := http.NewRequest("GET", "https://duckduckgo.com", nil)
	q := req.URL.Query()
	q.Add("q", keywords)
	req.URL.RawQuery = q.Encode()

	resp, err := d.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`vqd\s*=\s*["']?([\d-]+)["']?`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", ErrSearch
	}
	return matches[1], nil
}

// Images performs image search on DuckDuckGo
func (d *DDGS) Images(
	keywords string,
	region string,
	safesearch SafeSearchLevel,
	timelimit Timelimit,
	maxResults int,
) ([]map[string]interface{}, error) {
	vqd, err := d.getVQD(keywords)
	if err != nil {
		return nil, err
	}

	safesearchMap := map[SafeSearchLevel]string{
		SafeSearchOn:       "1",
		SafeSearchModerate: "1",
		SafeSearchOff:      "-1",
	}

	params := url.Values{}
	params.Set("o", "json")
	params.Set("q", keywords)
	params.Set("l", region)
	params.Set("vqd", vqd)
	params.Set("p", safesearchMap[safesearch])

	if timelimit != "" {
		params.Set("f", "time:"+string(timelimit))
	}

	var results []map[string]interface{}
	seen := map[string]struct{}{}

	for i := 0; i < 5; i++ {
		apiURL := fmt.Sprintf("https://duckduckgo.com/i.js?%s", params.Encode())
		req, _ := http.NewRequest("GET", apiURL, nil)

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		req.Header.Set("Referer", "https://duckduckgo.com/")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Sec-Fetch-Mode", "cors")

		resp, err := d.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var respData struct {
			Results []map[string]interface{} `json:"results"`
			Next    string                   `json:"next"`
		}

		if err := json.Unmarshal(body, &respData); err != nil {
			return nil, fmt.Errorf("json unmarshal error: %v", err)
		}

		for _, item := range respData.Results {
			imageURL, ok := item["image"].(string)
			if !ok || imageURL == "" {
				continue
			}
			if _, exists := seen[imageURL]; exists {
				continue
			}
			seen[imageURL] = struct{}{}

			results = append(results, item)

			if maxResults > 0 && len(results) >= maxResults {
				return results, nil
			}
		}

		if respData.Next == "" || maxResults == 0 {
			break
		}
		nextS := extractNextS(respData.Next)
		if nextS != "" {
			params.Set("s", nextS)
		}
	}

	return results, nil
}

// News performs news search on DuckDuckGo
func (d *DDGS) News(
	keywords string,
	region string,
	safesearch SafeSearchLevel,
	timelimit Timelimit, // d, w, m
	maxResults int,
) ([]map[string]interface{}, error) {
	if keywords == "" {
		return nil, fmt.Errorf("keywords is mandatory")
	}

	// Get VQD token
	vqd, err := d.getVQD(keywords)
	if err != nil {
		return nil, err
	}

	// Safesearch mapping
	safesearchMap := map[SafeSearchLevel]string{
		SafeSearchOn:       "1",
		SafeSearchModerate: "-1",
		SafeSearchOff:      "-2",
	}

	// Build query params
	params := url.Values{}
	params.Set("o", "json")
	params.Set("q", keywords)
	params.Set("l", region)
	params.Set("vqd", vqd)
	params.Set("noamp", "1")
	params.Set("p", safesearchMap[safesearch])
	if timelimit != "" {
		params.Set("df", string(timelimit))
	}

	// Cache for deduplication
	seen := map[string]struct{}{}
	var results []map[string]interface{}

	for i := 0; i < 5; i++ {
		apiURL := fmt.Sprintf("https://duckduckgo.com/news.js?%s", params.Encode())
		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		req.Header.Set("Referer", "https://duckduckgo.com/")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Sec-Fetch-Mode", "cors")

		resp, err := d.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Debug: Uncomment if needed
		// fmt.Println("DEBUG Response:", string(body))

		// Parse JSON
		var respData struct {
			Results []map[string]interface{} `json:"results"`
			Next    string                   `json:"next"`
		}
		if err := json.Unmarshal(body, &respData); err != nil {
			return nil, fmt.Errorf("json unmarshal error: %v", err)
		}

		for _, item := range respData.Results {
			urlStr, ok := item["url"].(string)
			if !ok || urlStr == "" {
				continue
			}
			if _, exists := seen[urlStr]; exists {
				continue
			}
			seen[urlStr] = struct{}{}

			// Convert timestamp
			dateInt, _ := item["date"].(float64)
			dateStr := ""
			if dateInt > 0 {
				date := time.Unix(int64(dateInt), 0).UTC()
				dateStr = date.Format(time.RFC3339)
			}

			// Build result map
			result := map[string]interface{}{
				"date":   dateStr,
				"title":  item["title"],
				"body":   item["excerpt"],
				"url":    item["url"],
				"image":  item["image"],
				"source": item["source"],
			}
			results = append(results, result)

			if maxResults > 0 && len(results) >= maxResults {
				return results, nil
			}
		}

		// No next page
		if respData.Next == "" || maxResults == 0 {
			break
		}

		// Extract next s parameter
		nextS := extractNextS(respData.Next)
		if nextS != "" {
			params.Set("s", nextS)
		}
	}

	return results, nil
}

// Videos performs video search on DuckDuckGo
func (d *DDGS) Videos(
	keywords string,
	region string,
	safesearch SafeSearchLevel,
	timelimit Timelimit,
	resolution resolution,
	duration durationTime,
	licenseVideos licenseVideos,
	maxResults int,
) ([]map[string]interface{}, error) {
	if keywords == "" {
		return nil, fmt.Errorf("keywords is mandatory")
	}

	// Get VQD token
	vqd, err := d.getVQD(keywords)
	if err != nil {
		return nil, err
	}

	// Safesearch mapping
	safesearchMap := map[SafeSearchLevel]string{
		SafeSearchOn:       "1",
		SafeSearchModerate: "-1",
		SafeSearchOff:      "-2",
	}

	//Build filters
	var filters []string
	if timelimit != "" {
		filters = append(filters, "publishedAfter:"+string(timelimit))
	}
	if resolution != ResolutionAll {
		filters = append(filters, "videoDefinition:"+string(resolution))
	}
	if duration != DurationAll {
		filters = append(filters, "videoDuration:"+string(duration))
	}
	if licenseVideos != LicenseAll {
		filters = append(filters, "videoLicense:"+string(licenseVideos))
	}
	// Build query params
	params := url.Values{}
	params.Set("o", "json")
	params.Set("q", keywords)
	params.Set("l", region)
	params.Set("vqd", vqd)
	params.Set("p", safesearchMap[safesearch])
	params.Set("f", strings.Join(filters, ","))

	// Deduplication cache
	seen := map[string]struct{}{}
	var results []map[string]interface{}

	for i := 0; i < 8; i++ {
		apiURL := fmt.Sprintf("https://duckduckgo.com/v.js?%s", params.Encode())
		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		req.Header.Set("Referer", "https://duckduckgo.com/")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Sec-Fetch-Mode", "cors")

		resp, err := d.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// fmt.Println("DEBUG Response:", string(body))

		var respData struct {
			Results []map[string]interface{} `json:"results"`
			Next    string                   `json:"next"`
		}
		if err := json.Unmarshal(body, &respData); err != nil {
			return nil, fmt.Errorf("json unmarshal error: %v", err)
		}

		for _, item := range respData.Results {
			contentID, ok := item["content"].(string)
			if !ok || contentID == "" {
				continue
			}
			if _, exists := seen[contentID]; exists {
				continue
			}
			seen[contentID] = struct{}{}

			results = append(results, item)

			if maxResults > 0 && len(results) >= maxResults {
				return results, nil
			}
		}

		// No more pages
		if respData.Next == "" || maxResults == 0 {
			break
		}

		// Pagination: extract "s" param
		nextS := extractNextS(respData.Next)
		if nextS != "" {
			params.Set("s", nextS)
		}
	}

	return results, nil
}

func extractNextS(next string) string {
	u, err := url.Parse(next)
	if err != nil {
		return ""
	}
	vals := u.Query()
	return vals.Get("s")
}

// Text performs text search on DuckDuckGo
func (d *DDGS) Text(
	keywords string,
	region string,
	safesearch SafeSearchLevel,
	timelimit Timelimit,
	backend Backend,
	maxResults int,
) ([]map[string]string, error) {
	if region == "" {
		region = "wt-wt"
	}
	if keywords == "" {
		return nil, ErrInvalidParams
	}

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	var results []map[string]string
	var err error

	switch backend {
	case BackendAuto:
		if rng.Intn(2) == 0 {
			results, err = d.textHTML(keywords, region, timelimit, maxResults, safesearch)
			if err != nil {
				results, err = d.textLite(keywords, region, timelimit, maxResults, safesearch)
			}
		} else {
			results, err = d.textLite(keywords, region, timelimit, maxResults, safesearch)
			if err != nil {
				results, err = d.textHTML(keywords, region, timelimit, maxResults, safesearch)
			}
		}
	case BackendHTML:
		results, err = d.textHTML(keywords, region, timelimit, maxResults, safesearch)
	case BackendLite:
		results, err = d.textLite(keywords, region, timelimit, maxResults, safesearch)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", backend)
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

// textHTML performs search using the HTML backend
func (d *DDGS) textHTML(
	keywords string,
	region string,
	timelimit Timelimit,
	maxResults int,
	safesearch SafeSearchLevel,
) ([]map[string]string, error) {
	headers := map[string]string{
		"Referer":        "https://html.duckduckgo.com/",
		"Sec-Fetch-User": "?1",
	}
	for k, v := range headers {
		d.headers[k] = v
	}

	payload := url.Values{
		"q":  []string{keywords},
		"b":  []string{""},
		"kl": []string{region},
	}
	payload = d.setSafeSearch(safesearch, payload)

	if timelimit != "" {
		payload.Add("df", string(timelimit))
	}

	req, _ := http.NewRequest("POST", "https://html.duckduckgo.com/html", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cache := make(map[string]bool)
	var results []map[string]string

	for i := 0; i < 5; i++ {
		if maxResults > 0 && len(results) >= maxResults {
			break
		}
		resp, err := d.doRequest(req)
		if err != nil {
			return nil, err
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if strings.Contains(doc.Text(), "No results.") {
			return results, nil
		}

		doc.Find("div.result").Each(func(_ int, s *goquery.Selection) {
			if maxResults > 0 && len(results) >= maxResults {
				return
			}
			title := strings.TrimSpace(s.Find("h2").Text())
			href, _ := s.Find("a.result__url").Attr("href")
			body := strings.TrimSpace(s.Find("a.result__snippet").Text())

			if href != "" && !cache[href] && !strings.HasPrefix(href, "http://www.google.com/search?q=") {
				cache[href] = true
				result := map[string]string{
					"title": normalize(title),
					"href":  normalizeURL(href),
					"body":  normalize(body),
				}
				results = append(results, result)
			}
		})
		nextPage := doc.Find("div.nav-link").Last()
		if nextPage.Length() == 0 {
			break
		}

		nextPage.Find("input[type='hidden']").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			value, _ := s.Attr("value")
			payload.Set(name, value)
		})
		req, _ = http.NewRequest("POST", "https://html.duckduckgo.com/html", strings.NewReader(payload.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return results, nil
}

// textLite performs search using the Lite backend
func (d *DDGS) textLite(
	keywords string,
	region string,
	timelimit Timelimit,
	maxResults int,
	safesearch SafeSearchLevel,
) ([]map[string]string, error) {
	headers := map[string]string{
		"Referer":        "https://lite.duckduckgo.com/",
		"Sec-Fetch-User": "?1",
	}
	for k, v := range headers {
		d.headers[k] = v
	}

	payload := url.Values{
		"q":  []string{keywords},
		"kl": []string{region},
	}
	if timelimit != "" {
		payload.Add("df", string(timelimit))
	}

	payload = d.setSafeSearch(safesearch, payload)

	cache := make(map[string]bool)
	var results []map[string]string

	req, _ := http.NewRequest("POST", "https://lite.duckduckgo.com/lite/", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for i := 0; i < 5; i++ {
		if maxResults > 0 && len(results) >= maxResults {
			break
		}
		resp, err := d.doRequest(req)
		if err != nil {
			return nil, err
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if strings.Contains(doc.Text(), "No more results.") {
			break
		}

		var href, title, body string
		rows := doc.Find("table").Last().Find("tr")

		rows.Each(func(i int, s *goquery.Selection) {
			if maxResults > 0 && len(results) >= maxResults {
				return
			}
			mod := i % 4
			switch mod {
			case 0:
				link := s.Find("a")
				href, _ = link.Attr("href")
				title = strings.TrimSpace(link.Text())
				if href == "" || cache[href] || strings.HasPrefix(href, "http://www.google.com/search?q=") || strings.Contains(href, "duckduckgo.com/y.js?ad_domain") {
					href = ""
					title = ""
				} else {
					cache[href] = true
				}
			case 1:
				if href != "" {
					body = strings.TrimSpace(s.Find("td.result-snippet").Text())
					results = append(results, map[string]string{
						"title": normalize(title),
						"href":  normalizeURL(href),
						"body":  normalize(body),
					})
				}
			}
		})
		nextForm := doc.Find(`form:has(input[value*="ext"])`).Last()
		if nextForm.Length() == 0 {
			break
		}

		nextPayload := url.Values{}
		nextForm.Find(`input[type="hidden"]`).Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			value, _ := s.Attr("value")
			if name != "" {
				nextPayload.Set(name, value)
			}
		})

		payload = nextPayload
	}

	return results, nil
}

// setSafeSearch configures the safe search parameter in the request payload
func (d *DDGS) setSafeSearch(safesearch SafeSearchLevel, payload url.Values) url.Values {
	switch safesearch {
	case SafeSearchOn:
		payload.Set("p", "1")
	case SafeSearchModerate:
		payload.Set("p", "-1")
	case SafeSearchOff:
		payload.Set("p", "-2")
	default:
		payload.Set("p", "-1")
	}
	return payload
}

// normalize cleans up whitespace in text
func normalize(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// normalizeURL removes fragments from URLs
func normalizeURL(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		parsed.Fragment = ""
		return parsed.String()
	}
	return u
}
