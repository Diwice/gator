package rss

import (
	"io"
	"fmt"
	"html"
	"context"
	"net/http"
	"encoding/xml"
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

func (o *RSSFeed) clean_feed() {
	o.Channel.Title = html.UnescapeString(o.Channel.Title)
	o.Channel.Description = html.UnescapeString(o.Channel.Description)

	for i, _ := range o.Channel.Item {
		o.Channel.Item[i].Title = html.UnescapeString(o.Channel.Item[i].Title)
		o.Channel.Item[i].Description = html.UnescapeString(o.Channel.Item[i].Description)
	}
}

func FetchFeed(ctx *context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(*ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	} else if resp.StatusCode > 299 {
		return &RSSFeed{}, fmt.Errorf("Response Status Code was not 200-: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, err
	}
	
	var res RSSFeed
	if err := xml.Unmarshal(body, &res); err != nil {
		return &RSSFeed{}, err
	}

	res.clean_feed()

	return &res, nil
}
