package youtube

import (
	"flag"
	"log"
	"net/http"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

const developerKey = "AIzaSyCiltKz_TjHfBQaqPTgK5nw-jqtnzU7p-k"

func SearchVideo(searchKeyword string, maxResults int64) (string, string) {
	//query = flag.String("query", searchKeyword, "Search term")
	//maxResults = flag.Int64("max-results", 1, "Max YouTube results")

	flag.Parse()

	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(searchKeyword).
		MaxResults(maxResults)
	response, err := call.Do()
	handleError(err, "")

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
		case "youtube#channel":
			channels[item.Id.ChannelId] = item.Snippet.Title
		case "youtube#playlist":
			playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
	}

	var videoId string
	var videoTitle string
	for id, title := range videos {
		videoId = id
		videoTitle = title
	}
	return videoId, videoTitle
}
