package ctrl

import (
	"fmt"
	"time"

	"github.com/elleven11/gophergotchi/model"
)

// represents the context for a controller
type CtrlContext struct {
	reqCtx model.ReqContext
	evFeed model.EventFeed
	user   string
}

// creates the initial controller, and gets the latest feed
func MakeController(reqCtx model.ReqContext, user string) CtrlContext {
	return CtrlContext{reqCtx: reqCtx, evFeed: reqCtx.GetFeedByUser(user), user: user}
}

// runs the program
func (c CtrlContext) Run() {
	var prevEvent model.IEvent
	for {
		select {
		case ev := <-c.evFeed:
			fmt.Printf("%v\n", ev.GetDate())
			prevEvent = ev
		default:
			c.updateRoutine(prevEvent)
		}
	}
}

// a routine that runs updates on the feed every 10 seconds until a new event is found.
func (c *CtrlContext) updateRoutine(prevEvent model.IEvent) {
	for {
		time.Sleep(10 * time.Second)
		if c.replenishFeedIfNew(prevEvent) {
			return
		}
		fmt.Println("same ev..")
	}
}

// checks the latest new event, and if its new (based if its not equal to the given event),
// add it onto the event feed. returns true if the event gets updated.
func (c *CtrlContext) replenishFeedIfNew(prevEvent model.IEvent) bool {
	ev, err := c.reqCtx.GetLastEventByUser(c.user)

	// return false on connection errors
	if err != nil {
		return false
	}

	if ev != nil && ev.GetId() != prevEvent.GetId() {
		c.evFeed <- ev
		return true
	}
	return false
}
