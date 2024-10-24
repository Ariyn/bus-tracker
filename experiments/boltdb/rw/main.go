package main

import (
	"github.com/boltdb/bolt"
	"log"
	"os"
)

func main() {
	const testDB = "my.db"

	db, _ := bolt.Open(testDB, 0600, nil)
	defer db.Close()
	defer os.Remove(testDB)

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	bucket, _ := tx.CreateBucketIfNotExists([]byte("mybucket"))
	bucket.Put([]byte("mykey"), []byte("myvalue"))

	tx.Commit()

	tx, _ = db.Begin(false)
	defer tx.Rollback()

	bucket = tx.Bucket([]byte("mybucket"))
	log.Println(string(bucket.Get([]byte("mykey"))))

	// when nil
	log.Println(string(bucket.Get([]byte("not-exists"))))
}
