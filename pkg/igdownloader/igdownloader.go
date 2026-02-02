package igdownloader

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type DownloadResult struct {
	VideoFile string
	VideoID   string
	Success   bool
	Error     error
}

type GraphQLResponse struct {
	Data struct {
		XdtShortcodeMedia struct {
			IsVideo  bool   `json:"is_video"`
			VideoURL string `json:"video_url"`
			Dimensions struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"dimensions"`
		} `json:"xdt_shortcode_media"`
	} `json:"data"`
}

// DownloadInstagramVideo downloads a video from Instagram URL using GraphQL API
func DownloadInstagramVideo(instaURL string) *DownloadResult {
	log.Printf("Starting Instagram download for: %s", instaURL)

	shortcode, err := extractShortcode(instaURL)
	if err != nil {
		return &DownloadResult{Success: false, Error: err}
	}

	result := &DownloadResult{
		VideoID: shortcode,
		Success: false,
	}

	videoURL, err := fetchVideoInfoFromGraphQL(shortcode)
	if err != nil {
		log.Printf("Error fetching video info: %v", err)
		result.Error = err
		return result
	}

	log.Printf("Got video URL: %s", videoURL)

	outputFile := fmt.Sprintf("%s.mp4", shortcode)
	if err := downloadFile(videoURL, outputFile); err != nil {
		log.Printf("Error downloading video: %v", err)
		result.Error = err
		return result
	}

	result.VideoFile = outputFile
	result.Success = true
	log.Printf("Successfully downloaded Instagram video: %s", outputFile)

	return result
}

// fetchVideoInfoFromGraphQL makes a GraphQL request to get video URL
func fetchVideoInfoFromGraphQL(shortcode string) (string, error) {
	formData := buildGraphQLRequestBody(shortcode)
	log.Printf("Request body: %s", formData)

	req, err := http.NewRequest("POST", "https://www.instagram.com/api/graphql", strings.NewReader(formData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 11; SAMSUNG SM-G973U) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/14.2 Chrome/87.0.4280.141 Mobile Safari/537.36")
	req.Header.Set("X-FB-Friendly-Name", "PolarisPostActionLoadPostQueryQuery")
	req.Header.Set("X-CSRFToken", "RVDUooU5MYsBbS1CNN3CzVAuEP8oHB52")
	req.Header.Set("X-IG-App-ID", "1217981644879628")
	req.Header.Set("X-FB-LSD", "AVqbxe3J_YA")
	req.Header.Set("X-ASBD-ID", "129477")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var graphqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !graphqlResp.Data.XdtShortcodeMedia.IsVideo {
		return "", fmt.Errorf("post does not contain a video")
	}

	videoURL := graphqlResp.Data.XdtShortcodeMedia.VideoURL
	if videoURL == "" {
		return "", fmt.Errorf("video URL is empty")
	}

	return videoURL, nil
}

// buildGraphQLRequestBody creates the form-encoded request body for GraphQL API
func buildGraphQLRequestBody(shortcode string) string {
	variables := fmt.Sprintf(`{
		"shortcode": "%s",
		"fetch_comment_count": "null",
		"fetch_related_profile_media_count": "null",
		"parent_comment_count": "null",
		"child_comment_count": "null",
		"fetch_like_count": "null",
		"fetch_tagged_user_count": "null",
		"fetch_preview_comment_count": "null",
		"has_threaded_comments": "false",
		"hoisted_comment_id": "null",
		"hoisted_reply_id": "null"
	}`, shortcode)

	formData := url.Values{}
	formData.Set("av", "0")
	formData.Set("__d", "www")
	formData.Set("__user", "0")
	formData.Set("__a", "1")
	formData.Set("__req", "3")
	formData.Set("__hs", "19624.HYP:instagram_web_pkg.2.1..0.0")
	formData.Set("dpr", "3")
	formData.Set("__ccg", "UNKNOWN")
	formData.Set("__rev", "1008824440")
	formData.Set("__s", "xf44ne:zhh75g:xr51e7")
	formData.Set("__hsi", "7282217488877343271")
	formData.Set("__dyn", "7xeUmwlEnwn8K2WnFw9-2i5U4e0yoW3q32360CEbo1nEhw2nVE4W0om78b87C0yE5ufz81s8hwGwQwoEcE7O2l0Fwqo31w9a9x-0z8-U2zxe2GewGwso88cobEaU2eUlwhEe87q7-0iK2S3qazo7u1xwIw8O321LwTwKG1pg661pwr86C1mwraCg")
	formData.Set("__csr", "gZ3yFmJkillQvV6ybimnG8AmhqujGbLADgjyEOWz49z9XDlAXBJpC7Wy-vQTSvUGWGh5u8KibG44dBiigrgjDxGjU0150Q0848azk48N09C02IR0go4SaR70r8owyg9pU0V23hwiA0LQczA48S0f-x-27o05NG0fkw")
	formData.Set("__comet_req", "7")
	formData.Set("lsd", "AVqbxe3J_YA")
	formData.Set("jazoest", "2957")
	formData.Set("__spin_r", "1008824440")
	formData.Set("__spin_b", "trunk")
	formData.Set("__spin_t", "1695523385")
	formData.Set("fb_api_caller_class", "RelayModern")
	formData.Set("fb_api_req_friendly_name", "PolarisPostActionLoadPostQueryQuery")
	formData.Set("variables", variables)
	formData.Set("server_timestamps", "true")
	formData.Set("doc_id", "10015901848480474")

	return formData.Encode()
}

// extractShortcode extracts shortcode from Instagram URL
func extractShortcode(instaURL string) (string, error) {
	u, err := url.Parse(instaURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if !strings.Contains(u.Host, "instagram.com") {
		return "", fmt.Errorf("invalid Instagram URL: domain must be instagram.com")
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 && (parts[0] == "p" || parts[0] == "reel" || parts[0] == "reels") {
		return parts[1], nil
	}

	if len(parts) >= 1 {
		return parts[0], nil
	}

	return "", fmt.Errorf("could not extract shortcode from URL: %s", instaURL)
}

// downloadFile downloads a file from URL and saves it locally
func downloadFile(url, filepath string) error {
	log.Printf("Downloading: %s", filepath)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ExtractInstagramURL extracts Instagram URLs from text content
func ExtractInstagramURL(text string) string {
	words := strings.Fields(text)
	for _, word := range words {
		if strings.Contains(word, "instagram.com") {
			url := strings.TrimRight(word, ".,!?;")
			return url
		}
	}
	return ""
}
