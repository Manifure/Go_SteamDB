package main

import (
	"SteamDB/src/SqlFunc"
	"SteamDB/src/SteamAPI"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Search struct {
	SearchKey    string
	NextPage     int
	TotalPages   int
	TotalResults int
	Result       []SteamAPI.SteamApp
}

var tpl = template.Must(template.ParseFiles("html/home.html"))

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	search := &Search{}
	search.SearchKey = searchKey

	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}

	search.NextPage = next

	psqlInfo := SqlFunc.GetPsqlInfo()
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	const query = "SELECT * FROM steam_games WHERE name ILIKE $1 limit 48"
	searchPattern := "%" + search.SearchKey + "%"
	res, err := db.Query(query, searchPattern)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()

	var showApp []SteamAPI.SteamApp

	for res.Next() {

		var app SteamAPI.SteamApp
		err = res.Scan(&app.Id, &app.AppID, &app.Name)
		if err != nil {
			log.Fatal(err)
		}
		showApp = append(showApp, app)
	}

	search.Result = showApp
	search.TotalResults = len(showApp)
	err = tpl.Execute(w, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", mux)
}
