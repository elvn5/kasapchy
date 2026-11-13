// Operator-only moderation tool. Run inside the API container; never expose it as HTTP.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"time"
)

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, os.Getenv("MARKET_DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("usage: moderate list | hide-product ID | show-product ID | suspend-seller UUID | restore-seller UUID | resolve-report ID")
	}
	if args[0] == "list" {
		rows, err := db.Query(ctx, `SELECT r.id,r.product_id,p.seller_id::text,r.reason,r.created_at FROM market_reports r JOIN market_products p ON p.id=r.product_id WHERE r.resolved_at IS NULL ORDER BY r.created_at LIMIT 100`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, product int64
			var seller, reason string
			var created time.Time
			if err := rows.Scan(&id, &product, &seller, &reason, &created); err != nil {
				return err
			}
			if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"reportId": id, "productId": product, "sellerId": seller, "reason": reason, "created": created}); err != nil {
				return err
			}
		}
		return rows.Err()
	}
	if len(args) != 2 {
		return fmt.Errorf("command requires one ID")
	}
	var query string
	switch args[0] {
	case "hide-product":
		query = "UPDATE market_products SET hidden=true WHERE id::text=$1"
	case "show-product":
		query = "UPDATE market_products SET hidden=false WHERE id::text=$1"
	case "suspend-seller":
		query = "UPDATE market_users SET suspended=true WHERE id::text=$1 AND role='Seller'"
	case "restore-seller":
		query = "UPDATE market_users SET suspended=false WHERE id::text=$1 AND role='Seller'"
	case "resolve-report":
		query = "UPDATE market_reports SET resolved_at=now() WHERE id::text=$1"
	default:
		return fmt.Errorf("unknown command")
	}
	tag, err := db.Exec(ctx, query, args[1])
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("record not found")
	}
	fmt.Println("Updated", args[0], args[1])
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
