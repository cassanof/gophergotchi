package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elleven11/gophergotchi/utils"
)

// get the latest 67 events made by the given user
func GetFeedByUser(user string) interface{} {
	// TODO: change back to 67
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/events?per_page=30", user))
	utils.Check(err)

	body, err := ioutil.ReadAll(resp.Body)
	utils.Check(err)

	var objmap []map[string]json.RawMessage
	err = json.Unmarshal(body, &objmap)
	utils.Check(err)

	// feed := make(EventFeed, 67)
	feed := make([]interface{}, 67)
	var wg sync.WaitGroup
	for i, ev := range objmap {
		wg.Add(1)
		go func(i int, ev map[string]json.RawMessage) {
			feed[i] = parseEvent(ev)
			wg.Done()
		}(i, ev)
	}
	wg.Wait()

	return feed
}

// parses the given event json map into a proper proper event data structure
func parseEvent(data map[string]json.RawMessage) IEvent {
	evType := strings.Trim(string(data["type"]), "\"")
	switch evType {
	case "PushEvent":
		return parsePushEvent(data)
	}

	// if we don't care about the given one, we return nil.
	return nil
}

// parses the json RawMessage ISO 8601 into a time.Time data struct
func parseDate(dateData json.RawMessage) time.Time {
	dateStr := strings.Trim(string(dateData), "\"")
	t, err := time.Parse("2006-01-02T15:04:05Z", dateStr)
	utils.Check(err)
	return t
}

// parses the given data into a PushEvent. it must be validated before,
// to ensure that this is indeed a push event
func parsePushEvent(data map[string]json.RawMessage) PushEvent {
	date := parseDate(data["created_at"])

	// get payload
	var payload map[string]json.RawMessage
	err := json.Unmarshal(data["payload"], &payload)
	utils.Check(err)

	// get size
	size, err := strconv.Atoi(strings.Trim(string(payload["size"]), "\""))
	utils.Check(err)

	var commitData []map[string]json.RawMessage
	err = json.Unmarshal(payload["commits"], &commitData)
	utils.Check(err)

	commits := parseCommits(commitData)

	return PushEvent{Size: size, Date: date, Commits: commits}
}

// parses the given list of json into a list of commit stats. it must be validated before,
// to ensure that these are indeed commit stats
func parseCommits(commitsData []map[string]json.RawMessage) []Commit {
	commits := make([]Commit, len(commitsData))

	var wg sync.WaitGroup
	for i, commitData := range commitsData {
		wg.Add(1)
		go func(i int, commitData map[string]json.RawMessage) {
			url := strings.Trim(string(commitData["url"]), "\"")

			resp, err := http.Get(url)
			utils.Check(err)

			body, err := ioutil.ReadAll(resp.Body)
			utils.Check(err)

			var innerCommitData map[string]json.RawMessage
			err = json.Unmarshal(body, &innerCommitData)
			utils.Check(err)

			var filesData []map[string]json.RawMessage
			err = json.Unmarshal(innerCommitData["files"], &filesData)
			utils.Check(err)

			additions := 0
			deletions := 0

			for _, fData := range filesData {
				filename := strings.Trim(string(fData["filename"]), "\"")
				if fileNameIsCode(filename) {
					adds, err := strconv.Atoi(strings.Trim(string(fData["additions"]), "\""))
					utils.Check(err)
					additions += adds

					dels, err := strconv.Atoi(strings.Trim(string(fData["deletions"]), "\""))
					utils.Check(err)
					deletions += dels
				}
			}

			commits[i] = Commit{Additions: additions, Deletions: deletions}
			wg.Done()
		}(i, commitData)
	}
	wg.Wait()

	return commits
}

// determines if the given file is code or not
func fileNameIsCode(filename string) bool {
	// TODO: make real check
	return true || len(filename) == 0
}
