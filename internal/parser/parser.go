package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type SlowQueryEntry struct {
	Time         string
	QueryTime    float64
	LockTime     float64
	RowsSent     int
	RowsExamined int
	Database     string
	SetTimestamp int64
	SQL          string
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
			if err := ExtractValues(block); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			block = nil
		}
		block = append(block, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func ExtractValues(block []string) error {
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
	fmt.Printf("%+v\n", entry)
	return nil
}
