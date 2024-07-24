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
	Result       []SteamAPI.SteamApp
	NextPage     int
	TotalPages   int
	TotalResults int
	CurrentPage  int
	HasNextPage  bool
	HasPrevPage  bool
}

var funcMap = template.FuncMap{
	"sub": func(a, b int) int {
		return a - b
	},
	"add": func(a, b int) int {
		return a + b
	},
}

var tpl = template.Must(template.New("home.html").Funcs(funcMap).ParseFiles("html/home.html"))

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

	currentPage, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page number", http.StatusInternalServerError)
		return
	}

	itemsPerPage := 48
	offset := (currentPage - 1) * itemsPerPage

	psqlInfo := SqlFunc.GetPsqlInfo()
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	countQuery := "SELECT COUNT(*) FROM steam_games WHERE name ILIKE $1"
	searchPattern := "%" + search.SearchKey + "%"
	var count int
	err = db.QueryRow(countQuery, searchPattern).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	query := "SELECT * FROM steam_games WHERE name ILIKE $1 LIMIT $2 OFFSET $3"
	res, err := db.Query(query, searchPattern, itemsPerPage, offset)
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
	search.TotalResults = count
	search.CurrentPage = currentPage
	search.TotalPages = (count + itemsPerPage - 1) / itemsPerPage
	search.HasNextPage = currentPage < search.TotalPages
	search.HasPrevPage = currentPage > 1

	err = tpl.Execute(w, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseIntOrDefault(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, nil
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
