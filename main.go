package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type MediaJSONMinimal struct {
	MediaResource struct {
		Alt struct {
			VideoURL string `json:"videoURL"`
		} `json:"alt"`
		PreviewImage string `json:"previewImage"`
	} `json:"mediaResource"`
	TrackerData struct {
		TrackerClipTitle string `json:"trackerClipTitle"`
	} `json:"trackerData"`
}

func main() {
	urls := parseUrls()

	i := 1
	fmt.Println("\n\n=== starting download ===")
	for _, url := range urls {
		fmt.Println(fmt.Sprintf("downloading from: %s (%d/%d)", url, i, len(urls)))
		downloadMausEpisode(url)
		i++
	}
	fmt.Println("=== finished download ===")
}

func parseUrls() []string {
	urlsFlag := flag.String("urls", "", "Comma-separated list of URLs")
	flag.Parse()

	if *urlsFlag == "" {
		fmt.Println("No URLs provided.")
		os.Exit(1)
	}
	return strings.Split(*urlsFlag, ",")
}

func downloadMausEpisode(url string) {
	html := getHtml(url)
	// extract json object that contains metadata for file download and download url

	r := regexp.MustCompile(`https://[^"' ]*jsonp`)
	metadataUrl := r.FindString(html)
	jsonStr := strings.ReplaceAll(strings.ReplaceAll(getHtml(metadataUrl), "$mediaObject.jsonpHelper.storeAndPlay(", ""), ");", "")
	var media MediaJSONMinimal
	err := json.Unmarshal([]byte(jsonStr), &media)
	if err != nil {
		panic(err)
	}

	title := media.TrackerData.TrackerClipTitle
	// previewImage := media.MediaResource.PreviewImage
	videoUrl := "https:" + media.MediaResource.Alt.VideoURL
	downloadFile(createFilename(title), videoUrl)
}

func createFilename(title string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(title), ".", "_"), " ", "_") + ".mp4"
}

func getHtml(link string) string {
	res, err := http.Get(link)
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	return string(content)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
