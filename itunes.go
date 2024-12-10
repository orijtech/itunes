// Copyright 2018 Orijtech, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package itunes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

type Client int

const baseURL = "https://itunes.apple.com/search"

var errUnimplemented = errors.New("unimplemented")
var errNilSearch = errors.New("nil search")

func (c *Client) Search(ctx context.Context, s *Search) (*SearchResult, error) {
	ctx, span := trace.StartSpan(ctx, "itunes.(*Client).Search")
	defer span.End()

	if s == nil {
		return nil, errNilSearch
	}

	if s.Id != "" {
		return c.SearchById(ctx, s.Id)
	}

	urlValues, err := valueToURLValues(ctx, s)
	if err != nil {
		return nil, err
	}
	queryString := urlValues.Encode()
	searchURL := fmt.Sprintf("%s?%s", baseURL, queryString)
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	client := &http.Client{Transport: &ochttp.Transport{}}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if !statusOK(res.StatusCode) {
		return nil, fmt.Errorf("status: %s", res.Status)
	}

	blob, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Search: %q => %s\n", queryString, blob)
	sres := new(SearchResult)
	if err := json.Unmarshal(blob, sres); err != nil {
		return nil, err
	}
	for _, res := range sres.Results {
		if res.TrackViewURL == "" {
			continue
			panic("no trackViewURL")
		}
	}

	return sres, nil
}

// The goal of this function is to transform any struct
// into a URL values map.
// type a { A int;B []string;C []float32}{10, ["a","b"], [23.4,-10]} -> A=10&B=a,b&C=23.4,-10
func valueToURLValues(ctx context.Context, ptrVal interface{}) (url.Values, error) {
	_, span := trace.StartSpan(ctx, "itunes.valueToURLValues")
	defer span.End()

	blob, err := json.Marshal(ptrVal)
	if err != nil {
		return nil, err
	}

	shadowMap := make(map[string]interface{})
	if err := json.Unmarshal(blob, &shadowMap); err != nil {
		return nil, err
	}

	outValues := url.Values{}
	for key, value := range shadowMap {
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		default:
			if value == reflect.Invalid {
				continue
			}
			str := fmt.Sprintf("%v", value)
			if str != "" {
				outValues[key] = []string{str}
			}
		case reflect.Array:
			var outL []string
			for i, n := 0, rv.Len(); i < n; i++ {
				ithItem := rv.Index(i)
				if ithItem.Kind() == reflect.Invalid {
					continue
				}
				str := fmt.Sprintf("%v", ithItem.Interface())
				if str != "" {
					outL = append(outL, str)
				}
			}
			if len(outL) >= 1 {
				outValues[key] = outL
			}
		}
	}

	return outValues, nil
}

func statusOK(code int) bool { return code >= 200 && code <= 299 }

type SearchResult struct {
	ResultCount uint64    `json:"resultCount"`
	Results     []*Result `json:"results"`
}

type Result struct {
	Kind              string  `json:"kind"`
	TrackId           uint64  `json:"trackId"`
	CollectionId      uint64  `json:"collectionId"`
	ArtistName        string  `json:"artistName"`
	TrackPrice        float64 `json:"trackPrice"`
	Country           string  `json:"country"`
	Currency          string  `josn:"currency"`
	CollectionName    string  `json:"collectionName"`
	PrimaryGenreName  string  `json:"primaryGenreName"`
	TrackName         string  `json:"trackName"`
	TrackCensoredName string  `json:"trackCensoredName"`
	TrackNumber       uint    `json:"trackNumber"`
	TrackTimeMillis   uint64  `json:"trackTimeMillis"`
	TrackViewURL      string  `json:"trackViewUrl"`
	CollectionPrice   float64 `json:"collectionPrice"`
	CollectionViewURL string  `json:"collectionViewUrl"`
	ArtistViewURL     string  `json:"artistViewUrl"`
	PreviewURL        string  `json:"previewUrl"`
	Streamable        bool    `json:"isStreamable"`
	ArtworkURL100Px   string  `json:"artworkUrl100"`
	ArtworkURL60Px    string  `json:"artworkUrl60"`
	ArtworkURL30Px    string  `json:"artworkUrl30"`
}

func (c *Client) SearchById(ctx context.Context, id string) (*SearchResult, error) {
	qURL := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s", id)
	req, err := http.NewRequestWithContext(ctx, "GET", qURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("failed with %q", res.Status)
	}
	blob, err := io.ReadAll(res.Body)
	res.Body.Close()
	blob = bytes.TrimSpace(blob)

	sres := new(SearchResult)
	if err := json.Unmarshal(blob, sres); err != nil {
		return nil, err
	}
	return sres, nil
}

type Search struct {
	Term            string    `json:"term"`
	Country         Country   `json:"country"`
	Media           Media     `json:"media"`
	Entity          Entity    `json:"entity"`
	Attribute       Attribute `json:"attribute"`
	Language        Language  `json:"lang"`
	Limit           uint      `json:"limit"`
	Version         string    `json:"version"`
	ExplicitContent bool      `json:"explicit"`
	Id              string    `json:"id"`
}

type Country string
type Language string
type Media string
type Attribute string

type Entity string

const (
	EntityMovie           Entity = "movie"
	EntityMovieArtist     Entity = "movieArtist"
	EntityPodcast         Entity = "podcast"
	EntityPodcastAuthor   Entity = "podcastAuthor"
	EntityMusic           Entity = "music"
	EntityMusicVideo      Entity = "musicVideo"
	EntityMusicArtist     Entity = "musicArtist"
	EntityAudioBook       Entity = "audiobook"
	EntityAudioBookAuthor Entity = "audiobookAuthor"
	EntityShortFilm       Entity = "shortFilm"
	EntityShortFilmArtist Entity = "shortFilmArtist"
	EntityTVShow          Entity = "tvShow"
	EntityTVEpisode       Entity = "tvEpisode"
	EntityTVSeason        Entity = "tvSeason"
	EntitySoftware        Entity = "software"
	EntityIPadSoftware    Entity = "iPadSoftware"
	EntityMacSoftware     Entity = "macSoftware"
	EntityEBook           Entity = "ebook"
	EntityAll             Entity = "all"
	EntityAllTrack        Entity = "allTrack"
)
