package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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
			fmt.Println("BLOCK:")
			fmt.Println("--------:")
			fmt.Println(strings.Join(block, "\n"))

			// this is where I send the block for json parsing

			block = nil
		}

		block = append(block, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
