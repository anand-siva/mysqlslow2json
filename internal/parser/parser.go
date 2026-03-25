package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type SlowQueryEntry struct {
	Time         string  `json:"time"`
	QueryTime    float64 `json:"query_time"`
	LockTime     float64 `json:"lock_time"`
	RowsSent     int     `json:"rows_sent"`
	RowsExamined int     `json:"rows_examined"`
	Database     string  `json:"database"`
	SetTimestamp int64   `json:"set_timestamp"`
	SQL          string  `json:"sql"`
}

// ParseSlowLog is a placeholder parser that reports the target log path.
func ParseSlowLog(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("slow query log does not exist: %s", path)
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var block []string

	for scanner.Scan() {
		line := scanner.Text()
		// This line is the block reset, shows that we are about to start a new block
		if strings.HasPrefix(line, "# Time:") && len(block) > 0 {
			if strings.HasPrefix(block[0], "# Time:") {
				slowQueryEntryStruct := ExtractValues(block)
				if err := OutputJson(slowQueryEntryStruct); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			block = nil
		}
		block = append(block, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(block) > 0 && strings.HasPrefix(block[0], "# Time:") {
		slowQueryEntryStruct := ExtractValues(block)
		if err := OutputJson(slowQueryEntryStruct); err != nil {
			return err
		}
	}

	return nil
}

func OutputJson(entry SlowQueryEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func ExtractValues(block []string) SlowQueryEntry {
	entry := &SlowQueryEntry{}
	for _, line := range block {
		if strings.HasPrefix(line, "# Time:") {
			entry.Time = strings.TrimPrefix(line, "# Time: ")
		}
		if strings.HasPrefix(line, "# Query_time:") {
			fmt.Sscanf(line,
				"# Query_time: %f Lock_time: %f Rows_sent: %d Rows_examined: %d",
				&entry.QueryTime,
				&entry.LockTime,
				&entry.RowsSent,
				&entry.RowsExamined,
			)
		}

		if strings.HasPrefix(line, "use ") {
			entry.Database = strings.TrimSuffix(strings.TrimPrefix(line, "use "), ";")
		}

		if strings.HasPrefix(line, "SET timestamp=") {
			fmt.Sscanf(line, "SET timestamp=%d;", &entry.SetTimestamp)
		}

		// Everything else is probably SQL
		if !strings.HasPrefix(line, "#") &&
			!strings.HasPrefix(line, "SET ") &&
			!strings.HasPrefix(line, "use ") &&
			strings.TrimSpace(line) != "" {

			entry.SQL += line + " "
		}

	}
	return *entry
}
