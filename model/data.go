package model

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
	GetDate() time.Time
	// returns the id of this event
	GetId() int64
	// returns the type of event, if was an event made by the user, or for the user
	MadeByWho() EventCreated
}

// represents a push event made by the user, with a total amount of commits
type PushEvent struct {
	size    int       // the number of commits
	date    time.Time // the time of the event
	id      int64     // the id of the event
	commits []Commit  // the commits made by this event
}

func (e PushEvent) GetDate() time.Time {
	return e.date
}

func (e PushEvent) GetId() int64 {
	return e.id
}

func (e PushEvent) MadeByWho() EventCreated {
	return MADE_BY_USER
}

// represents a feed of non-nil events
type EventFeed chan IEvent
