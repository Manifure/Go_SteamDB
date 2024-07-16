package SteamAPI

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	apiGetAppURL = "https://api.steampowered.com/IStoreService/GetAppList/v1/?key=%s&max_results=50000"
	key          = "BBF805EFC23EADDF7002101399628FCD"
)

type SteamApp struct {
	AppID               uint64
	Name                string
	last_modified       uint64
	price_change_number uint64
}

type AppListResponse struct {
	Response struct {
		Apps              []SteamApp
		have_more_results bool
		last_appid        uint64
	}
}

func GetAppListV2() (AppListResponse, error) {
	resp, err := http.Get(fmt.Sprintf(apiGetAppURL, key))
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
	for _, app := range appList.Response.Apps {
		if strings.Contains(strings.ToLower(app.Name), strings.ToLower(searchTitle)) {
			//fmt.Printf("Found game: %s (AppID: %d)\n", app.Name, app.AppID)
			matchedApps = append(matchedApps, app)
		}
	}
	return matchedApps
}
