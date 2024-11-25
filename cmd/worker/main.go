package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	bus_tracker "github.com/ariyn/bus-tracker"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
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

	cronTicker := time.NewTicker(1 * time.Minute)
	defer cronTicker.Stop()

	go func() {
		wg.Add(1)
		defer wg.Done()

		for range cronTicker.C {
			err := getCronjob()
			if err != nil {
				log.Println(err)
			}
		}
	}()

	taskTicker := time.NewTicker(1 * time.Second)
	defer taskTicker.Stop()

	go func(queue chan<- *function) {
		wg.Add(1)
		defer wg.Done()
		for range taskTicker.C {
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

	// TODO: This does not run parallel. It should be run in parallel.
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

type Cronjob struct {
	Expression string
	Minutes    []int
	Hours      []int
	DayOfMonth []int
	DaysOfWeek []int
	Month      []int
}

func getCronjob() (err error) {
	rows, err := db.Query("SELECT id, function_id, crontab FROM crontabs WHERE next_run_at < NOW()")
	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var id, functionId, crontabString string
		err = rows.Scan(&id, &functionId, &crontabString)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = db.Exec("INSERT INTO tasks (function_id, status, user_id, cron_id) VALUES ($1, 'pending', (SELECT user_id FROM crontabs WHERE id = $2), $2)", functionId, id)
		if err != nil {
			log.Println(err)
			continue
		}

		var cronjob Cronjob
		err = json.Unmarshal([]byte(crontabString), &cronjob)
		if err != nil {
			log.Println(err)
			continue
		}

		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		result, err := parser.Parse(cronjob.Expression)
		if err != nil {
			log.Println(err)
			continue
		}

		next := result.Next(time.Now())
		_, err = db.Exec("UPDATE crontabs SET next_run_at = $1 WHERE id = $2", next, id)
		if err != nil {
			log.Println(err)
			continue
		}
	}

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
		log.Println("error raised", err)
		writeResult(id, "", err)
		return
	}

	v, err := bts.Run()
	if err != nil {
		log.Println("error returned", err)
		writeResult(id, "", err)
		return
	}

	log.Printf("returned %#v", v)

	v, err = saveAndReplaceImages(v)
	if err != nil {
		writeResult(id, "", err)
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

type btImage struct {
	BtImage struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"_bt_data"`
	OriginalUrl string `json:"original_url"`
}

func saveAndReplaceImages(v interface{}) (replacedV interface{}, err error) {
	if image, ok := v.(*bus_tracker.Image); ok {
		path := uuid.New().String()

		_, err := bus_tracker.StorageClient.UploadFile("images", path, bytes.NewReader(image.Body), storage_go.FileOptions{
			ContentType: &image.ContentType,
		})
		if err != nil {
			return nil, err
		}

		publicUrl := bus_tracker.StorageClient.GetPublicUrl("images", path)
		return btImage{
			BtImage: struct {
				Type string `json:"type"`
				Url  string `json:"url"`
			}{
				Type: "image",
				Url:  publicUrl.SignedURL,
			},
			OriginalUrl: image.Url,
		}, nil

	}

	if arr, ok := v.([]interface{}); ok {
		var replacedArr []interface{}
		for _, item := range arr {
			replacedItem, err := saveAndReplaceImages(item)
			if err != nil {
				return nil, err
			}

			replacedArr = append(replacedArr, replacedItem)
		}

		return replacedArr, nil
	}

	if m, ok := v.(map[string]interface{}); ok {
		replacedM := make(map[string]interface{})
		for k, item := range m {
			replacedItem, err := saveAndReplaceImages(item)
			if err != nil {
				return nil, err
			}

			replacedM[k] = replacedItem
		}

		return replacedM, nil
	}

	return v, nil
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
