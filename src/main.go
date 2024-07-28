package main

import (
	"SteamDB/src/SqlFunc"
	"SteamDB/src/SteamAPI"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
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

func parseParams(r *http.Request) (string, int, error) {
	params := r.URL.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	currentPage, err := strconv.Atoi(page)
	if err != nil {
		return "", 0, fmt.Errorf("invalid page number: %w", err)
	}

	return searchKey, currentPage, nil
}

func fetchTotalResults(ctx context.Context, db *sql.DB, searchKey string) (int, error) {
	countQuery := "SELECT COUNT(*) FROM steam_games WHERE name ILIKE $1"
	searchPattern := "%" + searchKey + "%"
	var count int
	err := db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total results: %w", err)
	}
	return count, nil
}

func fetchResults(ctx context.Context, db *sql.DB, searchKey string, itemsPerPage, offset int) ([]SteamAPI.SteamApp, error) {
	query := "SELECT * FROM steam_games WHERE name ILIKE $1 LIMIT $2 OFFSET $3"
	rows, err := db.QueryContext(ctx, query, "%"+searchKey+"%", itemsPerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}
	defer rows.Close()

	var apps []SteamAPI.SteamApp
	for rows.Next() {
		var app SteamAPI.SteamApp
		err = rows.Scan(&app.Id, &app.AppID, &app.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	searchKey, currentPage, err := parseParams(r)
	const itemsPerPage = 48
	offset := (currentPage - 1) * itemsPerPage

	db, err := SqlFunc.GetDBConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ctx := r.Context()
	totalResults, err := fetchTotalResults(ctx, db, searchKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results, err := fetchResults(ctx, db, searchKey, itemsPerPage, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	search := &Search{
		SearchKey:    searchKey,
		Result:       results,
		TotalResults: totalResults,
		CurrentPage:  currentPage,
		TotalPages:   (totalResults + itemsPerPage - 1) / itemsPerPage,
		HasNextPage:  currentPage < (totalResults+itemsPerPage-1)/itemsPerPage,
		HasPrevPage:  currentPage > 1,
	}

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
