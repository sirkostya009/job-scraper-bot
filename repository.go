package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

var db *pgxpool.Pool

type Subscription struct {
	Url         string   `json:"url"`
	Subscribers []int64  `json:"subscribers"`
	Data        []string `json:"data"`
}

func init() {
	var err error
	db, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
}

func findByUrl(url string) (sub Subscription, _ error) {
	return sub, db.QueryRow(context.Background(), `select * from subscription where url = $1`, url).Scan(&sub.Url, &sub.Data, &sub.Subscribers)
}

func addSubscription(sub Subscription) (pgconn.CommandTag, error) {
	return db.Exec(context.Background(), `insert into subscription values ($1, $2, $3)`, sub.Url, sub.Data, sub.Subscribers)
}

func updateSubscription(sub Subscription) (pgconn.CommandTag, error) {
	return db.Exec(context.Background(), `update subscription set data = $1, subscribers = $2 where url = $3`, sub.Data, sub.Subscribers, sub.Url)
}

func deleteSubscription(url string) (pgconn.CommandTag, error) {
	return db.Exec(context.Background(), `delete from subscription where url = $1`, url)
}

func listSubscriptions(id int64) (subs []Subscription) {
	rows, _ := db.Query(context.Background(), `select * from subscription where $1 = any(subscribers)`, id)
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(&sub.Url, &sub.Data, &sub.Subscribers); err != nil {
			continue
		}
		subs = append(subs, sub)
	}
	rows.Close()
	return
}

func countSubscriptions(id int64) (count int, _ error) {
	return count, db.QueryRow(context.Background(), `select count(*) from subscription where $1 = any(subscribers)`, id).Scan(&count)
}

func deleteSubscriptionsByChatId(id int64) error {
	_, err := db.Exec(context.Background(), `
		update subscription
		set subscribers = array_remove(subscribers, $1)
		where $1 = any(subscribers)
	`, id)
	if err != nil {
		return err
	}
	_, err = db.Exec(context.Background(), `
		delete from subscription
		where subscribers = '{}'
	`)
	return err
}
