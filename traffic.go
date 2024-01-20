// Copyright (c) Alisdair MacLeod <copying@alisdairmacleod.co.uk>
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	roadTmpl = `<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<link rel="icon" href="data:,">
{{range .Entries}}<p>{{.}}</p>
{{end -}}`
	url = "https://m.highwaysengland.co.uk/feeds/rss/AllEvents.xml"
)

type entry struct {
	EntryTitle  string
	Link        string
	Description string
	Time        time.Time
}

type rss struct {
	Items []item `xml:"channel>item"`
}

type item struct {
	Title       string `xml:"title"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

func parseFeed(feed []byte) ([]entry, error) {
	var f rss
	if err := xml.Unmarshal(feed, &f); err != nil {
		return nil, fmt.Errorf("unmarshaling rss feed: %w", err)
	}
	var ret []entry
	for _, item := range f.Items {
		date, err := time.Parse(time.RFC1123, strings.TrimSpace(item.PubDate))
		if err != nil {
			return nil, fmt.Errorf("parsing pubDate: %w", err)
		}
		ret = append(ret, entry{
			EntryTitle:  item.Title,
			Link:        item.Link,
			Description: item.Description,
			Time:        date,
		})
	}
	return ret, nil
}

func main() {
	log.SetOutput(os.Stderr)
	tmpl := template.Must(template.New("road").Parse(roadTmpl))
	locationRegexp := regexp.MustCompile(`Location : The (.+?) `)
	statusRegexp := regexp.MustCompile(`Status : (.+?)\.`)
	roadworksRegexp := regexp.MustCompile(`Reason : .*?Roadworks.*?\n`)
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("error creating http request: %v\n", err)
	}
	req.Header.Add("User-Agent", "traffic (https://www.alisdairmacleod.co.uk/blog/projects/traffic.html)")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("error getting traffic news: %v\n", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("error closing request body: %v\n", err)
		}
	}()
	rawFeed, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("error reading http request: %v\n", err)
	}
	entries, err := parseFeed(rawFeed)
	if err != nil {
		log.Fatalf("error parsing traffic news: %v\n", err)
	}

	roads := make(map[string][]template.HTML)
	for _, entry := range entries {
		status := statusRegexp.FindStringSubmatch(entry.Description)
		if status == nil || status[1] != "Currently Active" || roadworksRegexp.MatchString(entry.Description) {
			continue
		}

		location := locationRegexp.FindStringSubmatch(entry.Description)
		if location == nil {
			continue
		}
		roads[location[1]] = append(roads[location[1]], template.HTML(strings.ReplaceAll(entry.Description, "\n", "<br>")))
	}

	if err := os.RemoveAll("traffic"); err != nil {
		log.Fatalf("error deleting directory %q: %v\n", "traffic", err)
	}

	if err := os.MkdirAll("traffic", os.ModePerm); err != nil {
		log.Fatalf("error creating directory %q: %v\n", "traffic", err)
	}

	for road, entries := range roads {
		outFile, err := os.Create(filepath.Join("traffic", fmt.Sprintf("%s.html", road)))
		if err != nil {
			log.Printf("error creating file for %q: %v\n", road, err)
			continue
		}
		if err := tmpl.Execute(outFile, struct {
			Title   string
			Entries []template.HTML
		}{
			Title:   road,
			Entries: entries,
		}); err != nil {
			log.Printf("error creating HTML for %q: %v\n", road, err)
			continue
		}
	}
}
