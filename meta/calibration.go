package meta

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	calibrationMake = iota
	calibrationModel
	calibrationSerial
	calibrationComponent
	calibrationSensitivity
	calibrationFrequency
	calibrationStart
	calibrationEnd
	calibrationLast
)

type Calibration struct {
	Install

	sensitivity string
	frequency   string

	Sensitivity float64
	Frequency   float64

	Component int
}

// Id returns a unique string which can be used for sorting or checking.
func (c Calibration) Id() string {
	return strings.Join([]string{c.Make, c.Model, c.Serial, strconv.Itoa(c.Component)}, ":")
}

// Less returns whether one Calibration sorts before another.
func (s Calibration) Less(calibration Calibration) bool {
	switch {
	case s.Install.Less(calibration.Install):
		return true
	case calibration.Install.Less(s.Install):
		return false
	case s.Component < calibration.Component:
		return true
	default:
		return false
	}
}

type CalibrationList []Calibration

func (s CalibrationList) Len() int           { return len(s) }
func (s CalibrationList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s CalibrationList) Less(i, j int) bool { return s[i].Less(s[j]) }

func (s CalibrationList) encode() [][]string {
	data := [][]string{{
		"Make",
		"Model",
		"Serial",
		"Component",
		"Sensitivity",
		"Frequency",
		"Start Date",
		"End Date",
	}}

	for _, v := range s {
		data = append(data, []string{
			strings.TrimSpace(v.Make),
			strings.TrimSpace(v.Model),
			strings.TrimSpace(v.Serial),
			strconv.Itoa(v.Component),
			strings.TrimSpace(v.sensitivity),
			strings.TrimSpace(v.frequency),
			v.Start.Format(DateTimeFormat),
			v.End.Format(DateTimeFormat),
		})
	}
	return data
}
func (s *CalibrationList) decode(data [][]string) error {
	var calibrations []Calibration
	if len(data) > 1 {
		for _, d := range data[1:] {
			if len(d) != calibrationLast {
				return fmt.Errorf("incorrect number of installed calibration fields")
			}
			var err error

			var sens, freq float64
			if sens, err = strconv.ParseFloat(d[calibrationSensitivity], 64); err != nil {
				return err
			}
			if d[calibrationFrequency] != "" {
				if freq, err = strconv.ParseFloat(d[calibrationFrequency], 64); err != nil {
					return err
				}
			}

			var comp int
			if d[calibrationComponent] != "" {
				if comp, err = strconv.Atoi(d[calibrationComponent]); err != nil {
					return err
				}
			}

			var start, end time.Time
			if start, err = time.Parse(DateTimeFormat, d[calibrationStart]); err != nil {
				return err
			}
			if end, err = time.Parse(DateTimeFormat, d[calibrationEnd]); err != nil {
				return err
			}

			calibrations = append(calibrations, Calibration{
				Install: Install{
					Equipment: Equipment{
						Make:   strings.TrimSpace(d[calibrationMake]),
						Model:  strings.TrimSpace(d[calibrationModel]),
						Serial: strings.TrimSpace(d[calibrationSerial]),
					},
					Span: Span{
						Start: start,
						End:   end,
					},
				},
				Sensitivity: sens,
				Frequency:   freq,
				Component:   comp,

				sensitivity: strings.TrimSpace(d[calibrationSensitivity]),
				frequency:   strings.TrimSpace(d[calibrationFrequency]),
			})
		}
	}

	*s = CalibrationList(calibrations)

	return nil
}

func LoadCalibrations(path string) ([]Calibration, error) {
	var s []Calibration

	if err := LoadList(path, (*CalibrationList)(&s)); err != nil {
		return nil, err
	}

	sort.Sort(CalibrationList(s))

	return s, nil
}
