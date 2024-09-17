package main

import (
	"SteamDB/configs"
	"SteamDB/internal/HtmlFunc"
	"SteamDB/internal/SqlFunc"
	"SteamDB/internal/SteamAPI"
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
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

func handleError(c *gin.Context, err error, statusCode int) {
	log.Println(err)
	c.String(statusCode, err.Error())
}

func parseParams(c *gin.Context) (string, int, error) {
	searchKey := c.Query("q")
	page := c.DefaultQuery("page", "1")

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

func searchHandler(c *gin.Context) {
	searchKey, currentPage, err := parseParams(c)
	if err != nil {
		handleError(c, err, http.StatusBadRequest)
		return
	}
	const itemsPerPage = 48
	offset := (currentPage - 1) * itemsPerPage

	db, err := SqlFunc.GetDBConnection()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	totalResults, err := fetchTotalResults(c.Request.Context(), db, searchKey)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	results, err := fetchResults(c.Request.Context(), db, searchKey, itemsPerPage, offset)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
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

	c.HTML(http.StatusOK, "search", search)
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "home", nil)
}

func main() {
	cfg := configs.MustLoad()

	dbConfig := configs.Config{
		SQLConfig: configs.SQLConfig{
			Host:     cfg.Host,
			Port:     cfg.Port,
			User:     cfg.User,
			Password: cfg.Password,
			Database: cfg.Database,
		},
	}

	apiConfig := configs.Config{
		SteamAPIConfig: configs.SteamAPIConfig{
			Key: cfg.Key,
		},
	}

	SqlFunc.SetDBConfig(dbConfig)
	SteamAPI.SetAPIConfig(apiConfig)

	r := gin.Default()

	r.SetFuncMap(HtmlFunc.FuncMap)
	r.LoadHTMLGlob("html/*")

	r.Static("/static", "static")

	r.GET("/search", searchHandler)
	r.GET("/", homePage)

	err := r.Run(cfg.Address)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
