package converter

import (
	"fmt"
	"strings"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertDowntime converts a DataDog downtime to a Dynatrace maintenance window.
func ConvertDowntime(dd *datadog.Downtime) (*dynatrace.MaintenanceWindow, error) {
	if dd.Disabled {
		return nil, fmt.Errorf("downtime is disabled, skipping")
	}

	mw := &dynatrace.MaintenanceWindow{
		Name:        fmt.Sprintf("Migrated: %s", dd.Message),
		Description: dd.Message,
		Type:        "PLANNED",
		Suppression: "DETECT_PROBLEMS_DONT_ALERT",
	}

	// Convert schedule
	tz := dd.Timezone
	if tz == "" {
		tz = "UTC"
	}

	startTime := time.Unix(dd.Start, 0)
	var endTime time.Time
	if dd.End != nil {
		endTime = time.Unix(*dd.End, 0)
	} else {
		endTime = startTime.Add(24 * time.Hour) // Default to 24h
	}

	mw.Schedule = dynatrace.MaintenanceSchedule{
		Start:  startTime.Format("2006-01-02 15:04"),
		End:    endTime.Format("2006-01-02 15:04"),
		ZoneID: tz,
	}

	if dd.Recurrence != nil {
		mw.Schedule.RecurrenceType = mapRecurrenceType(dd.Recurrence.Type)
		duration := int(endTime.Sub(startTime).Minutes())
		mw.Schedule.Recurrence = &dynatrace.MaintenanceRecurrence{
			DurationMinutes: duration,
			StartTime:       startTime.Format("15:04"),
		}
		if dd.Recurrence.Type == "weeks" && len(dd.Recurrence.WeekDays) > 0 {
			mw.Schedule.Recurrence.DayOfWeek = strings.ToUpper(dd.Recurrence.WeekDays[0])
		}
	} else {
		mw.Schedule.RecurrenceType = "ONCE"
	}

	// Convert scope to maintenance scope
	if len(dd.Scope) > 0 || len(dd.MonitorTags) > 0 {
		scope := &dynatrace.MaintenanceScope{}
		for _, tag := range dd.MonitorTags {
			parts := strings.SplitN(tag, ":", 2)
			t := dynatrace.METag{Key: parts[0]}
			if len(parts) == 2 {
				t.Value = parts[1]
			}
			scope.Matches = append(scope.Matches, dynatrace.MaintenanceScopeMatch{
				Type:           "TAG",
				Tags:           []dynatrace.METag{t},
				TagCombination: "OR",
			})
		}
		mw.Scope = scope
	}

	return mw, nil
}

func mapRecurrenceType(ddType string) string {
	switch ddType {
	case "days":
		return "DAILY"
	case "weeks":
		return "WEEKLY"
	case "months":
		return "MONTHLY"
	default:
		return "ONCE"
	}
}
