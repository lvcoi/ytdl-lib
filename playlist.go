package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	playlistIDRegex    = regexp.MustCompile("^[A-Za-z0-9_-]{13,42}$")
	playlistInURLRegex = regexp.MustCompile("[&?]list=([A-Za-z0-9_-]{13,42})(&.*)?$")
)

type Playlist struct {
	ID          string
	Title       string
	Description string
	Author      string
	Videos      []*PlaylistEntry
}

type PlaylistEntry struct {
	ID         string
	Title      string
	Author     string
	Duration   time.Duration
	Thumbnails Thumbnails
}

func extractPlaylistID(url string) (string, error) {
	if playlistIDRegex.Match([]byte(url)) {
		return url, nil
	}

	matches := playlistInURLRegex.FindStringSubmatch(url)

	if matches != nil {
		return matches[1], nil
	}

	return "", ErrInvalidPlaylist
}

// structs for playlist extraction

// Title: metadata.playlistMetadataRenderer.title | sidebar.playlistSidebarRenderer.items[0].playlistSidebarPrimaryInfoRenderer.title.runs[0].text
// Description: metadata.playlistMetadataRenderer.description
// Author: sidebar.playlistSidebarRenderer.items[1].playlistSidebarSecondaryInfoRenderer.videoOwner.videoOwnerRenderer.title.runs[0].text

// Videos: contents.twoColumnBrowseResultsRenderer.tabs[0].tabRenderer.content.sectionListRenderer.contents[0].itemSectionRenderer.contents[0].playlistVideoListRenderer.contents
// ID: .videoId
// Title: title.runs[0].text
// Author: .shortBylineText.runs[0].text
// Duration: .lengthSeconds
// Thumbnails .thumbnails

func (p *Playlist) parsePlaylistInfo(ctx context.Context, client *Client, body []byte) error {
	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}

	// Check for error alerts
	alertRenderer := jsonGet(jsonGetIndex(jsonGet(raw, "alerts"), 0), "alertRenderer")
	if jsonHasContent(alertRenderer) && jsonString(jsonGet(alertRenderer, "type")) == "ERROR" {
		message := jsonString(jsonGet(jsonGetIndex(jsonGet(alertRenderer, "text", "runs"), 0), "text"))
		return ErrPlaylistStatus{Reason: message}
	}

	// Metadata can be located in multiple places depending on client type
	var metadata json.RawMessage
	if node := jsonGet(raw, "metadata"); jsonHasContent(node) {
		metadata = node
	} else if node := jsonGet(raw, "header"); jsonHasContent(node) {
		metadata = node
	} else {
		return fmt.Errorf("no playlist header / metadata found")
	}

	metadata = jsonGet(metadata, "playlistHeaderRenderer")

	p.Title = jsonGetText(metadata, "title")
	p.Description = jsonGetText(metadata, "description", "descriptionText")
	p.Author = jsonString(jsonGet(
		jsonGetIndex(jsonGet(
			jsonGetIndex(jsonGet(raw, "sidebar", "playlistSidebarRenderer", "items"), 1),
			"playlistSidebarSecondaryInfoRenderer", "videoOwner", "videoOwnerRenderer", "title", "runs",
		), 0),
		"text",
	))

	if len(p.Author) == 0 {
		p.Author = jsonGetText(metadata, "owner", "ownerText")
	}

	contents := jsonGet(raw, "contents")
	if !jsonHasContent(contents) {
		return fmt.Errorf("contents not found in json body")
	}

	// contents can have different keys with same child structure
	firstPart := jsonGetIndex(jsonGet(
		jsonGetIndex(jsonGet(jsonFirstKey(contents), "tabs"), 0),
		"tabRenderer", "content", "sectionListRenderer", "contents",
	), 0)

	// This extra nested item is only set with the web client
	if n := jsonGetIndex(jsonGet(firstPart, "itemSectionRenderer", "contents"), 0); jsonHasContent(n) {
		firstPart = n
	}

	vJSON := jsonGet(firstPart, "playlistVideoListRenderer", "contents")
	if !jsonHasContent(vJSON) {
		return fmt.Errorf("no video data found in JSON")
	}

	entries, continuation, err := extractPlaylistEntries(vJSON)
	if err != nil {
		return err
	}

	if len(continuation) == 0 {
		continuation = jsonGetContinuation(jsonGet(firstPart, "playlistVideoListRenderer"))
	}

	if len(entries) == 0 {
		return fmt.Errorf("no videos found in playlist")
	}

	p.Videos = entries

	for continuation != "" {
		data := prepareInnertubePlaylistData(continuation, true, *client.ClientType)

		body, err := client.httpPostBodyBytes(ctx, "https://www.youtube.com/youtubei/v1/browse?key="+client.ClientType.Key, data)
		if err != nil {
			return err
		}

		var contRaw json.RawMessage
		if err := json.Unmarshal(body, &contRaw); err != nil {
			return err
		}

		next := jsonGet(
			jsonGetIndex(jsonGet(contRaw, "onResponseReceivedActions"), 0),
			"appendContinuationItemsAction", "continuationItems",
		)

		if !jsonHasContent(next) {
			next = jsonGet(contRaw, "continuationContents", "playlistVideoListContinuation", "contents")
		}

		if !jsonHasContent(next) {
			break
		}

		entries, token, err := extractPlaylistEntries(next)
		if err != nil {
			return err
		}

		if len(token) > 0 {
			continuation = token
		} else {
			continuation = jsonGetContinuation(jsonGet(contRaw, "continuationContents", "playlistVideoListContinuation"))
		}

		p.Videos = append(p.Videos, entries...)
	}

	return nil
}

func extractPlaylistEntries(data json.RawMessage) ([]*PlaylistEntry, string, error) {
	var vids []*videosJSONExtractor

	if err := json.Unmarshal(data, &vids); err != nil {
		return nil, "", err
	}

	entries := make([]*PlaylistEntry, 0, len(vids))

	var continuation string
	for _, v := range vids {
		if v.Renderer == nil {
			if v.Continuation.Endpoint.Command.Token != "" {
				continuation = v.Continuation.Endpoint.Command.Token
			}

			continue
		}

		entry, err := v.PlaylistEntry()
		if err != nil {
			return nil, "", err
		}
		entries = append(entries, entry)
	}

	return entries, continuation, nil
}

type videosJSONExtractor struct {
	Renderer *struct {
		ID        string   `json:"videoId"`
		Title     withRuns `json:"title"`
		Author    withRuns `json:"shortBylineText"`
		Duration  string   `json:"lengthSeconds"`
		Thumbnail struct {
			Thumbnails []Thumbnail `json:"thumbnails"`
		} `json:"thumbnail"`
	} `json:"playlistVideoRenderer"`
	Continuation struct {
		Endpoint struct {
			Command struct {
				Token string `json:"token"`
			} `json:"continuationCommand"`
		} `json:"continuationEndpoint"`
	} `json:"continuationItemRenderer"`
}

func (vje videosJSONExtractor) PlaylistEntry() (*PlaylistEntry, error) {
	ds, err := strconv.Atoi(vje.Renderer.Duration)
	if err != nil {
		return nil, fmt.Errorf("invalid video duration %q: %w", vje.Renderer.Duration, err)
	}
	return &PlaylistEntry{
		ID:         vje.Renderer.ID,
		Title:      vje.Renderer.Title.String(),
		Author:     vje.Renderer.Author.String(),
		Duration:   time.Second * time.Duration(ds),
		Thumbnails: vje.Renderer.Thumbnail.Thumbnails,
	}, nil
}

type withRuns struct {
	Runs []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

func (wr withRuns) String() string {
	if len(wr.Runs) > 0 {
		return wr.Runs[0].Text
	}
	return ""
}
