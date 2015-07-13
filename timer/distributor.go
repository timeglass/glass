package timer

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

//A block of time that is positioned
//on a timeline
type Block struct {
	Staged bool          `json:"s"`
	Width  time.Duration `json:"w"`
	Time   time.Time     `json:"t"`
}

//A timeline holds a number of lines
//for a single file
type Timeline struct {
	Blocks []*Block `json:"b"`
}

func NewTimeline() *Timeline {
	return &Timeline{}
}

func (tl *Timeline) Length(upto time.Time) time.Duration {
	res := time.Millisecond * 0
	for _, b := range tl.Blocks {
		if math.Floor(upto.Sub(b.Time).Seconds()) >= 0 {
			res += b.Width
		}
	}

	return res
}

func (tl *Timeline) Staged() time.Duration {
	res := time.Millisecond * 0
	for _, b := range tl.Blocks {
		if b.Staged {
			res += b.Width
		}
	}

	return res
}

func (tl *Timeline) Unstaged() time.Duration {
	res := time.Millisecond * 0
	for _, b := range tl.Blocks {
		if !b.Staged {
			res += b.Width
		}
	}

	return res
}

func (tl *Timeline) Stage(upto time.Time) {
	for _, b := range tl.Blocks {
		if math.Floor(upto.Sub(b.Time).Seconds()) >= 0 {
			b.Staged = true
		}
	}
}

func (tl *Timeline) Expand(w time.Duration, t time.Time) {
	tl.Blocks = append(tl.Blocks, &Block{false, w, t})
}

//A distributor managed various files
//for a single timer using timelines
type Distributor struct {
	data *distrData
}

type distrData struct {
	ActiveFiles map[string]string    `json:"af"`
	Timelines   map[string]*Timeline `json:"tl"`
}

var OverheadTimeline = "__overhead"

func NewDistributor() *Distributor {
	d := &Distributor{
		data: &distrData{},
	}

	d.Reset(false)
	return d
}

func (d *Distributor) Break() {
	d.data.ActiveFiles = map[string]string{}
}

func (d *Distributor) Reset(staged bool) {
	d.Break()
	if staged {
		//only remove staged blocks
		for _, tl := range d.data.Timelines {
			blcks := []*Block{}
			for _, b := range tl.Blocks {
				if !b.Staged {
					blcks = append(blcks, b)
				}
			}

			tl.Blocks = blcks
		}
	} else {
		//reset all
		d.data.Timelines = map[string]*Timeline{
			OverheadTimeline: NewTimeline(),
		}
	}
}

func (d *Distributor) Distribute(dur time.Duration, t time.Time) {
	if len(d.data.ActiveFiles) == 0 {
		return
	}

	partd := dur.Nanoseconds() / int64(len(d.data.ActiveFiles))
	for path, _ := range d.data.ActiveFiles {
		if tl, ok := d.data.Timelines[path]; !ok {

			//@todo no timeline while it should, emit error?
			continue
		} else {
			tl.Expand(time.Duration(partd), t)
		}
	}
}

func (d *Distributor) Register(fpath string) {
	if fpath == "" {
		fpath = OverheadTimeline
	}

	var tl *Timeline
	var ok bool
	if tl, ok = d.data.Timelines[fpath]; !ok {
		tl = NewTimeline()
		d.data.Timelines[fpath] = tl
	}

	if _, ok = d.data.ActiveFiles[fpath]; !ok {
		d.data.ActiveFiles[fpath] = ""
	}
}

//stage blocks to be removed on the next .Reset(true)
func (d *Distributor) Stage(fpath string, upto time.Time) {
	if fpath == "" {
		fpath = OverheadTimeline
	}

	//no timeline is a noop
	if tl, ok := d.data.Timelines[fpath]; ok {
		tl.Stage(upto)
	}
}

//extract the time spent on a file from the first point a given point in time
func (d *Distributor) Extract(fpath string, upto time.Time) (time.Duration, error) {
	if fpath == "" {
		fpath = OverheadTimeline
	}

	if tl, ok := d.data.Timelines[fpath]; !ok {
		return 0, fmt.Errorf("No known timeline for file '%s'", fpath)
	} else {
		return tl.Length(upto), nil
	}
}

func (d *Distributor) Timelines() map[string]*Timeline {
	return d.data.Timelines
}

func (d *Distributor) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &d.data)
	if err != nil {
		return err
	}

	return nil
}

func (d *Distributor) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.data)
}
