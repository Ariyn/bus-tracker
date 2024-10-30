package main

import (
	"encoding/json"
	"fmt"
	bus_tracker "github.com/ariyn/bus-tracker"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var expirationDays time.Duration = 90
var jwtSecret []byte
var boltdb *bolt.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET is not set")
	}

	dbPath := os.Getenv("BOLTDB_PATH")
	if len(dbPath) == 0 {
		log.Fatal("BOLTDB_PATH is not set")
	}

	boltdb, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	initDB()
}

func initDB() {
	tx, err := boltdb.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists([]byte("keys"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.CreateBucketIfNotExists([]byte("functions"))
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	sample := e.Group("/sample")
	sample.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	sample.GET("/json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Hello, World!",
			"nested": map[string]any{
				"key":      "value",
				"children": []any{1, 2, 3, "test"},
			},
		})
	})

	auth := e.Group("/auth")
	auth.POST("", requestKey)

	functions := e.Group("/functions")
	functions.Use(middleware.KeyAuth(keyValidator))

	functions.GET("/:name", functionInvoke)
	functions.POST("/:name", functionCreate)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func requestKey(c echo.Context) (err error) {
	var keyRequest struct {
		Username    string `json:"name"`
		Description string `json:"description"`
	}
	err = c.Bind(&keyRequest)
	if err != nil {
		return err
	}

	tx, err := boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	keysBucket := tx.Bucket([]byte("keys"))
	if keysBucket == nil {
		return c.JSON(http.StatusInternalServerError, "keys bucket is not found")
	}

	key := generateKey()
	for keysBucket.Get([]byte(key)) != nil {
		key = generateKey()
	}

	b, err := json.Marshal(keyRequest)
	if err != nil {
		return err
	}

	err = keysBucket.Put([]byte(key), b)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"key": key,
	})
}

func generateKey() string {
	return uuid.New().String()
}

func keyValidator(key string, c echo.Context) (bool, error) {
	tx, err := boltdb.Begin(false)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	keysBucket := tx.Bucket([]byte("keys"))
	if keysBucket == nil {
		return false, nil
	}

	ok := keysBucket.Get([]byte(key)) != nil
	if ok {
		c.Set("key", key)
	}
	return ok, nil
}

type Function struct {
	Name string `json:"name"`
	Code []byte `json:"code"`
	Key  string `json:"key"`
}

func functionInvoke(c echo.Context) (err error) {
	key := c.Get("key").(string)
	name := c.Param("name")

	tx, err := boltdb.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("functions"))
	if bucket == nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("functions bucket is not found: %s", err))
	}

	b := bucket.Get([]byte(name + key))
	if b == nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("function not found: %s", err))
	}

	var f Function

	err = json.Unmarshal(b, &f)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to unmarshal function: %s", err))
	}

	bts, err := bus_tracker.NewBusTrackerScript(string(f.Code))
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to instantiate scripting environment: %s", err))
	}

	v, err := bts.Run()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("failed: %s", err))
	}

	return c.JSON(http.StatusOK, v)
}

func functionCreate(c echo.Context) (err error) {
	key := c.Get("key").(string)
	name := c.Param("name")

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return
	}
	defer c.Request().Body.Close()

	tx, err := boltdb.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	functionsBucket := tx.Bucket([]byte("functions"))
	if functionsBucket == nil {
		return c.JSON(http.StatusInternalServerError, "functions bucket is not found")
	}

	f := Function{
		Name: name,
		Code: body,
		Key:  key,
	}

	b, err := json.Marshal(f)
	if err != nil {
		return
	}

	err = functionsBucket.Put([]byte(name+key), b)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	return c.String(http.StatusOK, name)
}
