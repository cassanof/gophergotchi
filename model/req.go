package model

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

type ReqContext struct {
	apikey   string
	username string
	client   *http.Client
}

func MakeReqContext(apikey string, username string) ReqContext {
	req := ReqContext{apikey: apikey, username: username, client: nil}
	client := &http.Client{
		CheckRedirect: req.redirectPolicyFunc,
	}
	req.client = client
	return req
}

// get the latest 67 events made by the given user. errors are unrecoverable.
func (r ReqContext) GetFeedByUser(user string) EventFeed {
	resp, err := r.makeGetReq(fmt.Sprintf("https://api.github.com/users/%s/events?per_page=67", user))
	utils.Check(err)

	body, err := ioutil.ReadAll(resp.Body)
	utils.Check(err)

	var objmap []map[string]json.RawMessage
	err = json.Unmarshal(body, &objmap)
	utils.Check(err)

	// make a list, to then be dumped into the channel. append in reverse
	listLen := 67
	events := make([]IEvent, listLen)
	var wg sync.WaitGroup
	for i, ev := range objmap {
		wg.Add(1)
		go func(i int, ev map[string]json.RawMessage) {
			events[listLen-i-1] = r.parseEvent(ev)
			wg.Done()
		}(i, ev)
	}
	wg.Wait()

	feed := make(EventFeed, 67)

	// dump list into channel, only non-nil elements
	for _, e := range events {
		if e != nil {
			feed <- e
		}
	}

	return feed
}

// gets the last event made by the user, returns an error if the request failed
func (r ReqContext) GetLastEventByUser(user string) (IEvent, error) {
	resp, err := r.makeGetReq(fmt.Sprintf("https://api.github.com/users/%s/events?per_page=1", user))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	utils.Check(err)

	var objmap []map[string]json.RawMessage
	err = json.Unmarshal(body, &objmap)
	utils.Check(err)

	return r.parseEvent(objmap[0]), nil
}

func (r ReqContext) redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	req.SetBasicAuth(r.username, r.apikey)
	return nil
}

func (r ReqContext) makeGetReq(url string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		url,
		nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(r.username, r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// parses the given event json map into a proper proper event data structure
func (r ReqContext) parseEvent(data map[string]json.RawMessage) IEvent {
	evType := strings.Trim(string(data["type"]), "\"")
	switch evType {
	case "PushEvent":
		return r.parsePushEvent(data)
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
func (r ReqContext) parsePushEvent(data map[string]json.RawMessage) PushEvent {
	date := parseDate(data["created_at"])

	// get payload
	var payload map[string]json.RawMessage
	err := json.Unmarshal(data["payload"], &payload)
	utils.Check(err)

	// get size
	size, err := strconv.Atoi(strings.Trim(string(payload["size"]), "\""))
	utils.Check(err)

	// get id
	id, err := strconv.ParseInt(strings.Trim(string(data["id"]), "\""), 10, 64)

	var commitData []map[string]json.RawMessage
	err = json.Unmarshal(payload["commits"], &commitData)
	utils.Check(err)

	// commits := parseCommits(commitData)
	commits := []Commit{}

	return PushEvent{size: size, date: date, commits: commits, id: id}
}

// parses the given list of json into a list of commit stats. it must be validated before,
// to ensure that these are indeed commit stats. errors are unrecoverable.
func (r ReqContext) parseCommits(commitsData []map[string]json.RawMessage) []Commit {
	commits := make([]Commit, len(commitsData))

	var wg sync.WaitGroup
	for i, commitData := range commitsData {
		wg.Add(1)
		go func(i int, commitData map[string]json.RawMessage) {
			url := strings.Trim(string(commitData["url"]), "\"")

			resp, err := r.makeGetReq(url)
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
