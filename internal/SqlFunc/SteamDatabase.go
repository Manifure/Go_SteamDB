package SqlFunc

import (
	"SteamDB/configs"
	"SteamDB/internal/SteamAPI"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
)

var dbConfig configs.Config

func SetDBConfig(cfg configs.Config) {
	dbConfig = cfg
}

// Вспомогательная функция для одной иттерации получения и записи игр в sql таблицу
// Это нужно, потому что SteamAPI позволяет вывести максимум 50k значений из базы данных
func getAppToSql(data SteamAPI.AppListResponse, lastAppid uint64, stmt *sql.Stmt) SteamAPI.AppListResponse {
	var err error
	data, err = SteamAPI.GetAppListV2(lastAppid)
	if err != nil {
		log.Fatal(err)
	}

	for _, value := range data.Response.Apps {
		_, err = stmt.Exec(value.AppID, value.Name)
		if err != nil {
			log.Fatal(err)
		}
	}
	return data
}

// Функция для заполнения sql таблицы списком всех игр из базы данных данных Steam
func FillDbFromSteam() {
	psqlInfo := GetPsqlInfo()
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("steam_games", "appid", "name"))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var data SteamAPI.AppListResponse
	data = getAppToSql(data, data.Response.Last_appid, stmt)

	for data.Response.Have_more_results {
		data.Response.Have_more_results = false
		data = getAppToSql(data, data.Response.Last_appid, stmt)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}
	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
}

func GetPsqlInfo() string {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Database)
	return psqlInfo
}

func GetDBConnection() (*sql.DB, error) {
	psqlInfo := GetPsqlInfo()
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return db, nil
}
