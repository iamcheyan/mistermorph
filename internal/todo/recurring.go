package todo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	everyNDaysPattern  = regexp.MustCompile(`^every\s+([1-9][0-9]*)\s+days?$`)
	everyNHoursPattern = regexp.MustCompile(`^every\s+([1-9][0-9]*)\s+hours?$`)
)

type recurringInterval struct {
	Days  int
	Hours int
}

type RecurringMaterializeResult struct {
	Generated int `json:"generated"`
	Advanced  int `json:"advanced"`
}

func (s *Store) AddRecurringWithChatID(raw string, nextAt string, repeat string, tz string, chatID string) (RecurringUpdateResult, error) {
	recur, _, err := s.readRECUR(s.nowUTC())
	if err != nil {
		return RecurringUpdateResult{}, err
	}
	entry, err := ParseRecurringEntryFromInput(raw, nextAt, repeat, tz)
	if err != nil {
		return RecurringUpdateResult{}, err
	}
	if parsedChatID := normalizeEntryChatID(chatID); parsedChatID != "" {
		entry.ChatID = parsedChatID
	}
	if err := validateRecurringEntry(entry); err != nil {
		return RecurringUpdateResult{}, err
	}
	recur.Entries = append([]RecurringEntry{entry}, recur.Entries...)
	if err := s.writeRECUR(recur); err != nil {
		return RecurringUpdateResult{}, err
	}
	return RecurringUpdateResult{
		OK:             true,
		Action:         "add_recurring",
		RecurringCount: len(recur.Entries),
		Entry:          &entry,
	}, nil
}

func ParseRecurringEntryFromInput(raw string, nextAt string, repeat string, tz string) (RecurringEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RecurringEntry{}, fmt.Errorf("content is required")
	}
	if item, ok := parseRecurringEntryLine(raw); ok {
		return item, nil
	}
	nextAt = strings.TrimSpace(nextAt)
	repeat = normalizeRecurringRepeat(repeat)
	tz = normalizeRecurringTZ(tz)
	entry := RecurringEntry{
		NextAt:  nextAt,
		Repeat:  repeat,
		TZ:      tz,
		Content: raw,
	}
	if err := validateRecurringEntry(entry); err != nil {
		return RecurringEntry{}, err
	}
	return entry, nil
}

func (s *Store) MaterializeDueRecurring() (RecurringMaterializeResult, error) {
	if s == nil {
		return RecurringMaterializeResult{}, fmt.Errorf("todo store is nil")
	}
	now := s.nowUTC()
	recur, exists, err := s.readRECUR(now)
	if err != nil {
		return RecurringMaterializeResult{}, err
	}
	if !exists || len(recur.Entries) == 0 {
		return RecurringMaterializeResult{}, nil
	}

	wip, err := s.readWIP(now)
	if err != nil {
		return RecurringMaterializeResult{}, err
	}

	generated := make([]Entry, 0, len(recur.Entries))
	for idx, item := range recur.Entries {
		if err := validateRecurringEntry(item); err != nil {
			return RecurringMaterializeResult{}, err
		}
		loc, err := recurringLocation(item.TZ)
		if err != nil {
			return RecurringMaterializeResult{}, err
		}
		nextAt, err := parseRecurringTime(item.NextAt, loc)
		if err != nil {
			return RecurringMaterializeResult{}, err
		}
		if nextAt.After(now) {
			continue
		}
		generated = append(generated, Entry{
			Done:      false,
			CreatedAt: now.Format(TimestampLayout),
			ChatID:    normalizeEntryChatID(item.ChatID),
			Content:   materializedRecurringContent(nextAt, loc, item.TZ, item.Content),
		})
		advanced, err := nextRecurringTimeAfterInLocation(nextAt, item.Repeat, now, loc)
		if err != nil {
			return RecurringMaterializeResult{}, err
		}
		recur.Entries[idx].NextAt = advanced.Format(TimestampLayout)
		recur.Entries[idx].Repeat = normalizeRecurringRepeat(item.Repeat)
		recur.Entries[idx].TZ = normalizeRecurringTZ(item.TZ)
	}
	if len(generated) == 0 {
		return RecurringMaterializeResult{}, nil
	}

	wip.Entries = append(generated, wip.Entries...)
	if err := validateWIPEntries(wip.Entries); err != nil {
		return RecurringMaterializeResult{}, err
	}
	if err := s.writeWIP(wip); err != nil {
		return RecurringMaterializeResult{}, err
	}
	if err := s.writeRECUR(recur); err != nil {
		return RecurringMaterializeResult{}, err
	}
	return RecurringMaterializeResult{
		Generated: len(generated),
		Advanced:  len(generated),
	}, nil
}

func validateRecurringEntry(item RecurringEntry) error {
	loc, err := recurringLocation(item.TZ)
	if err != nil {
		return err
	}
	if _, err := parseRecurringTime(item.NextAt, loc); err != nil {
		return err
	}
	if !validRecurringRepeat(item.Repeat) {
		return fmt.Errorf("invalid Repeat: %s", strings.TrimSpace(item.Repeat))
	}
	if err := validateEntryChatID(item.ChatID); err != nil {
		return err
	}
	if err := validateEntryReferences(item.Content); err != nil {
		return err
	}
	if strings.TrimSpace(item.Content) == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

func materializedRecurringContent(nextAt time.Time, loc *time.Location, tz string, content string) string {
	if loc == nil {
		loc = time.UTC
	}
	prefix := nextAt.In(loc).Format(TimestampLayout)
	tz = normalizeRecurringTZ(tz)
	if tz != "" {
		prefix += " (" + tz + ")"
	}
	return strings.TrimSpace(prefix + " " + strings.TrimSpace(content))
}

func parseRecurringTime(raw string, loc *time.Location) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("Next is required")
	}
	if loc == nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation(TimestampLayout, value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid Next: %s", value)
	}
	return t, nil
}

func nextRecurringTimeAfter(from time.Time, repeat string, after time.Time) (time.Time, error) {
	return nextRecurringTimeAfterInLocation(from, repeat, after, time.UTC)
}

func nextRecurringTimeAfterInLocation(from time.Time, repeat string, after time.Time, loc *time.Location) (time.Time, error) {
	repeat = normalizeRecurringRepeat(repeat)
	interval, err := recurringRepeatInterval(repeat)
	if err != nil {
		return time.Time{}, err
	}
	if loc == nil {
		loc = time.UTC
	}
	next := from.In(loc)
	after = after.In(loc)
	if next.After(after) {
		return next, nil
	}

	if interval.Days > 0 {
		elapsedDays := int(after.Sub(next).Hours() / 24)
		steps := elapsedDays/interval.Days + 1
		next = next.AddDate(0, 0, steps*interval.Days)
		for !next.After(after) {
			next = next.AddDate(0, 0, interval.Days)
		}
		return next, nil
	}

	step := time.Duration(interval.Hours) * time.Hour
	elapsedHours := int(after.Sub(next).Hours())
	steps := elapsedHours/interval.Hours + 1
	next = next.Add(time.Duration(steps) * step)
	for !next.After(after) {
		next = next.Add(step)
	}
	return next, nil
}

func recurringRepeatInterval(repeat string) (recurringInterval, error) {
	repeat = normalizeRecurringRepeat(repeat)
	switch {
	case repeat == "daily":
		return recurringInterval{Days: 1}, nil
	case repeat == "weekly":
		return recurringInterval{Days: 7}, nil
	default:
		if days, ok := parseEveryNDays(repeat); ok {
			return recurringInterval{Days: days}, nil
		}
		if hours, ok := parseEveryNHours(repeat); ok {
			return recurringInterval{Hours: hours}, nil
		}
		return recurringInterval{}, fmt.Errorf("invalid Repeat: %s", strings.TrimSpace(repeat))
	}
}

func validRecurringRepeat(raw string) bool {
	_, err := recurringRepeatInterval(raw)
	return err == nil
}

func normalizeRecurringRepeat(raw string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(raw))), " ")
}

func parseEveryNDays(raw string) (int, bool) {
	matches := everyNDaysPattern.FindStringSubmatch(normalizeRecurringRepeat(raw))
	if len(matches) != 2 {
		return 0, false
	}
	days, err := strconv.Atoi(matches[1])
	if err != nil || days <= 0 {
		return 0, false
	}
	return days, true
}

func parseEveryNHours(raw string) (int, bool) {
	matches := everyNHoursPattern.FindStringSubmatch(normalizeRecurringRepeat(raw))
	if len(matches) != 2 {
		return 0, false
	}
	hours, err := strconv.Atoi(matches[1])
	if err != nil || hours <= 0 {
		return 0, false
	}
	return hours, true
}

func recurringLocation(tz string) (*time.Location, error) {
	tz = normalizeRecurringTZ(tz)
	if tz == "" {
		return time.Local, nil
	}
	if strings.EqualFold(tz, "UTC") {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid TZ: %s", tz)
	}
	return loc, nil
}

func validRecurringTZ(raw string) bool {
	_, err := recurringLocation(raw)
	return err == nil
}

func normalizeRecurringTZ(raw string) string {
	return strings.TrimSpace(raw)
}
