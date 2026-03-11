package parser

import (
	"fmt"
	"os"
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

	fmt.Println(file.Name())
	return nil
}
