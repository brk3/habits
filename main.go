package main

import (
	"log"

	"brk3.github.io/cmd"
	bolt "go.etcd.io/bbolt"
)

func initBolt() *bolt.DB {
	db, err := bolt.Open("habits.db", 0600, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	return db
}

func main() {
	db := initBolt()
	defer db.Close()

	cmd.InitDB(db)
	cmd.Execute()
}
