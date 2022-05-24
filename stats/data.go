package stats

import "time"

// represent an enumeration of who made the event, either the user, or someone for the user
type EventCreated uint

const (
	MADE_BY_USER EventCreated = iota
	MADE_FOR_USER
)

// represents stats about a commit. these stats are only for real code, we check the file extension
// of each file that has been committed
type Commit struct {
	Additions int
	Deletions int
}

// represents the interface for an event
type IEvent interface {
	// returns the date of creation
	getDate() time.Time
	// returns the type of event, if was an event made by the user, or for the user
	madeByWho() EventCreated
}

// represents a push event made by the user, with a total amount of commits
type PushEvent struct {
	Size    int       // the number of commits
	Date    time.Time // the time of the event
	Commits []Commit
}

func (e PushEvent) getDate() time.Time {
	return e.Date
}

func (e PushEvent) madeByWho() EventCreated {
	return MADE_BY_USER
}

// represents a feed of events. 67 events max.
type EventFeed []IEvent
