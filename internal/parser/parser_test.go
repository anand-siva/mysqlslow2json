package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractValues(t *testing.T) {
	tests := []struct {
		name  string
		block []string
		want  SlowQueryEntry
	}{
		{
			name: "with database",
			block: []string{
				"# Time: 2026-03-06T08:10:00.123456Z",
				"# User@Host: app_user[app_user] @  [192.168.1.10]  Id:    101",
				"# Query_time: 2.500200  Lock_time: 0.000050 Rows_sent: 50  Rows_examined: 1250000",
				"use ecom_db;",
				"SET timestamp=1772784600;",
				"SELECT c.name, o.order_date, o.total FROM customers c JOIN orders o ON c.id = o.customer_id WHERE o.status = 'PENDING' ORDER BY o.order_date DESC;",
			},
			want: SlowQueryEntry{
				Time:         "2026-03-06T08:10:00.123456Z",
				QueryTime:    2.5002,
				LockTime:     0.00005,
				RowsSent:     50,
				RowsExamined: 1250000,
				Database:     "ecom_db",
				SetTimestamp: 1772784600,
				SQL:          "SELECT c.name, o.order_date, o.total FROM customers c JOIN orders o ON c.id = o.customer_id WHERE o.status = 'PENDING' ORDER BY o.order_date DESC; ",
				User:         "app_user",
				Host:         "192.168.1.10",
				ThreadID:     101,
			},
		},
		{
			name: "without database",
			block: []string{
				"# Time: 2026-03-06T08:15:22.987654Z",
				"# User@Host: admin_user[admin] @ localhost [127.0.0.1]  Id:    105",
				"# Query_time: 15.234110  Lock_time: 0.012300 Rows_sent: 0  Rows_examined: 500000",
				"SET timestamp=1772784922;",
				"UPDATE products SET stock = stock - 1 WHERE category_id IN (SELECT id FROM categories WHERE name LIKE '%Electronics%') AND status = 'active';",
			},
			want: SlowQueryEntry{
				Time:         "2026-03-06T08:15:22.987654Z",
				QueryTime:    15.23411,
				LockTime:     0.0123,
				RowsSent:     0,
				RowsExamined: 500000,
				Database:     "",
				SetTimestamp: 1772784922,
				SQL:          "UPDATE products SET stock = stock - 1 WHERE category_id IN (SELECT id FROM categories WHERE name LIKE '%Electronics%') AND status = 'active'; ",
				User:         "admin_user",
				Host:         "localhost",
				ThreadID:     105,
			},
		},
		{
			name: "user host line with empty hostname and only ip",
			block: []string{
				"# Time: 2026-03-06T09:18:33.999000Z",
				"# User@Host: app_user[app_user] @  [192.168.1.10]  Id:    101",
				"# Query_time: 1.950000  Lock_time: 0.000040 Rows_sent: 250  Rows_examined: 800000",
				"use ecom_db;",
				"SET timestamp=1772788713;",
				"SELECT 1;",
			},
			want: SlowQueryEntry{
				Time:         "2026-03-06T09:18:33.999000Z",
				QueryTime:    1.95,
				LockTime:     0.00004,
				RowsSent:     250,
				RowsExamined: 800000,
				Database:     "ecom_db",
				SetTimestamp: 1772788713,
				SQL:          "SELECT 1; ",
				User:         "app_user",
				Host:         "192.168.1.10",
				ThreadID:     101,
			},
		},
		{
			name: "user host line with hostname and ip",
			block: []string{
				"# Time: 2026-03-06T09:45:01.121212Z",
				"# User@Host: reporting[reporting] @ localhost [127.0.0.1]  Id:    302",
				"# Query_time: 12.005000  Lock_time: 1.500000 Rows_sent: 0  Rows_examined: 15000",
				"SET timestamp=1772790301;",
				"SELECT 2;",
			},
			want: SlowQueryEntry{
				Time:         "2026-03-06T09:45:01.121212Z",
				QueryTime:    12.005,
				LockTime:     1.5,
				RowsSent:     0,
				RowsExamined: 15000,
				Database:     "",
				SetTimestamp: 1772790301,
				SQL:          "SELECT 2; ",
				User:         "reporting",
				Host:         "localhost",
				ThreadID:     302,
			},
		},
		{
			name: "user host line with extra spaces before id",
			block: []string{
				"# Time: 2026-03-06T10:30:50.565656Z",
				"# User@Host: bi_tool[bi_tool] @  [10.0.1.100]    Id:    550",
				"# Query_time: 65.432000  Lock_time: 0.000000 Rows_sent: 5000  Rows_examined: 25000000",
				"SET timestamp=1772793050;",
				"SELECT 3;",
			},
			want: SlowQueryEntry{
				Time:         "2026-03-06T10:30:50.565656Z",
				QueryTime:    65.432,
				LockTime:     0,
				RowsSent:     5000,
				RowsExamined: 25000000,
				Database:     "",
				SetTimestamp: 1772793050,
				SQL:          "SELECT 3; ",
				User:         "bi_tool",
				Host:         "10.0.1.100",
				ThreadID:     550,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractValues(tt.block)

			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseSlowLog_WritesJSONL(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "slow.log")
	outputPath := filepath.Join(tempDir, "slow-query.jsonl")

	logContents := strings.Join([]string{
		"/usr/sbin/mysqld, Version: 8.0.35 (MySQL Community Server - GPL). started with:",
		"Tcp port: 3306  Unix socket: /var/run/mysqld/mysqld.sock",
		"Time                 Id Command    Argument",
		"# Time: 2026-03-06T08:10:00.123456Z",
		"# User@Host: app_user[app_user] @  [192.168.1.10]  Id:    101",
		"# Query_time: 2.500200  Lock_time: 0.000050 Rows_sent: 50  Rows_examined: 1250000",
		"use ecom_db;",
		"SET timestamp=1772784600;",
		"SELECT 1;",
		"# Time: 2026-03-06T08:15:22.987654Z",
		"# User@Host: admin_user[admin] @ localhost [127.0.0.1]  Id:    105",
		"# Query_time: 15.234110  Lock_time: 0.012300 Rows_sent: 0  Rows_examined: 500000",
		"SET timestamp=1772784922;",
		"UPDATE products SET stock = stock - 1;",
	}, "\n") + "\n"

	if err := os.WriteFile(logPath, []byte(logContents), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := ParseSlowLog(logPath, outputPath, false); err != nil {
		t.Fatalf("ParseSlowLog: %v", err)
	}

	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(outputBytes)), "\n")
	if len(lines) != 2 {
		t.Fatalf("line count: got %d, want %d\noutput:\n%s", len(lines), 2, string(outputBytes))
	}

	if !strings.Contains(lines[0], `"user":"app_user"`) {
		t.Fatalf("first line missing expected user: %s", lines[0])
	}

	if !strings.Contains(lines[1], `"thread_id":105`) {
		t.Fatalf("second line missing expected thread id: %s", lines[1])
	}
}

func TestParseSlowLog_IgnoresStartupHeaderNoise(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "slow.log")
	outputPath := filepath.Join(tempDir, "slow-query.jsonl")

	logContents := strings.Join([]string{
		"# Time: 2026-04-01T23:17:31.046828Z",
		"# User@Host: appuser[appuser] @  [192.168.65.1]  Id:    12",
		"# Query_time: 2.000897  Lock_time: 0.000000 Rows_sent: 1  Rows_examined: 1",
		"SET timestamp=1775085449;",
		"select sleep(2);",
		"/usr/sbin/mysqld, Version: 8.0.45 (MySQL Community Server - GPL). started with:",
		"Tcp port: 0  Unix socket: /var/run/mysqld/mysqld.sock",
		"Time                 Id Command    Argument",
		"/usr/sbin/mysqld, Version: 8.0.45 (MySQL Community Server - GPL). started with:",
		"Tcp port: 3306  Unix socket: /var/run/mysqld/mysqld.sock",
		"Time                 Id Command    Argument",
		"# Time: 2026-04-01T23:18:00.000000Z",
		"# User@Host: appuser[appuser] @  [192.168.65.1]  Id:    13",
		"# Query_time: 2.100000  Lock_time: 0.000000 Rows_sent: 1  Rows_examined: 1",
		"SET timestamp=1775085480;",
		"select sleep(2);",
	}, "\n") + "\n"

	if err := os.WriteFile(logPath, []byte(logContents), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := ParseSlowLog(logPath, outputPath, false); err != nil {
		t.Fatalf("ParseSlowLog: %v", err)
	}

	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	output := string(outputBytes)
	if strings.Contains(output, "/usr/sbin/mysqld, Version:") {
		t.Fatalf("startup header leaked into output:\n%s", output)
	}

	if strings.Contains(output, "Tcp port:") {
		t.Fatalf("startup tcp header leaked into output:\n%s", output)
	}

	if strings.Count(strings.TrimSpace(output), "\n")+1 != 2 {
		t.Fatalf("expected two JSON lines, got output:\n%s", output)
	}
}
