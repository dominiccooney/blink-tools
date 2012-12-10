package main

import (
	"fmt"
	"time"
	"history"
)

const (
	NO_FLAGS = iota
	REVIEW
	REVIEW_RESPONSE
)

type Attachment struct {
	state int
	timeReviewRequested time.Time
	observeRplus chan<- time.Duration
	observeRminus chan<- time.Duration
}

func (a *Attachment) UpdateFlags(when time.Time, removed []history.AttachmentFlag, added []history.AttachmentFlag) {
	switch {
	case len(added) > 0 && added[0] == history.Review && a.state != REVIEW:
		a.state = REVIEW
		a.timeReviewRequested = when
	case len(added) > 0 && added[0] == history.ReviewGranted && a.state == REVIEW:
		a.state = REVIEW_RESPONSE
		a.observeRplus <- when.Sub(a.timeReviewRequested)
	case len(added) > 0 && added[0] == history.ReviewDenied && a.state == REVIEW:
		a.state = REVIEW_RESPONSE
		a.observeRminus <- when.Sub(a.timeReviewRequested)
	}
}

type Observer struct {
	observeAttachment chan<- int
	observeRplus chan<- time.Duration
	observeRminus chan<- time.Duration
	attachments map[int] *Attachment
}

func newObserver(observeAttachment chan<- int, observeRplus chan<- time.Duration, observeRminus chan<- time.Duration) (o *Observer) {
	o = &Observer{observeAttachment, observeRplus, observeRminus, make(map[int] *Attachment)}
	return
}

func (o *Observer) getAttachment(id int) *Attachment {
	if a, ok := o.attachments[id]; ok {
		return a
	}
	return o.makeAttachment(id)
}

func (o *Observer) makeAttachment(id int) (a *Attachment) {
	o.observeAttachment <- 1
	a = &Attachment{state: NO_FLAGS, observeRplus: o.observeRplus, observeRminus: o.observeRminus}
	o.attachments[id] = a
	return
}

func (o *Observer) AttachmentFlag(when time.Time, id int, removed []history.AttachmentFlag, added []history.AttachmentFlag) {
	a := o.getAttachment(id)
	a.UpdateFlags(when, removed, added)
}

func consumer(ids <-chan int, observeAttachment chan<- int, observeRplus chan<- time.Duration, observeRminus chan<- time.Duration, done chan<- int) {
	for {
		id := <-ids
		consumeOne(id, observeAttachment, observeRplus, observeRminus)
		done <- 1
	}
}

func consumeOne(id int, observeAttachment chan<- int, observeRplus chan<- time.Duration, observeRminus chan<- time.Duration) {
	fmt.Print(".")
	b := &history.Bug{Id: id}
	o := newObserver(observeAttachment, observeRplus, observeRminus)
	err := history.ReplayHistory(b, o)
	if err != nil {
		fmt.Println(err)
	}
}

type Observations struct {
	nrplus int
	nrminus int
	nattachments int
	nbugs int
	elapsedTimes []time.Duration
	observeAttachment chan int
	observeRplus chan time.Duration
	observeRminus chan time.Duration
}

func newObservations() (o *Observations) {
	o = &Observations{}
	o.elapsedTimes = make([]time.Duration, 0, 1000)
	o.observeAttachment = make(chan int)
	o.observeRplus = make(chan time.Duration)
	o.observeRminus = make(chan time.Duration)
	return
}

func (o *Observations) collectResults(done <-chan int) {
	for {
		select {
		case <-o.observeAttachment:
			o.nattachments++
		case d := <-o.observeRplus:
			o.nrplus++
			o.elapsedTimes = append(o.elapsedTimes, d)
		case d := <-o.observeRminus:
			o.nrminus++
			o.elapsedTimes = append(o.elapsedTimes, d)
		case <-done:
			return
		}
	}
}

func (o *Observations) report() {
	fmt.Printf("number of bugs: %d\n", o.nbugs)
	fmt.Printf("number of attachments: %d\n", o.nattachments)
	fmt.Printf("number of R+: %d\n", o.nrplus)
	fmt.Printf("number of R-: %d\n", o.nrminus)

	for _, d := range(o.elapsedTimes) {
		fmt.Println(d.Hours())
	}
}

func main() {
	done := make(chan int)
	o := newObservations()
	go o.collectResults(done)

	ids := make(chan int, 1000)
	doneUnit := make(chan int, 1000)
	for i := 0; i < 10; i++ {
		go consumer(ids, o.observeAttachment, o.observeRplus, o.observeRminus, doneUnit)
	}

	n := 0
	for {
		var id int
		_, err := fmt.Scan(&id)
		if err != nil {
			break
		}
		ids <- id
		n++
	}
	o.nbugs = n
	for i := 0; i < n; i++ {
		<-doneUnit
	}
	done <- 1
	fmt.Println()
	o.report()
}
