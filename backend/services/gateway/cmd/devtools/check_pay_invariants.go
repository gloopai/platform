// check-pay-invariants runs DB invariant checks for payin/payout without the mysql CLI.
//
// Env (same defaults as backend/deploy 脚本):
//
//	PAY_PLATFORM_MYSQL_DSN — full DSN, takes precedence over:
//	MYSQL_HOST (127.0.0.1), MYSQL_USER (root), MYSQL_PASSWORD (your_password), MYSQL_DB (pay)
//	STUCK_PAYOUT_MINUTES — pending payout warning threshold; "0" disables (default 180)
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func runCheckPayInvariants() {
	dsn := dsnFromEnv()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(2)
	}

	stuckMin := 180
	if v := strings.TrimSpace(os.Getenv("STUCK_PAYOUT_MINUTES")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(os.Stderr, "STUCK_PAYOUT_MINUTES must be a non-negative integer\n")
			os.Exit(2)
		}
		stuckMin = n
	}

	failures := 0
	fmt.Println("=== check-pay-invariants (no mysql client) ===")

	q1 := `
SELECT o.order_no, o.merchant_id, o.paid_amount, o.net_amount
FROM payin_orders o
LEFT JOIN fund_logs f ON f.order_no = o.order_no AND f.change_type = 'ORDER_PAID'
WHERE o.status = 1 AND f.id IS NULL`
	n1, err := printRows(db, "[代收] 已支付但无 ORDER_PAID 流水", q1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(2)
	}
	if n1 > 0 {
		failures++
	}

	q2 := `
SELECT o.order_no, o.merchant_id, o.paid_amount, o.net_amount, f.amount AS credited
FROM payin_orders o
JOIN fund_logs f ON f.order_no = o.order_no AND f.change_type = 'ORDER_PAID'
WHERE o.status = 1
  AND f.amount <> IF(o.net_amount > 0, o.net_amount, o.paid_amount)`
	n2, err := printRows(db, "[代收] ORDER_PAID 金额与订单 net/paid 快照不一致", q2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(2)
	}
	if n2 > 0 {
		failures++
	}

	q3 := `
SELECT o.order_no, o.merchant_id, o.status
FROM payin_orders o
JOIN fund_logs f ON f.order_no = o.order_no AND f.change_type = 'ORDER_PAID'
WHERE o.status <> 1`
	n3, err := printRows(db, "[代收] 非已支付状态但存在 ORDER_PAID", q3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(2)
	}
	if n3 > 0 {
		failures++
	}

	q4 := `
SELECT f.order_no, f.merchant_id, f.amount
FROM fund_logs f
LEFT JOIN payin_orders o ON o.order_no = f.order_no
WHERE f.change_type = 'ORDER_PAID'
  AND (o.order_no IS NULL OR o.status <> 1)`
	n4, err := printRows(db, "[代收] ORDER_PAID 孤儿或与订单状态矛盾", q4)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(2)
	}
	if n4 > 0 {
		failures++
	}

	if stuckMin > 0 {
		q5 := fmt.Sprintf(`
SELECT order_no, merchant_id, merchant_order_no, amount, status, updated_at
FROM payout_orders
WHERE status = 0 AND updated_at < (NOW() - INTERVAL %d MINUTE)
ORDER BY updated_at ASC
LIMIT 50`, stuckMin)
		n5, err := printRows(db, fmt.Sprintf("[代付] pending 超过 %d 分钟（提醒，默认不计入失败）", stuckMin), q5)
		if err != nil {
			fmt.Fprintf(os.Stderr, "query: %v\n", err)
			os.Exit(2)
		}
		_ = n5
	} else {
		fmt.Println("[代付] 跳过 pending 超时检查（STUCK_PAYOUT_MINUTES=0）")
	}

	fmt.Println("[代付] DebitPayout 当前不写 fund_logs；代付深对账请配合 test_payout_flow.sh / 管理台")
	fmt.Printf("=== 结束：代收硬性失败项 = %d ===\n", failures)
	if failures > 0 {
		os.Exit(1)
	}
}

func dsnFromEnv() string {
	if d := strings.TrimSpace(os.Getenv("PAY_PLATFORM_MYSQL_DSN")); d != "" {
		return d
	}
	host := getenv("MYSQL_HOST", "127.0.0.1")
	user := getenv("MYSQL_USER", "root")
	pass := getenv("MYSQL_PASSWORD", "your_password")
	dbName := getenv("MYSQL_DB", "pay")
	port := getenv("MYSQL_PORT", "3306")
	// 密码含 @ : / 等时请用 PAY_PLATFORM_MYSQL_DSN（与 gateway Mysql.DataSource 一致）。
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true", user, pass, host, port, dbName)
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

// printRows runs query, prints a header and table; returns row count.
func printRows(db *gorm.DB, title, query string) (int, error) {
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}
	var lines [][]string
	for rows.Next() {
		raw := make([]interface{}, len(cols))
		ptr := make([]interface{}, len(cols))
		for i := range raw {
			ptr[i] = &raw[i]
		}
		if err := rows.Scan(ptr...); err != nil {
			return 0, err
		}
		line := make([]string, len(cols))
		for i, v := range raw {
			if v == nil {
				line[i] = "NULL"
				continue
			}
			switch t := v.(type) {
			case []byte:
				line[i] = string(t)
			default:
				line[i] = fmt.Sprint(t)
			}
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	fmt.Printf("%s: %d 条\n", title, len(lines))
	if len(lines) == 0 {
		return 0, nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(cols, "\t"))
	for _, line := range lines {
		fmt.Fprintln(w, strings.Join(line, "\t"))
	}
	_ = w.Flush()
	return len(lines), nil
}
