package meta

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	gainStation = iota
	gainLocation
	gainChannel
	gainScaleFactor
	gainScaleBias
	gainStart
	gainEnd
	gainLast
)

// Gain defines times where sensor scaling or offsets are needed.
type Gain struct {
	Span
	Scale

	Station  string
	Location string
	Channel  string
}

// Id returns a unique string which can be used for sorting or checking.
func (g Gain) Id() string {
	return strings.Join([]string{g.Station, g.Location, g.Channel}, ":")
}

// Less returns whether one Gain sorts before another.
func (g Gain) Less(gain Gain) bool {
	switch {
	case g.Station < gain.Station:
		return true
	case g.Station > gain.Station:
		return false
	case g.Location < gain.Location:
		return true
	case g.Location > gain.Location:
		return false
	case g.Channel < gain.Channel:
		return true
	case g.Channel > gain.Channel:
		return false
	case g.Span.Start.Before(gain.Span.Start):
		return true
	default:
		return false
	}
}

// Channels returns a sorted slice of single defined components.
func (g Gain) Channels() []string {
	var comps []string
	for _, c := range g.Channel {
		comps = append(comps, string(c))
	}
	return comps
}

// Gains returns a sorted slice of single Gain entries.
func (g Gain) Gains() []Gain {
	var gains []Gain
	for _, c := range g.Channel {
		gains = append(gains, Gain{
			Span:     g.Span,
			Scale:    g.Scale,
			Station:  g.Station,
			Location: g.Location,
			Channel:  string(c),
		})
	}

	sort.Slice(gains, func(i, j int) bool { return gains[i].Less(gains[j]) })

	return gains
}

type GainList []Gain

func (s GainList) Len() int           { return len(s) }
func (s GainList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s GainList) Less(i, j int) bool { return s[i].Less(s[j]) }

func (s GainList) encode() [][]string {
	data := [][]string{{
		"Station",
		"Location",
		"Channel",
		"Scale Factor",
		"Scale Bias",
		"Start Date",
		"End Date",
	}}

	for _, v := range s {
		data = append(data, []string{
			strings.TrimSpace(v.Station),
			strings.TrimSpace(v.Location),
			strings.TrimSpace(v.Channel),
			strings.TrimSpace(v.factor),
			strings.TrimSpace(v.bias),
			v.Start.Format(DateTimeFormat),
			v.End.Format(DateTimeFormat),
		})
	}

	return data
}

func (s *GainList) decode(data [][]string) error {
	var gains []Gain
	if len(data) > 1 {
		for _, d := range data[1:] {
			if len(d) != gainLast {
				return fmt.Errorf("incorrect number of installed gain fields")
			}
			var err error

			var factor, bias float64
			switch {
			case d[gainScaleFactor] != "":
				if factor, err = strconv.ParseFloat(d[gainScaleFactor], 64); err != nil {
					return err
				}
			default:
				factor = 1.0
			}
			if d[gainScaleBias] != "" {
				if bias, err = strconv.ParseFloat(d[gainScaleBias], 64); err != nil {
					return err
				}
			}

			var start, end time.Time
			if start, err = time.Parse(DateTimeFormat, d[gainStart]); err != nil {
				return err
			}
			if end, err = time.Parse(DateTimeFormat, d[gainEnd]); err != nil {
				return err
			}

			gains = append(gains, Gain{
				Span: Span{
					Start: start,
					End:   end,
				},
				Scale: Scale{
					Factor: factor,
					Bias:   bias,

					factor: strings.TrimSpace(d[gainScaleFactor]),
					bias:   strings.TrimSpace(d[gainScaleBias]),
				},
				Station:  strings.TrimSpace(d[gainStation]),
				Location: strings.TrimSpace(d[gainLocation]),
				Channel:  strings.TrimSpace(d[gainChannel]),
			})
		}

		*s = GainList(gains)
	}

	return nil
}

func LoadGains(path string) ([]Gain, error) {
	var s []Gain

	if err := LoadList(path, (*GainList)(&s)); err != nil {
		return nil, err
	}

	sort.Sort(GainList(s))

	return s, nil
}
