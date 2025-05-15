package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var feedOut RSSFeed
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, geterr := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if geterr != nil {
		return nil, fmt.Errorf("error getting feedURL: %w", geterr)
	}

	req.Header.Set("User-Agent", "gator")

	res, reserr := client.Do(req)
	if reserr != nil {
		return nil, fmt.Errorf("error with http client response: %w", reserr)
	}
	defer res.Body.Close()

	xmlData, readerr := io.ReadAll(res.Body)
	if readerr != nil {
		return nil, fmt.Errorf("error reading http response: %w", readerr)
	}
	err := xml.Unmarshal(xmlData, &feedOut)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling xml: %w", err)
	}
	feedOut.Channel.Description = html.UnescapeString(feedOut.Channel.Description)
	feedOut.Channel.Title = html.UnescapeString(feedOut.Channel.Title)

	for i := range feedOut.Channel.Item {
		feedOut.Channel.Item[i].Title = html.UnescapeString(feedOut.Channel.Item[i].Title)
		feedOut.Channel.Item[i].Description = html.UnescapeString(feedOut.Channel.Item[i].Description)
	}

	return &feedOut, nil
}
