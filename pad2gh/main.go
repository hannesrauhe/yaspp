package main

/* pad2gh is a simple tool to get the first link from https://pad.ccc-p.org/Radio, extract the information from the markdown text and create a github PR with the information */

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// CiRaudio is the audio information for the podcast
type CiRaudio struct {
	Url      string // format: https://cdn.ccc-p.org/episodes/2021-01-01-episode.mp3 `yaml:"url"`
	MimeType string // format: audio/mpeg `yaml:"mimeType"`
}

// CiRChapter is the chapter information for the podcast
type CiRChapter struct {
	Start string // format: 00:00:00.000 `yaml:"start"`
	Title string `yaml:"title"`
}

// CiREntry is the podcast episode information
type CiREntry struct {
	UUID            string       `yaml:"uuid"`
	Title           string       `yaml:"title"`
	Subtitle        string       `yaml:"subtitle"`
	Summary         string       `yaml:"summary"`
	PublicationDate string       `yaml:"publicationDate"`
	Audio           CiRaudio     `yaml:"audio"`
	Chapters        []CiRChapter `yaml:"chapters"`
	LongSummaryMD   string       `yaml:"long_summary_md"`
}

func getPadContent(padURL string) io.ReadCloser {
	// append the HedgeDoc API path to get the raw pad content
	padURL = strings.TrimSuffix(padURL, "/")
	padURL = fmt.Sprintf("%s/download", padURL)
	resp, err := http.Get(padURL)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("pad url must be accessible")
	}
	return resp.Body
}

func getTitleFromFMA(fmaURL string) string {
	// append the HedgeDoc API path to get the raw pad content
	resp, err := http.Get(fmaURL)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("pad url must be accessible")
	}
	defer resp.Body.Close()
	// find the title tag in the html and return the content
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "<title>") {
			return strings.TrimSuffix(strings.TrimPrefix(line, "<title>"), "</title>")
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return fmaURL
}

func getFirstLink(padURL string) (string, error) {
	padContent := getPadContent(padURL)
	defer padContent.Close()

	// parse the content to find the first link
	scanner := bufio.NewScanner(padContent)
	for scanner.Scan() {
		line := scanner.Text()
		for _, linkCandidate := range strings.Split(line, "(") {
			if strings.HasPrefix(linkCandidate, "https://pad.ccc-p.org/") {
				if strings.HasSuffix(linkCandidate, ")") {
					link := strings.Split(linkCandidate, ")")
					if len(link) > 1 {
						return link[0], nil
					}
				}
			}
		}

	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func findFirstLink(line string) string {
	for _, linkCandidate := range strings.Split(line, " ") {
		if strings.HasPrefix(linkCandidate, "http") {
			return linkCandidate
		}
	}
	return ""
}

func getMarkdownContentBySection(padURL string) map[string][]string {
	padContent := getPadContent(padURL)
	defer padContent.Close()

	// parse the content to find the first link
	scanner := bufio.NewScanner(padContent)
	currentSection := "pre-section"
	currentSectionContent := []string{}
	contentBySection := make(map[string][]string)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			contentBySection[strings.ToLower(currentSection)] = currentSectionContent
			currentSectionContent = []string{}
			currentSection = strings.TrimPrefix(line, "## ")
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		}
		currentSectionContent = append(currentSectionContent, strings.Trim(line, " "))
	}
	contentBySection[currentSection] = currentSectionContent
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return contentBySection
}

func main() {
	logger := logrus.StandardLogger()
	outputFile := flag.String("o", "../content.yaml", "specify the yaml file to write to")
	padURLPtr := flag.String("l", "", "specify the link to the pad entry you want to parse")
	verbose := flag.Bool("v", false, "verbose output")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	flag.Parse()
	padURL := ""
	var err error
	if padURLPtr == nil || *padURLPtr == "" {
		padURL, err = getFirstLink("https://pad.ccc-p.org/Radio")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if !strings.HasPrefix(padURL, "https://pad.ccc-p.org/") {
			log.Fatal("pad url must start with https://pad.ccc-p.org/")
		}
		padURL = *padURLPtr
	}
	logger.Debugf("pad url: %s\n", padURL)
	contentBySection := getMarkdownContentBySection(padURL)

	if len(strings.Split(padURL, "_")) < 2 {
		log.Fatal("pad url must contain a date in the format YYYY-MM-DD_")
	}
	entryDate := strings.Split(padURL, "_")[1]
	if len(entryDate) < 10 {
		log.Fatal("pad url must contain a date in the format YYYY-MM-DD_")
	}
	fmt.Printf("entrydate=%s\n", entryDate)

	year := entryDate[0:4]
	month := entryDate[5:7]
	day := entryDate[8:10]
	longSummary, exists := contentBySection["shownotes"]
	if !exists {
		longSummary, exists = contentBySection["long summary"]
	}
	if !exists {
		log.Print("=== Found following Sections:")
		for k, c := range contentBySection {
			log.Printf("=== %s: %d\n", k, len(c))
		}
		log.Fatal("no shownotes Section in Pad")
	}
	shortSummary, exists := contentBySection["summary"]
	if !exists {
		log.Fatal("no Summary Section in Pad")
	}

	var entry CiREntry
	entry.UUID = fmt.Sprintf("nt-%s-%s-%s", year, month, day)
	entry.Title = fmt.Sprintf("CiR am %s.%s.%s", day, month, year)
	entry.Subtitle = "Der Chaostreff im Freien Radio Potsdam"
	entry.Summary = strings.Join(shortSummary, "\n")
	entry.PublicationDate = fmt.Sprintf("%s-%s-%sT00:00:00+02:00", year, month, day)
	entry.Audio.Url = fmt.Sprintf("$media_base_url/%s_%s_%s-chaos-im-radio.mp3", year, month, day)
	entry.Audio.MimeType = "audio/mp3"
	entry.LongSummaryMD = "**Shownotes:**\n" + strings.Join(longSummary, "\n")

	chapter, exists := contentBySection["chapters"]
	if exists {
		entry.Chapters = []CiRChapter{}
		for _, c := range chapter {
			chapter := strings.Split(c, " ")
			if len(chapter) < 2 {
				continue
			}
			entry.Chapters = append(entry.Chapters, CiRChapter{Start: chapter[0], Title: strings.Join(chapter[1:], " ")})
		}
	}

	mukke, exists := contentBySection["mukke"]
	if exists {
		for _, m := range mukke {
			link := findFirstLink(m)
			if link == "" {
				log.Printf("no link found in mukke entry: %s", m)
				continue
			}
			entry.LongSummaryMD = entry.LongSummaryMD + fmt.Sprintf("\n&#x1f3b6;&nbsp;[%s](%s)", getTitleFromFMA(link), link)
		}
	} else {
		log.Print("no mukke Section in Pad")
	}

	b, _ := yaml.Marshal(entry)

	if outputFile == nil || *outputFile == "" {
		fmt.Printf("%s", b)
		return
	}

	file, err := os.OpenFile(*outputFile, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		logger.Debugf("%s", b)
		logger.Fatalf("Error while opening %v: %v", *outputFile, err)
		return
	}
	_, err = file.Write(b)
	if err != nil {
		logger.Debugf("%s", b)
		logger.Fatalf("Error writing %v: %v", *outputFile, err)
		return
	}
	err = file.Close()
	if err != nil {
		logger.Debugf("%s", b)
		logger.Fatalf("Error closing %v: %v", *outputFile, err)
		return
	}
}
