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
	functionID string
	taskID     string
	code       string
}

func main() {
	defer db.Close()

	wg := sync.WaitGroup{}

	queue := make(chan function, 100)

	go func(queue chan<- function) {
		wg.Add(1)
		defer wg.Done()
		for {
			rows, err := db.Query("SELECT id, function_id from tasks WHERE done_at IS NULL AND status = 'pending' LIMIT 1")
			if err != nil {
				log.Println(err)
				continue
			}

			for rows.Next() {
				var functionID, taskID string
				err = rows.Scan(&taskID, &functionID)
				if err != nil {
					log.Println(err)
					continue
				}

				_, err = db.Exec("UPDATE tasks SET status = 'running', started_at = NOW() WHERE id = $1", taskID)
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
					functionID: functionID,
					taskID:     taskID,
					code:       code,
				}
			}
		}
	}(queue)

	for f := range queue {
		runScript(f.taskID, f.code)
	}
}

func runScript(id string, code string) {
	defer func() {
		_, err := db.Exec("UPDATE tasks SET done_at = NOW(), status='done' WHERE id = $1", id)

		if err != nil {
			log.Println(err)
		}
	}()

	bts, err := bus_tracker.NewBusTrackerScript(code)
	if err != nil {
		writeResult(id, "", err)
		return
	}

	v, err := bts.Run()
	if err != nil {
		writeResult(id, "", err)
		return
	}

	b, err := json.Marshal(v)
	if err != nil {
		writeResult(id, "", err)
		return
	}

	writeResult(id, string(b), nil)
	return
}
func writeResult(id string, result string, err error) {
	log.Println(id, result, err)
	errorString := ""
	if err != nil {
		errorString = err.Error()
	}

	_, insertErr := db.Exec("UPDATE tasks SET result = $1, error = $2 WHERE id = $3", result, errorString, id)
	if insertErr != nil {
		log.Println(insertErr)
	}
}
