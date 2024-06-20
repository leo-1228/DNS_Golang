package report

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Report struct {
	screenLastLine int
	records        []*Record
	mux            sync.Mutex
	newBatch       chan bool
	wg             sync.WaitGroup

	RenderSpeed int // 0: realtime, 5: every 5 seconds, 10: every 10 seconds, -1: none

	start time.Time
	total int32

	lastRender time.Time
}
type ReportInfo struct {
	Domain string
	Valid  bool
}

func (r *Report) NewScreen(title string, recordsCount int) {

	du := 0
	tot := int32(atomic.LoadInt32(&r.total))
	if r.total > 0 {
		du = int(time.Since(r.start).Seconds())
	}
	r.start = time.Now()
	atomic.StoreInt32(&r.total, 0)

	line := strings.Repeat("-", 30)
	fmt.Printf("\033[0m%v\033[0m\n", line)
	fmt.Printf("\033[0m%v %d domains has been checked in %ds\033[0m\n", title, tot, du)
	fmt.Printf("\033[0m%v\033[0m\n", line)
	r.screenLastLine = 0
	r.newBatch = make(chan bool)
	r.records = nil
	r.records = []*Record{}

	if r.RenderSpeed >= 0 {
		enters := strings.Repeat("\n", recordsCount)
		fmt.Print(enters)
	}
}

func (r *Report) NewBatch() {
	close(r.newBatch)
	r.wg.Wait()
}
func clearMultiLines(n int) {
	ClearMultiLine := "\033[" + fmt.Sprint(n) + "A"
	fmt.Print(ClearMultiLine)
}
func (r *Report) Render() {
	/*
		if r.RenderSpeed < 0 {
			return
		}

		if r.RenderSpeed == 0 {
			clearMultiLines(len(r.records))
			for i := range r.records {
				fmt.Println(r.records[i].String())
			}

		} else {
			if r.lastRender.IsZero() {
				r.lastRender = time.Now()
			} else {
				if time.Since(r.lastRender) >= time.Duration(r.RenderSpeed)*time.Second {
					clearMultiLines(len(r.records))
					for i := range r.records {
						fmt.Println(r.records[i].String())
					}
					r.lastRender = time.Now()
				}
			}
		}*/
}

func (r *Report) Log(domain string, all int, info <-chan ReportInfo) {
	r.wg.Add(1)
	go func(info <-chan ReportInfo) {
		r.mux.Lock()
		r.records = append(r.records, &Record{
			Id:     0,
			Domain: domain,
			Total:  0,
			Error:  0,
			Valid:  0,
			All:    all,
		})
		r.mux.Unlock()

		for {
			select {
			case <-r.newBatch:
				r.wg.Done()
				return
			case d := <-info:
				r.mux.Lock()
				atomic.AddInt32(&r.total, 1)
				if d.Valid {
					for i := range r.records {
						if r.records[i].Domain == domain {
							r.records[i].IncValid()
							break
						}
					}

				} else {
					for i := range r.records {
						if r.records[i].Domain == domain {
							r.records[i].IncError()
							break
						}
					}
				}
				r.Render()

				r.mux.Unlock()
			}
		}

	}(info)
}

func New(renderSpeed int) *Report {
	return &Report{
		records:     make([]*Record, 0),
		mux:         sync.Mutex{},
		newBatch:    make(chan bool),
		wg:          sync.WaitGroup{},
		RenderSpeed: renderSpeed,
	}
}
