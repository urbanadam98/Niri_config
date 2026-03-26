package main

import (
	//"encoding/json"
	"fmt"
	// 	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	eventRRuleRegex      = regexp.MustCompile(`RRULE:.*?\n`)
	freqRegex            = regexp.MustCompile(`FREQ=.*?;`)
	eventSummaryRegex    = regexp.MustCompile(`SUMMARY:.*?\n`)
	eventFreqWeeklyRegex = regexp.MustCompile(`RRULE:FREQ=WEEKLY\n`)
	eventFreqYearlyRegex = regexp.MustCompile(`RRULE:FREQ=YEARLY\n`)
)

// unixtimestamp
const (
	uts = "1136239445"
	//ics date time format
	// Y-m-d H:i:S time format
	YmdHis = "2006-01-02 15:04:05"
)

func trimField(field, cutset string) string {
	re, _ := regexp.Compile(cutset)
	cutsetRem := re.ReplaceAllString(field, "")
	return strings.TrimRight(cutsetRem, "\r\n")
}

func explodeEvent(eventData *string) (string, string) {
	reEvent, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	Event := reEvent.FindString(*eventData)
	calInfo := reEvent.ReplaceAllString(*eventData, "")
	return Event, calInfo
}
func splitIcal(ical string) []string {
	splits := regexp.MustCompile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	//reEvent, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	Events := splits.FindAllString(ical, -1)
	/*for i := range Events {
		//fmt.Println(eventData)
		fmt.Println(i)
		fmt.Println(Events[i])
	}*/
	//fmt.Println(Events[1])
	//os.Exit(1)
	return Events
}

func parseIcalName(eventData string) string {
	re, _ := regexp.Compile(`X-WR-CALNAME:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "X-WR-CALNAME:")
}

func parseTimeField(fieldName string, eventData string) (time.Time, string) {
	reWholeDay, _ := regexp.Compile(fmt.Sprintf(`%s;VALUE=DATE:.*?\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))
	// correct regex: .+:(.+)$
	re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.+))?(;VALUE=DATE-TIME)?:(.+?)\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))

	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time
	var thisTime time.Time
	//var thisTime time.Time
	//var datetime time.Time
	var tzID string
	var format string

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, fmt.Sprintf("%s;VALUE=DATE:", fieldName))
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		result := re.FindStringSubmatch(eventData)
		//fmt.Println(result)

		if result == nil || len(result) < 4 {
			return t, tzID
		}

		tzID = result[2]
		//fmt.Println(tzID)
		dt := strings.Trim(result[4], "\r") // trim these newlines!

		if strings.HasSuffix(dt, "Z") {
			// If string ends in 'Z', timezone is UTC
			format = "20060102T150405Z"
			thisTime, _ := time.Parse(format, dt)
			//fmt.Println(thisTime)
			t = thisTime.Local()
		} else if tzID != "" {
			format = "20060102T150405"
			location, err := time.LoadLocation(tzID)
			//fmt.Println(location)
			// if tzID not readable use configured timezone
			if err != nil {
				location, _ = time.LoadLocation(config.Timezone)
				// timezone from defines gives CEST, which is not working with parseinlocation:
				//location, _ = time.LoadLocation(timezone)
			}
			// set foreign timezone
			thisTime, _ = time.ParseInLocation(format, dt, location)
			// convert to local timezone
			//t = time.In(myLocation)
			t = thisTime.Local()
		} else {
			// Floating time, use configured timezone
			format = "20060102T150405"
			location, _ := time.LoadLocation(config.Timezone)
			t, _ = time.ParseInLocation(format, dt, location)
		}

	}

	return t, tzID
}

func parseEventStart(eventData *string) (time.Time, string) {
	return parseTimeField("DTSTART", *eventData)
}

func parseEventEnd(eventData *string) (time.Time, string) {
	return parseTimeField("DTEND", *eventData)
}

// ParseISODuration parses an ISO 8601 duration string and returns a time.Duration
// Supports full ISO 8601 format: P[n]Y[n]M[n]DT[n]H[n]M[n]S with flexible ordering and fractions
func ParseISODuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, nil
	}

	// Check for 'P' prefix
	if !strings.HasPrefix(durationStr, "P") {
		return 0, fmt.Errorf("invalid ISO 8601 duration: missing 'P' prefix")
	}
	durationStr = durationStr[1:]

	// Split by 'T' to separate date and time parts
	parts := strings.SplitN(durationStr, "T", 2)
	datePart := parts[0]
	timePart := ""
	if len(parts) > 1 {
		timePart = parts[1]
	}

	var totalSeconds float64

	// Regex to match number (with optional fraction) followed by unit
	componentRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)([YMWDHMS])`)

	// Parse date part (Y, M=months, W, D)
	if datePart != "" {
		matches := componentRegex.FindAllStringSubmatch(datePart, -1)
		matchedStr := ""
		for _, match := range matches {
			matchedStr += match[0]
			if len(match) != 3 {
				continue
			}
			value, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number in duration: %s", match[1])
			}
			unit := match[2]
			switch unit {
			case "Y":
				totalSeconds += value * 365.25 * 24 * 3600 // approximate year
			case "M":
				totalSeconds += value * 30.44 * 24 * 3600 // approximate month
			case "W":
				totalSeconds += value * 7 * 24 * 3600
			case "D":
				totalSeconds += value * 24 * 3600
			default:
				return 0, fmt.Errorf("invalid unit in date part: %s", unit)
			}
		}
		if matchedStr != datePart {
			return 0, fmt.Errorf("invalid characters in date part: %s", datePart)
		}
	}

	// Parse time part (H, M=minutes, S)
	if timePart != "" {
		matches := componentRegex.FindAllStringSubmatch(timePart, -1)
		matchedStr := ""
		for _, match := range matches {
			matchedStr += match[0]
			if len(match) != 3 {
				continue
			}
			value, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number in duration: %s", match[1])
			}
			unit := match[2]
			switch unit {
			case "H":
				totalSeconds += value * 3600
			case "M":
				totalSeconds += value * 60
			case "S":
				totalSeconds += value
			default:
				return 0, fmt.Errorf("invalid unit in time part: %s", unit)
			}
		}
		if matchedStr != timePart {
			return 0, fmt.Errorf("invalid characters in time part: %s", timePart)
		}
	}

	return time.Duration(totalSeconds * float64(time.Second)), nil
}

func parseEventDuration(eventData *string) time.Duration {
	reDuration, _ := regexp.Compile(`DURATION:.*?\n`)
	result := reDuration.FindString(*eventData)
	trimmed := trimField(result, "DURATION:")
	parsedDuration, err := ParseISODuration(trimmed)
	var output time.Duration

	if err == nil {
		output = parsedDuration
	}

	return output
}

func parseEventSummary(eventData *string) string {
	re, _ := regexp.Compile(`SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, `SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?:`)
}

func parseEventDescription(eventData *string) string {
	re, _ := regexp.Compile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)

	resultA := re.FindAllString(*eventData, -1)
	result := strings.Join(resultA, ", ")
	result = strings.Replace(result, "\n ", "", -1)

	result = strings.Replace(result, "\\N", "\n", -1)
	//better := strings.Replace(re.FindString(result), "\n ", "", -1)
	//better = strings.Replace(better, "\\n", " ", -1)
	//better = strings.Replace(better, "\\", "", -1)

	//return trimField(better, "DESCRIPTION:")
	//return trimField(result, "DESCRIPTION:")
	return trimField(strings.Replace(result, "\r\n ", "", -1), "DESCRIPTION:")
}

func parseEventLocation(eventData *string) string {
	re, _ := regexp.Compile(`LOCATION:.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, "LOCATION:")
}

func parseEventAttendees(eventData *string) []string {
	//re, _ := regexp.Compile(`ATTENDEE;.*?\n`)
	re, _ := regexp.Compile(`ATTENDEE;.+\"(.+?)\".*\n`)
	attendeesstring := re.FindAllString(*eventData, -1)
	var attendees []string

	for i := range attendeesstring {
		//fmt.Println(eventData)
		result := re.FindStringSubmatch(attendeesstring[i])
		attendees = append(attendees, result[1])

		//attendee := trimField(attendees[i], `ATTENDEE;.*\"`)
		//fmt.Println(result[1])
	}

	return attendees
}

func parseEventRRule(eventData *string) string {
	re, _ := regexp.Compile(`RRULE:.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, "RRULE:")
}

func parseEventFreq(eventData *string) string {
	re, _ := regexp.Compile(`FREQ=[^;]*(;){0,1}`)
	result := re.FindString(parseEventRRule(eventData))
	return trimField(result, `(FREQ=|;)`)
}

func parseEventUntil(eventData *string) (time.Time, error) {
	re, err := regexp.Compile(`UNTIL=([^;\r\n]+)`)
	if err != nil {
		return time.Time{}, err
	}
	match := re.FindStringSubmatch(*eventData)
	if len(match) < 2 {
		return time.Time{}, nil // no UNTIL found
	}
	value := match[1]
	var parsed time.Time
	var parseErr error
	if strings.HasSuffix(value, "Z") {
		parsed, parseErr = time.Parse(IcsFormatZ, value)
	} else {
		parsed, parseErr = time.Parse(IcsFormat, value)
	}
	if parseErr != nil {
		return time.Time{}, parseErr
	}
	return parsed, nil
}

func parseICalTimezone(eventData *string) time.Location {
	re, _ := regexp.Compile(`X-WR-TIMEZONE:.*?\n`)
	result := re.FindString(*eventData)

	// parse the timezone result to time.Location
	timezone := trimField(result, "X-WR-TIMEZONE:")
	// create location instance
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go Local
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}
	return *loc
}

func parseMain(eventData *string, elementsP *[]Event, href, color string) {
	eventStart, tzId := parseEventStart(eventData)
	eventEnd, tzId := parseEventEnd(eventData)
	eventDuration := parseEventDuration(eventData)
	eventUntil, _ := parseEventUntil(eventData)
	freq := parseEventFreq(eventData)

	if eventEnd.Before(eventStart) {
		eventEnd = eventStart.Add(eventDuration)
	}

	start, _ := time.Parse(IcsFormat, startDate)
	end, _ := time.Parse(IcsFormat, endDate)
	//fmt.Println(eventStart)

	var years, days, months int
	switch freq {
	case "DAILY":
		days = 1
		months = 0
		years = 0
		break
	case "WEEKLY":
		days = 7
		months = 0
		years = 0
		break
	case "MONTHLY":
		days = 0
		months = 1
		years = 0
		break
	case "YEARLY":
		days = 0
		months = 0
		years = 1
		break
	}
	//fmt.Println(eventStart)

	for {
		if inTimeSpan(start, end, eventStart) {
			data := Event{
				Href:        href,
				Color:       color,
				Start:       eventStart,
				End:         eventEnd,
				TZID:        tzId,
				Summary:     parseEventSummary(eventData),
				Description: parseEventDescription(eventData),
				Location:    parseEventLocation(eventData),
				Attendees:   parseEventAttendees(eventData),
			}
			*elementsP = append(*elementsP, data)

		}

		if freq == "" {
			break
		}

		eventStart = eventStart.AddDate(years, months, days)
		eventEnd = eventEnd.AddDate(years, months, days)

		if eventStart.After(end) || (!eventUntil.IsZero() && eventStart.After(eventUntil)) {
			break
		}
	}
}
