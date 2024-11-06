package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	bus_tracker "github.com/ariyn/bus-tracker"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	storage_go "github.com/supabase-community/storage-go"
	"log"
	"os"
	"sync"
	"time"
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

	log.Println(os.Getenv("SUPABASE_SERVICE_KEY"))
	bus_tracker.StorageClient = storage_go.NewClient(os.Getenv("SUPABASE_STORAGE_BASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"), nil)
}

type function struct {
	functionID string
	taskID     string
	code       string
	envVar     map[string]string
}

func main() {
	defer db.Close()

	wg := sync.WaitGroup{}

	queue := make(chan *function, 100)

	go func(queue chan<- *function) {
		wg.Add(1)
		defer wg.Done()
		for {
			time.Sleep(1 * time.Second)

			f, err := getTask()
			if err == sql.ErrNoRows {
				continue
			}

			if err != nil {
				log.Println(err)
				continue
			}

			log.Printf("RUN %s for %s", f.taskID, f.functionID)
			queue <- f
		}
	}(queue)

	for f := range queue {
		runScript(f.taskID, f.code, f.envVar)
	}
}

func getTask() (f *function, err error) {
	row := db.QueryRow("SELECT id, function_id from tasks WHERE done_at IS NULL AND status = 'pending'")
	err = row.Err()
	if err != nil {
		return
	}

	var functionID, taskID string
	err = row.Scan(&taskID, &functionID)
	if err != nil {
		return
	}

	envVar, err := getEnvironmentVariables(functionID)
	if err != nil {
		return
	}

	_, err = db.Exec("UPDATE tasks SET status = 'running', started_at = NOW() WHERE id = $1", taskID)
	if err != nil {
		return
	}

	code, err := getCode(functionID)
	if err != nil {
		return
	}

	return &function{
		functionID: functionID,
		taskID:     taskID,
		code:       code,
		envVar:     envVar,
	}, nil

}

func getEnvironmentVariables(functionId string) (map[string]string, error) {
	log.Println(functionId)
	rows, err := db.Query("SELECT key, value FROM environments WHERE function_id = $1", functionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	envVar := make(map[string]string)
	for rows.Next() {
		var key, value string
		err = rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}

		envVar[key] = value
	}

	return envVar, nil
}

func getCode(functionId string) (code string, err error) {
	row := db.QueryRow("SELECT code FROM functions WHERE id = $1", functionId)
	err = row.Err()
	if err != nil {
		return
	}

	err = row.Scan(&code)
	return
}

func runScript(id string, code string, envVar map[string]string) {
	defer func() {
		_, err := db.Exec("UPDATE tasks SET done_at = NOW(), status='done' WHERE id = $1", id)

		if err != nil {
			log.Println(err)
		}
	}()

	bts, err := bus_tracker.NewBusTrackerScript(code, envVar)
	if err != nil {
		writeResult(id, "", err)
		return
	}

	v, err := bts.Run()
	if err != nil {
		writeResult(id, "", err)
		return
	}

	log.Printf("returned %#v, %s", v, err)
	if image, ok := v.(*bus_tracker.Image); ok {
		path := fmt.Sprintf("%s", image.Name)

		_, err = bus_tracker.StorageClient.UploadFile("images", path, bytes.NewReader(image.Body), storage_go.FileOptions{
			ContentType: &image.ContentType,
		})
		if err != nil {
			writeResult(id, "", err)
		}

		publicUrl := bus_tracker.StorageClient.GetPublicUrl("images", path)

		writeResult(id, publicUrl.SignedURL, nil)
		return
	}

	if str, ok := v.(string); ok {
		writeResult(id, str, nil)
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
