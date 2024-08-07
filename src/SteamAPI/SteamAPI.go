package SteamAPI

import (
	"SteamDB/src/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const apiGetAppURL = "https://api.steampowered.com/IStoreService/GetAppList/v1/?key=%s&last_appid=%d&max_results=50000"

var apiConfig config.Config

func SetAPIConfig(cfg config.Config) {
	apiConfig = cfg
}

type SteamApp struct {
	AppID               uint64
	Id                  uint64
	Name                string
	last_modified       uint64
	price_change_number uint64
}

type AppListResponse struct {
	Response struct {
		Apps              []SteamApp
		Have_more_results bool
		Last_appid        uint64
	}
}

func GetAppListV2(last_appid uint64) (AppListResponse, error) {
	resp, err := http.Get(fmt.Sprintf(apiGetAppURL, apiConfig.Key, last_appid))
	if err != nil {
		fmt.Println("Error fetching app list:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	var appList AppListResponse
	err = json.Unmarshal(body, &appList)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
	}
	return appList, err
}

func SearchGameFromAppList(appList AppListResponse, searchTitle string) []SteamApp {
	var matchedApps []SteamApp
	if searchTitle == "" {
		return matchedApps
	}
	for _, app := range appList.Response.Apps {
		if strings.Contains(strings.ToLower(app.Name), strings.ToLower(searchTitle)) {
			//fmt.Printf("Found game: %s (AppID: %d)\n", app.Name, app.AppID)
			matchedApps = append(matchedApps, app)
		}
	}
	return matchedApps
}
