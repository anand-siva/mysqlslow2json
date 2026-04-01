package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
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
	User         string  `json:"user"`
	Host         string  `json:"host"`
	ThreadID     int     `json:"thread_id"`
}

// ParseSlowLog reads a MySQL slow query log and writes JSON lines to outputPath.
func ParseSlowLog(path string, outputPath string, follow bool) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("slow query log does not exist: %s", path)
		}
		return err
	}
	defer file.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	reader := bufio.NewReader(file)
	var block []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if follow {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				if len(block) > 0 && strings.HasPrefix(block[0], "# Time:") {
					entry := ExtractValues(block)
					if err := OutputJSON(outputFile, entry); err != nil {
						return err
					}
				}

				return nil
			}
			return err
		}

		line = strings.TrimRight(line, "\r\n")

		if isStartupHeaderLine(line) {
			if len(block) > 0 && strings.HasPrefix(block[0], "# Time:") {
				entry := ExtractValues(block)
				if err := OutputJSON(outputFile, entry); err != nil {
					return err
				}
			}
			block = nil
			continue
		}

		if strings.HasPrefix(line, "# Time:") && len(block) > 0 {
			if strings.HasPrefix(block[0], "# Time:") {
				entry := ExtractValues(block)
				if err := OutputJSON(outputFile, entry); err != nil {
					return err
				}
			}
			block = nil
		}

		block = append(block, line)
	}
}

func OutputJSON(writer io.Writer, entry SlowQueryEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer, string(data)); err != nil {
		return err
	}

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
		if strings.HasPrefix(line, "# User@Host:") {
			rest := strings.TrimPrefix(line, "# User@Host: ")

			if parts := strings.Split(rest, "  Id:"); len(parts) == 2 {
				userHostPart := strings.TrimSpace(parts[0])
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &entry.ThreadID)

				if parts := strings.Split(userHostPart, " @ "); len(parts) == 2 {
					left := strings.TrimSpace(parts[0])
					right := strings.TrimSpace(parts[1])

					if i := strings.Index(left, "["); i != -1 {
						entry.User = left[:i]
					} else {
						entry.User = left
					}

					if i := strings.LastIndex(right, "["); i != -1 {
						host := strings.TrimSpace(right[:i])
						if host == "" {
							host = strings.Trim(right[i:], "[]")
						}
						entry.Host = host
					} else {
						entry.Host = right
					}
				}
			}
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

func isStartupHeaderLine(line string) bool {
	return strings.HasPrefix(line, "/usr/sbin/mysqld, Version:") ||
		strings.HasPrefix(line, "Tcp port:") ||
		line == "Time                 Id Command    Argument"
}
