package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type postUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type user struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin" gorm:"default:false"`
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"hello": "world", "foo": "bar"})
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUrl, dbUrlErr := os.LookupEnv("DB_URL")

	if dbUrlErr == false {
		log.Fatal("DB_URL must be defined in the environment variables")
	}

	db, dbErr := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	sqlDb, err := db.DB()
	sqlDb.SetMaxIdleConns(10)

	if dbErr != nil {
		log.Fatal("Could not connect to database", dbErr)
	}

	router := chi.NewRouter()
	router.Get("/json", jsonHandler)
	router.Get("/profile/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Id must be an integer")
		}
		var found user
		db.First(&found, id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(found)
	})

	router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var data user

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		result := db.Create(&data)
		if result.RowsAffected == 0 {
			http.Error(w, result.Error.Error(), http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	router.Get("/users/{page}", func(w http.ResponseWriter, r *http.Request) {
		page := chi.URLParam(r, "page")
		pageParsed, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, "Page must be an integer", http.StatusBadRequest)
		}

		var users []user

		db.Limit(10).Offset((pageParsed - 1) * 10).Find(&users)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	port, exists := os.LookupEnv("PORT")
	if exists {
		fmt.Printf("Listening at port %s\n", port)
		err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
		fmt.Println(err.Error())
	} else {
		fmt.Println("Listening at default port: 5000")
		err := http.ListenAndServe(":5000", router)
		fmt.Println(err.Error())
	}
}
