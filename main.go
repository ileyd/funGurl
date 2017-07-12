/*
funGurl URL shortener

To the extent possible under law, the author has waived all copyright and related or neighboring rights to funGurl.

https://github.cmo/ileyd/funGirl
*/

package main

import (
	"database/sql"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// string to hold DSN
var databaseParameters string

// template file
var templates = template.Must(template.ParseFiles("index.html"))

// generate identifier to uniquely identify the URL
func generateIdentifier() string {
	identifier := uniuri.NewLen(3)

	// generate new identifiers until one without an existing entry is found
	for getURL(identifier) != "" {
		identifier = uniuri.NewLen(3)
	}

	return identifier
}

// generate a unique identifier, and then make a new DB entry for the long+short URL pair
func allocateURL(longURL string) string {
	identifier := generateIdentifier()

	database, _ := sql.Open("sqlite3", databaseParameters)
	defer database.Close()

	query, _ := database.Prepare("INSERT INTO funGurl VALUES(?, ?);")
	query.Exec(identifier, longURL)

	return identifier
}

// accepts short URL identifier and returns long URL
func getURL(identifier string) string {
	var longURL string

	database, _ := sql.Open("sqlite3", databaseParameters)
	defer database.Close()
	database.QueryRow("SELECT longURL FROM funGurl WHERE id=?;", html.EscapeString(identifier)).Scan(&longURL)

	return longURL
}

// handles index page
func index(writer http.ResponseWriter, request *http.Request) {
	err := templates.Execute(writer, "index.html")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// handles /shorten page, creating new short URL and returning results
//TODO: make this return a proper page
func shorten(writer http.ResponseWriter, request *http.Request) {
	longURL := request.FormValue("longURL")
	identifier := allocateURL(longURL)
	writer.Header().Set("Content-Type", "text/html")
	io.WriteString(writer, "<p><b>http://localhost:9666/s/"+identifier+"</b></p><br><p><b>"+longURL+"</b></p>")
}

// handles redirection of short URL to long URL
func lengthen(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, getURL(mux.Vars(request)["identifier"]), 303)
}

func main() {

	databaseParameters = "./test.db" // temp static testing database

	router := mux.NewRouter()
	router.HandleFunc("/", index)
	router.HandleFunc("/shorten", shorten)
	router.HandleFunc("/s/{identifier}", lengthen)
	router.HandleFunc("/l/{identifier}", lengthen)

	// TODO: make this listen on port specified in configuration rather than static 9666
	err := http.ListenAndServe(":9666", router)

	if err != nil {
		log.Fatal("Fatal error HTTP server:", err)
	}
}
