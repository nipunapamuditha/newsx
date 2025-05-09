package audio_generation_azure

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nipunapamuditha/NEXO/utils"
)

// implimentation of dickerdockgo instat search with LLM

func Fetach_substack_rss(usernames []string) ([]string, error) {
	// Slice to store all fetched articles
	var articles []string

	// Process each username
	for _, username := range usernames {
		// Create the RSS feed URL
		feedURL := fmt.Sprintf("https://%s.substack.com/feed", username)

		// Fetch the RSS feed
		resp, err := http.Get(feedURL)
		if err != nil {
			log.Printf("Error fetching RSS feed for %s: %v", username, err)
			continue // Skip to next username on error
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to fetch RSS feed for %s: status code %d", username, resp.StatusCode)
			continue
		}

		// Parse the XML feed (using encoding/xml)
		type Item struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
			Content string `xml:"encoded"`
		}

		type Channel struct {
			Title string `xml:"title"`
			Items []Item `xml:"item"`
		}

		type Feed struct {
			Channel Channel `xml:"channel"`
		}

		var feed Feed
		decoder := xml.NewDecoder(resp.Body)
		if err := decoder.Decode(&feed); err != nil {
			log.Printf("Error parsing RSS feed for %s: %v", username, err)
			continue
		}

		// Process each article
		for _, item := range feed.Channel.Items {
			// Parse publication date
			pubDate, err := time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				// Try alternative parsing if standard format fails
				pubDate, err = parseDate(item.PubDate)
				if err != nil {
					log.Printf("Error parsing date %s: %v", item.PubDate, err)
					continue
				}
			}

			// Check if article is from the last 24 hours
			if time.Since(pubDate) <= 24*time.Hour {
				// Clean the content to remove images and unnecessary elements
				cleanContent := cleanArticleContent(item.Content)

				// Create article string with clean content
				articleText := fmt.Sprintf("Title: %s\nLink: %s\nPublished: %s\n\nContent:\n%s",
					item.Title, item.Link, item.PubDate, cleanContent)

				// Add to articles slice
				articles = append(articles, articleText)
			}
		}
	}

	// Check if we found any articles
	if len(articles) == 0 {
		return nil, fmt.Errorf("no recent articles found for the provided usernames")
	}

	return articles, nil
}

// cleanArticleContent removes images, divs, scripts and other unwanted elements from HTML
func cleanArticleContent(content string) string {
	// Remove image containers, figures, and other similar elements
	imagePatterns := []string{
		`<div class="captioned-image-container">.*?</div>`,
		`<figure>.*?</figure>`,
		`<img.*?>`,
		`<picture>.*?</picture>`,
		`<div class="image.*?">.*?</div>`,
	}

	for _, pattern := range imagePatterns {
		re := regexp.MustCompile(`(?s)` + pattern) // (?s) makes . match newlines
		content = re.ReplaceAllString(content, "")
	}

	// Remove other unwanted elements
	unwantedElements := []string{
		`<div class="pencraft.*?">.*?</div>`,
		`<svg.*?</svg>`,
		`<iframe.*?</iframe>`,
		`<style.*?</style>`,
		`<script.*?</script>`,
	}

	for _, pattern := range unwantedElements {
		re := regexp.MustCompile(`(?s)` + pattern)
		content = re.ReplaceAllString(content, "")
	}

	// Remove remaining HTML tags but preserve paragraphs and headings
	content = convertToPlainText(content)

	// Remove excessive whitespace and blank lines
	re := regexp.MustCompile(`\n\s*\n`)
	content = re.ReplaceAllString(content, "\n\n")
	content = strings.TrimSpace(content)

	return content
}

// convertToPlainText preserves paragraph structure while removing HTML tags
func convertToPlainText(html string) string {
	// Replace paragraph and heading tags with newlines
	re := regexp.MustCompile(`<(?:p|h[1-6])[^>]*>`)
	html = re.ReplaceAllString(html, "")

	re = regexp.MustCompile(`</(?:p|h[1-6])>`)
	html = re.ReplaceAllString(html, "\n\n")

	// Replace <br> tags with newlines
	re = regexp.MustCompile(`<br\s*/?>`)
	html = re.ReplaceAllString(html, "\n")

	// Remove remaining HTML tags
	re = regexp.MustCompile(`<[^>]*>`)
	html = re.ReplaceAllString(html, "")

	// Decode HTML entities
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#8220;", "\u201C")
	html = strings.ReplaceAll(html, "&#8221;", "\u201D")
	html = strings.ReplaceAll(html, "&#8217;", "\u2019")

	return html
}

// Helper function to try multiple date formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 GMT",
		"Mon, 2 Jan 2006 15:04:05 GMT",
	}

	var parseErr error
	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
		parseErr = err
	}
	return time.Time{}, parseErr
}

func Generate_script_azure(articles []string) (string, error) {
	if len(articles) == 0 {
		log.Printf("no articles provided")
		return "", fmt.Errorf("no articles provided")
	}

	var combinedText strings.Builder

	for i, article := range articles {
		combinedText.WriteString(article)

		// Add separator between articles (except after the last one)
		if i < len(articles)-1 {
			combinedText.WriteString("\n\n--------------------\n\n")

		}
	}

	// sending to azure
	url := "https://ai-nipunak995299ai116832957294.services.ai.azure.com/models/chat/completions?api-version=2024-05-01-preview"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type RequestPayload struct {
		Messages  []Message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
		Model     string    `json:"model"`
	}

	// Create the request payload
	payload := RequestPayload{
		Messages: []Message{
			{
				Role:    "system",
				Content: "Act as a news scriptwriter channeling the energetic, conversational tone of a YouTube host like Philip DeFranco. Write a TTS-ready voiceover script that synthesizes key information from multiple articles into one cohesive, engaging narrative. Important: Do not include any formatting markers, section headers, asterisks, bullet points, or non-verbal cues—every character in the script should be intended to be spoken aloud. Emulate Philip DeFranco's style in tone and pacing: use a lively, casual voice with modern phrasing. The script should sound natural and unscripted, as if a real person is speaking. Incorporate snappy transitions between topics (e.g., \"Now, in other news…\" or \"So let’s talk about…\") to keep the flow dynamic. Add light commentary and humor to give color to the facts without distorting them (for example, a brief witty aside or an emotional remark like, \"Can you believe that?\"). Use rhetorical questions and asides to hook listeners and make the delivery feel interactive (e.g., \"Crazy, right?\" or \"What would you do in that situation?\"). Maintain conversational pacing with natural pauses where appropriate. Vary sentence length for rhythm—mix short, punchy sentences with occasional longer ones for explanation. Keep the script highly speakable and clear: short sentences and straightforward words ensure the TTS system pronounces everything cleanly. Break up long paragraphs into smaller chunks or lines so the voiceover doesn’t sound monotonous or overwhelming. Each thought or idea can start on a new line to help pacing. Avoid robotic or overly formal language. The tone should be friendly and relatable, as if talking with the audience, not at them. Contractions and relaxed wording (gonna, kinda, you guys) are welcome if they enhance the natural feel. Ensure smooth audio transitions between stories by using transitional phrases and a consistent upbeat energy. Goal – The final script should read like something Philip DeFranco would say on his show—informative and factual, but also energetic, witty, and personable. It must sound like it’s delivered by a real human host, with enthusiasm and clarity, while remaining perfectly suited for text-to-speech narration. Every part of the script should flow conversationally, keeping listeners hooked from start to finish.",
			},
			{
				Role:    "user",
				Content: combinedText.String(),
			},
		},
		MaxTokens: 3000,
		Model:     "DeepSeek-R1",
	}

	// Check if the payload is empty
	//log.Printf("combinxed text status all --- : %+v", combinedText.String())

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON payload: %v", err)
		return "", fmt.Errorf("error marshaling JSON payload: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// set api key in the env files

	Azure_LLM_key, err := utils.GetEnvVariable("AZURE_LLM_KEY")
	if err != nil {
		log.Printf("Error getting Azure LLM key: %v", err)
		return "", fmt.Errorf("error getting Azure LLM key: %v", err)
	}

	req.Header.Set("api-key", Azure_LLM_key)

	// Send the request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to Azure: %v", err)
		return "", fmt.Errorf("error sending request to Azure: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Azure API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("Azure API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	type Choice struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	type ResponseBody struct {
		Choices []Choice `json:"choices"`
	}

	var responseBody ResponseBody
	// Note: Don't log resp.Body directly as it's not a string
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		log.Printf("Error decoding response: %v", err)
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Check if we have choices and content in the response
	if len(responseBody.Choices) == 0 {
		log.Printf("No choices in response from Azure AI")
		return "", fmt.Errorf("no content in response from Azure AI")
	}

	generatedScript := responseBody.Choices[0].Message.Content

	log.Printf("Generated script: %s", generatedScript)
	if generatedScript == "" {
		log.Printf("Empty script generated from Azure AI")
		return "", fmt.Errorf("empty script generated")
	}

	// modify generated script to remove think tag
	thinkPattern := regexp.MustCompile(`(?s)<think>.*?</think>`)
	generatedScript = thinkPattern.ReplaceAllString(generatedScript, "")

	// Trim any whitespace that might be left after removing the tags
	generatedScript = strings.TrimSpace(generatedScript)

	return generatedScript, nil
}

// ...existing code...

func Generate_audio_file_azure(text string, name string) (bool, error) {
	// Azure TTS settings
	speechKey, err := utils.GetEnvVariable("AZURE_SPEECH_KEY")
	if err != nil {
		// Fallback to the hardcoded key only if env variable isn't available
		return false, fmt.Errorf("AZURE_SPEECH_KEY environment variable not set: %v", err)
	}
	serviceRegion := "eastus2"
	voiceName := "en-US-Andrew2:DragonHDLatestNeural"

	// Format today's date for filename
	todayDate := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s.wav", todayDate)

	// API endpoint
	endpoint := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", serviceRegion)

	// Prepare SSML payload
	ssml := fmt.Sprintf(`<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>
        <voice name='%s'>%s</voice>
    </speak>`, voiceName, text)

	// Create HTTP request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(ssml))
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}

	// Set required headers
	req.Header.Set("Ocp-Apim-Subscription-Key", speechKey)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "riff-24khz-16bit-mono-pcm") // WAV format
	req.Header.Set("User-Agent", "GoSpeechClient")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)

		if strings.Contains(body, "error") {
			return false, fmt.Errorf("speech synthesis canceled: %s", body)
		}
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the entire response body
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response: %v", err)
	}

	// Configure MinIO client
	minioEndpoint := "minioapi.newsloop.xyz"
	bucketName := "newsx"

	// Get MinIO credentials from environment variables
	accessKeyID, err := utils.GetEnvVariable("MINIO_ACCESS_KEY")
	if err != nil {
		return false, fmt.Errorf("MINIO_ACCESS_KEY environment variable not set: %v", err)
	}

	secretAccessKey, err := utils.GetEnvVariable("MINIO_SECRET_KEY")
	if err != nil {
		return false, fmt.Errorf("MINIO_SECRET_KEY environment variable not set: %v", err)
	}

	// Create MinIO client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return false, fmt.Errorf("error creating MinIO client: %v", err)
	}

	// Define object path (folder/filename)
	objectName := fmt.Sprintf("%s/%s", name, filename)

	// Upload file to MinIO
	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, bytes.NewReader(audioData), int64(len(audioData)), minio.PutObjectOptions{
		ContentType: "audio/wav",
	})
	if err != nil {
		return false, fmt.Errorf("error uploading to MinIO: %v", err)
	}

	// Verify the object exists before returning success
	// Add a short delay to allow for any potential consistency issues
	time.Sleep(500 * time.Millisecond)
	log.Printf("Verifying uploaded file in MinIO...")

	// Check if the object exists in MinIO
	xcx, err := minioClient.StatObject(context.Background(), bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.Printf("Error verifying uploaded file: %v", err)
		return false, fmt.Errorf("file upload verification failed: %v", err)
	}

	log.Printf("File uploaded successfully: %s, Size: %d bytes", objectName, xcx.Size)

	fmt.Printf("Speech synthesized and saved to MinIO: '%s/%s'\n", bucketName, objectName)
	return true, nil
}
