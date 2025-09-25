package aitools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Satr10/wa-userbot/internal/logger"
	"github.com/likexian/whois"
)

type Tools struct {
	log                *slog.Logger
	client             *http.Client
	redirectClient     *http.Client
	safeBrowsingApiKey string
}

func NewTools(safeBrowsingApiKey string) *Tools {
	if safeBrowsingApiKey == "" {
		panic("no safe safeBrowsingApiKey")
	}
	return &Tools{
		client: &http.Client{},
		redirectClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		log:                logger.Get(),
		safeBrowsingApiKey: safeBrowsingApiKey,
	}
}

func (t *Tools) ResolveShortUrl(shortURL string) (string, error) {
	t.log.Info("resolving short url for", "url", shortURL)
	resp, err := t.redirectClient.Get(shortURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// kalau statusnya redirect, ambil Location
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc, err := resp.Location()
		if err != nil {
			return "", err
		}
		return loc.String(), nil
	}

	// kalau bukan redirect, return URL itu sendiri
	return shortURL, nil
}

func (t *Tools) GetWhoisData(url string) (string, error) {
	t.log.Info("getting whois data for", "url", url)
	result, err := whois.Whois(url)
	if err != nil {
		return "", err
	}

	return result, nil
}

// Tambahkan fungsi mock lain yang Anda perlukan di sini
func (t *Tools) CheckGoogleSafeBrowsing(u string) (string, error) {
	t.log.Info("Checking GSB for", "url", u)

	apiKey := t.safeBrowsingApiKey
	if apiKey == "" {
		t.log.Error("missing GOOGLE_SAFE_BROWSING_API_KEY environment variable")
		return "", fmt.Errorf("missing GOOGLE_SAFE_BROWSING_API_KEY environment variable")
	}

	apiURL := "https://safebrowsing.googleapis.com/v4/threatMatches:find?key=" + apiKey

	body := map[string]interface{}{
		"client": map[string]string{
			"clientId":      "url-check-app",
			"clientVersion": "1.0.0",
		},
		"threatInfo": map[string]interface{}{
			"threatTypes":      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE", "POTENTIALLY_HARMFUL_APPLICATION"},
			"platformTypes":    []string{"ANY_PLATFORM"},
			"threatEntryTypes": []string{"URL"},
			"threatEntries":    []map[string]string{{"url": u}},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.log.Error("failed to marshal GSB request body", "error", err)
		return "", fmt.Errorf("failed to marshal GSB request body: %w", err)
	}
	t.log.Debug("sending request to GSB API", "body", string(jsonBody))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		t.log.Error("failed to create GSB request", "error", err)
		return "", fmt.Errorf("failed to create GSB request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		t.log.Error("failed to send request to GSB API", "error", err)
		return "", fmt.Errorf("failed to send request to GSB API: %w", err)
	}
	defer resp.Body.Close()

	t.log.Info("GSB API response", "status_code", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.log.Error("GSB API returned non-OK status", "status_code", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("GSB API returned status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.log.Error("failed to read GSB response body", "error", err)
		return "", fmt.Errorf("failed to read GSB response body: %w", err)
	}

	t.log.Debug("GSB API response body", "body", string(respBody))

	return string(respBody), nil
}

func (t *Tools) FetchPageContent(url string) (string, error) {
	t.log.Info("getting page Content for", "url", url)
	resp, err := t.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	return string(body), nil
}

type LexicalAnalysisResult struct {
	SuspicionScore int      `json:"suspicion_score"`
	Findings       []string `json:"findings"`
}

// LexicalAnalysis disempurnakan untuk menghasilkan temuan yang netral & berbasis data
func (t *Tools) LexicalAnalysis(rawURL string) (*LexicalAnalysisResult, error) {
	var findings []string
	var suspicionScore int

	suspiciousKeywords := []string{
		"login", "signin", "secure", "account", "update", "verify", "password",
		"banking", "confirm", "recovery", "admin",
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("URL tidak valid: %w", err)
	}
	hostname := parsedURL.Hostname()

	// 2. Cek apakah host adalah alamat IP
	if net.ParseIP(hostname) != nil {
		suspicionScore += 3
		findings = append(findings, "host_is_ip:true") // Laporan berbasis data
	}

	// 3. Cek panjang URL
	if len(rawURL) > 75 {
		suspicionScore++
		findings = append(findings, fmt.Sprintf("url_length:%d", len(rawURL))) // Laporan berbasis data
	}

	// 4. Cek keberadaan karakter '@'
	if strings.Contains(parsedURL.User.String(), "@") || strings.Contains(hostname, "@") {
		suspicionScore += 2
		findings = append(findings, "at_symbol_in_host:true") // Laporan berbasis data
	}

	// 5. Cek jumlah tanda hubung '-'
	dashCount := strings.Count(hostname, "-")
	if dashCount > 2 {
		suspicionScore++
		findings = append(findings, fmt.Sprintf("dash_count:%d", dashCount)) // Laporan berbasis data
	}

	// 6. Cek jumlah titik (subdomain)
	dotCount := strings.Count(hostname, ".")
	if dotCount > 3 {
		suspicionScore++
		findings = append(findings, fmt.Sprintf("subdomain_dot_count:%d", dotCount)) // Laporan berbasis data
	}

	// 7. Cek kata kunci mencurigakan
	for _, keyword := range suspiciousKeywords {
		if strings.Contains(strings.ToLower(rawURL), keyword) {
			suspicionScore++
			findings = append(findings, fmt.Sprintf("keyword_found:%s", keyword)) // Laporan berbasis data
			// Dihapus 'break' agar bisa melaporkan semua keyword yang ditemukan
		}
	}

	// 8. Cek jumlah garis miring '/'
	slashCount := strings.Count(parsedURL.Path, "/")
	if slashCount > 4 {
		suspicionScore++
		findings = append(findings, fmt.Sprintf("path_slash_count:%d", slashCount)) // Laporan berbasis data
	}

	return &LexicalAnalysisResult{
		SuspicionScore: suspicionScore,
		Findings:       findings,
	}, nil
}
