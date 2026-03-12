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
	fmt.Println("BLOCK:")
	fmt.Println("--------")
	for line_number, line := range block {
		fmt.Printf("%d: %s\n", line_number, line)
	}
	return nil
}
