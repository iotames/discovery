package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

const baseURL = "https://www.dsite.com/"
const baseDomain = "www.dsite.com"

var visited = make(map[string]bool)
var visitedMutex = sync.Mutex{}
var downloadDir = "downloads"

func main() {
	fmt.Println("Starting site download...")

	// Create downloads directory
	err := os.MkdirAll(downloadDir, 0755)
	if err != nil {
		log.Fatal("Error creating downloads directory:", err)
	}

	// Create resource directories
	dirs := []string{"css", "js", "images", "fonts", "media", "fetch_xhr"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(downloadDir, dir), 0755)
		if err != nil {
			log.Fatal("Error creating resource directory:", err)
		}
	}

	// Start crawling from the base URL
	err = downloadPage(baseURL)
	if err != nil {
		log.Fatal("Error downloading page:", err)
	}

	fmt.Println("Site download completed.")
}

func downloadPage(pageURL string) error {
	// Check if already visited
	visitedMutex.Lock()
	if visited[pageURL] {
		visitedMutex.Unlock()
		return nil
	}
	visited[pageURL] = true
	visitedMutex.Unlock()

	// Parse URL
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("error parsing URL %s: %v", pageURL, err)
	}

	// Only crawl HTML pages from the base domain
	isBaseDomain := parsedURL.Host == baseDomain
	if !isBaseDomain {
		return fmt.Errorf("error URL %s is not in baseDomain %s", pageURL, baseDomain)
	}

	// Check if HTML file already exists locally
	if isBaseDomain && resourceExists(pageURL, "html") {
		fmt.Printf("HTML file already exists locally: %s\n", pageURL)
		localFilePath := getFilePath(pageURL, "html")

		file, err := os.Open(localFilePath)
		if err != nil {
			return fmt.Errorf("error opening local file %s: %v", localFilePath, err)
		}
		defer file.Close()

		// Process HTML document from local file
		return processHTMLDocument(file, pageURL)
	}

	// For non-HTML resources, check if they exist locally
	if !isBaseDomain {
		// Try to determine resource type from URL
		resourceType := getResourceType(pageURL)
		if resourceType != "html" && resourceExists(pageURL, resourceType) {
			fmt.Printf("Resource already exists locally: %s\n", pageURL)
			return nil
		}
	}

	fmt.Printf("Downloading: %s (base domain: %t)\n", pageURL, isBaseDomain)

	// Download content
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request for %s: %v", pageURL, err)
	}

	// Set User-Agent header to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading %s: %v", pageURL, err)
	}
	defer resp.Body.Close()

	// Process based on content type
	contentType := resp.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/html") && isBaseDomain {
		// Process HTML document
		return processHTMLDocument(resp.Body, pageURL)
	} else if strings.Contains(contentType, "text/html") {
		// Non-base domain HTML - treat as resource
		return saveResource(resp.Body, pageURL, "html")
	} else if strings.Contains(contentType, "text/css") {
		return saveResource(resp.Body, pageURL, "css")
	} else if strings.Contains(contentType, "application/javascript") || strings.Contains(contentType, "text/javascript") {
		return saveResource(resp.Body, pageURL, "js")
	} else if strings.HasPrefix(contentType, "image/") {
		return saveResource(resp.Body, pageURL, "images")
	} else if strings.Contains(contentType, "font/") || strings.Contains(contentType, "application/font") {
		return saveResource(resp.Body, pageURL, "fonts")
	} else if strings.Contains(contentType, "audio/") || strings.Contains(contentType, "video/") {
		return saveResource(resp.Body, pageURL, "media")
	} else if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/json") {
		return saveResource(resp.Body, pageURL, "fetch_xhr")
	} else {
		// Handle as generic resource
		return saveResource(resp.Body, pageURL, "other")
	}
}

func processHTMLDocument(body io.Reader, pageURL string) error {
	// Read the body into a buffer so we can scan it multiple times
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	content := buf.String()

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("error parsing HTML: %v", err)
	}

	// Extract links
	links := extractLinks(doc, pageURL)
	fmt.Printf("Found %d links\n", len(links))

	// Save the HTML file
	filePath := getFilePath(pageURL, "html")

	// Create directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory for %s: %v", filePath, err)
	}

	// Rewrite URLs in the content before saving
	rewrittenContent := rewriteURLs(content, pageURL)

	// Save the file if it doesn't already exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file %s: %v", filePath, err)
		}
		defer file.Close()

		_, err = file.WriteString(rewrittenContent)
		if err != nil {
			return fmt.Errorf("error writing to file %s: %v", filePath, err)
		}
		fmt.Printf("Saved HTML document: %s\n", filePath)
	} else {
		fmt.Printf("HTML document already exists: %s\n", filePath)
	}

	// Process extracted links
	for _, link := range links {
		if link != "" && !strings.HasPrefix(link, "#") {
			// Resolve relative URLs
			fullURL, err := url.Parse(pageURL)
			if err != nil {
				continue
			}

			resolvedURL, err := fullURL.Parse(link)
			if err != nil {
				continue
			}

			// Download the linked resource/page
			err = downloadPage(resolvedURL.String())
			if err != nil {
				fmt.Printf("Warning: failed to download %s: %v\n", resolvedURL.String(), err)
			}
		}
	}

	return nil
}

func extractLinks(n *html.Node, base string) []string {
	var links []string

	// Recursive function to traverse the HTML tree
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var attrName string
			switch n.Data {
			case "a":
				attrName = "href"
			case "link":
				attrName = "href"
			case "script":
				attrName = "src"
			case "img":
				attrName = "src"
			case "source":
				attrName = "src"
			case "iframe":
				attrName = "src"
			case "audio":
				attrName = "src"
			case "video":
				attrName = "src"
			case "track":
				attrName = "src"
			}

			// Extract the attribute value
			for _, attr := range n.Attr {
				if attr.Key == attrName && attr.Val != "" {
					links = append(links, attr.Val)
					break
				}

				// Also check for srcset attribute (used in img and source elements)
				if (n.Data == "img" || n.Data == "source") && attr.Key == "srcset" && attr.Val != "" {
					// Simple parsing of srcset - just get the first URL
					srcsetParts := strings.Split(attr.Val, ",")
					if len(srcsetParts) > 0 {
						firstURL := strings.TrimSpace(strings.Split(srcsetParts[0], " ")[0])
						if firstURL != "" {
							links = append(links, firstURL)
						}
					}
				}
			}
		}

		// Recursively process child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return links
}

func saveResource(body io.Reader, resourceURL string, resourceType string) error {
	// Get file path for this resource
	filePath := getFilePath(resourceURL, resourceType)

	// Check if file already exists locally
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("Resource already exists locally: %s\n", filePath)
		return nil
	}

	// Create directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory for %s: %v", filePath, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", filePath, err)
	}
	defer file.Close()

	// Copy the content
	_, err = io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %v", filePath, err)
	}

	fmt.Printf("Saved resource: %s\n", filePath)
	return nil
}

func getFilePath(rawURL, contentType string) string {
	return urlToLocalPath(rawURL, contentType)
}

func rewriteURLs(content, base string) string {
	// Parse the base URL for reference
	baseURL, err := url.Parse(base)
	if err != nil {
		return content
	}

	currentPageDir := getCurrentPageDir(base)

	// Replace href attributes
	hrefRe := regexp.MustCompile(`href=["']([^"']*)["']`)
	content = hrefRe.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the URL using a new regex instance inside the closure
		urlMatch := regexp.MustCompile(`href=["']([^"']*)["']`).FindStringSubmatch(match)
		if len(urlMatch) < 2 {
			return match
		}

		rawURL := urlMatch[1]
		if rawURL == "" || strings.HasPrefix(rawURL, "#") {
			return match
		}

		// Resolve the URL
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return match
		}

		resolved := baseURL.ResolveReference(parsedURL)
		localPath := urlToLocalPath(resolved.String(), "text/html")

		// Convert to relative path
		relPath, err := filepath.Rel(currentPageDir, localPath)
		if err != nil {
			return match
		}

		// Ensure path separator is forward slash for web
		relPath = filepath.ToSlash(relPath)
		return fmt.Sprintf(`href="%s"`, relPath)
	})

	// Replace src attributes
	srcRe := regexp.MustCompile(`src=["']([^"']*)["']`)
	content = srcRe.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the URL using a new regex instance inside the closure
		urlMatch := regexp.MustCompile(`src=["']([^"']*)["']`).FindStringSubmatch(match)
		if len(urlMatch) < 2 {
			return match
		}

		rawURL := urlMatch[1]
		if rawURL == "" {
			return match
		}

		// Resolve the URL
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return match
		}

		resolved := baseURL.ResolveReference(parsedURL)
		resourceType := getResourceType(resolved.String())
		localPath := urlToLocalPath(resolved.String(), resourceType)

		// Convert to relative path
		relPath, err := filepath.Rel(currentPageDir, localPath)
		if err != nil {
			return match
		}

		// Ensure path separator is forward slash for web
		relPath = filepath.ToSlash(relPath)
		return fmt.Sprintf(`src="%s"`, relPath)
	})

	// Handle CDN image paths in URLs
	content = strings.ReplaceAll(content, "/cdn-cgi/image/", "/files/")

	return content
}

// Helper function to determine resource type based on URL
func getResourceType(rawURL string) string {
	// Try to determine resource type based on file extension
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "other"
	}

	path := parsedURL.Path
	ext := filepath.Ext(path)

	switch strings.ToLower(ext) {
	case ".css":
		return "css"
	case ".js":
		return "js"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "images"
	case ".woff", ".woff2", ".ttf", ".eot":
		return "fonts"
	case ".mp4", ".avi", ".mov", ".wmv":
		return "media"
	case ".json":
		return "fetch_xhr"
	case ".html":
		return "html"
	default:
		// For unknown extensions, return "other" to avoid network requests during URL rewriting
		return "other"
	}
}

// Convert URL to local file path
func urlToLocalPath(rawURL, contentType string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Handle CDN image path transformation
	path := parsedURL.Path
	if strings.Contains(path, "/cdn-cgi/image/") {
		// Transform CDN image paths
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "cdn-cgi" && i+2 < len(parts) && parts[i+1] == "image" {
				// Reconstruct path starting from the part after "image"
				if i+3 < len(parts) {
					path = "/files/" + strings.Join(parts[i+3:], "/")
				} else {
					path = "/files/"
				}
				break
			}
		}
	}

	// Determine file extension if not present
	ext := filepath.Ext(path)
	if ext == "" {
		switch contentType {
		case "css":
			path += ".css"
		case "js":
			path += ".js"
		case "html":
			if strings.HasSuffix(path, "/") || path == "" {
				path += "index.html"
			} else {
				path += ".html"
			}
		case "fetch_xhr":
			path += ".json"
		}
	}

	// For HTML documents, save in the domain subdirectory
	if contentType == "html" {
		hostDir := strings.ReplaceAll(parsedURL.Host, ".", "_")
		// Clean the path to remove any .. elements
		cleanPath := filepath.Clean(path)
		// Remove leading slash for filepath.Join
		cleanPath = strings.TrimPrefix(cleanPath, "/")
		return filepath.Join(downloadDir, hostDir, cleanPath)
	}

	// For resources, organize by type and domain
	hostDir := strings.ReplaceAll(parsedURL.Host, ".", "_")
	// Clean the path to remove any .. elements
	cleanPath := filepath.Clean(path)
	// Remove leading slash for filepath.Join
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	return filepath.Join(downloadDir, contentType, hostDir, cleanPath)
}

// Get directory of current page for relative path calculation
func getCurrentPageDir(pageURL string) string {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		hostDir := strings.ReplaceAll(baseDomain, ".", "_")
		return filepath.Join(downloadDir, hostDir)
	}

	path := parsedURL.Path
	if strings.HasSuffix(path, ".html") {
		path = filepath.Dir(path)
	} else if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Handle root path
	if path == "" || path == "/" {
		path = "/"
	}

	hostDir := strings.ReplaceAll(parsedURL.Host, ".", "_")

	// Special handling for root path
	if path == "/" {
		return filepath.Join(downloadDir, hostDir)
	}

	// Remove leading slash for filepath.Join
	path = strings.TrimPrefix(path, "/")
	return filepath.Join(downloadDir, hostDir, path)
}

// Check if a resource already exists locally
func resourceExists(resourceURL string, contentType string) bool {
	localPath := getFilePath(resourceURL, contentType)
	_, err := os.Stat(localPath)
	return err == nil
}
