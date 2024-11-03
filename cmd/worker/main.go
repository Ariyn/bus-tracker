package main

import (
	"database/sql"
	"encoding/json"
	bus_tracker "github.com/ariyn/bus-tracker"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"sync"
)

var db *sql.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
}

func getCode(functionId string) (code string, err error) {
	row := db.QueryRow("SELECT code FROM functions WHERE id = $1", functionId)
	err = row.Scan(&code)
	return
}

type function struct {
	id   string
	code string
}

func main() {
	defer db.Close()

	wg := sync.WaitGroup{}

	queue := make(chan function, 100)

	go func(queue chan<- function) {
		wg.Add(1)
		defer wg.Done()
		for {
			rows, err := db.Query("SELECT function_id from tasks WHERE done_at IS NULL AND status = 'pending' LIMIT 1")
			if err != nil {
				log.Println(err)
				continue
			}

			for rows.Next() {
				var functionID string
				err = rows.Scan(&functionID)
				if err != nil {
					log.Println(err)
					continue
				}

				_, err = db.Exec("UPDATE tasks SET status = 'running' WHERE function_id = $1", functionID)
				if err != nil {
					log.Println(err)
					continue
				}

				code, err := getCode(functionID)
				if err != nil {
					log.Println(err)
					continue
				}
				queue <- function{
					id:   functionID,
					code: code,
				}
			}
		}
	}(queue)

	for f := range queue {
		id, code := f.id, f.code
		bts, err := bus_tracker.NewBusTrackerScript(code)
		if err != nil {
			log.Println(err)
			continue
		}

		v, err := bts.Run()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println(v)

		result, err := json.Marshal(v)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = db.Exec("UPDATE tasks SET done_at = NOW(), result=$2, status='done' WHERE function_id = $1", id, string(result))
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
