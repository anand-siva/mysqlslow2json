package parser

import "testing"

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
			},
		},
		{
			name: "without database",
			block: []string{
				"# Time: 2026-03-06T08:15:22.987654Z",
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
