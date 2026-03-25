package main

import (
	"flag"
	"fmt"
	"github.com/anand-siva/mysqlslow2json/internal/parser"
	"os"
)

func main() {
	fmt.Println("MySQL Slow Log to json initialized")

	slowQueryLog := flag.String("slow-query-log", "", "Path to MySQL slow query log (required)")
	outputFile := flag.String("output-file", "slow-query.jsonl", "Path to output JSON lines file")

	// Parse the flags
	flag.Parse()

	if *slowQueryLog == "" {
		fmt.Fprintln(os.Stderr, "--slow-query-log is required")
		os.Exit(1)
	}

	if err := parser.ParseSlowLog(*slowQueryLog, *outputFile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
