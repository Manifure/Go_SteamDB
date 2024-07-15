package

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	apiKey = "YOUR_STEAM_API_KEY"
	apiURL = "https://api.steampowered.com/ISteamApps/GetAppList/v2/"
)

type AppListResponse struct {
	Applist struct {
		Apps []struct {
			AppID int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"apps"`
	} `json:"applist"`
}

func getAppListV2() AppListResponse {
	resp, err := http.Get(apiURL)
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
	return appList
}

func searchGameFromAppList(appList AppListResponse, searchTitle string) {
	for _, app := range appList.Applist.Apps {
		if strings.Contains(strings.ToLower(app.Name), strings.ToLower(searchTitle)) {
			fmt.Printf("Found game: %s (AppID: %d)\n", app.Name, app.AppID)
		}
	}
}

func main() {
	appList := getAppListV2()
	searchGameFromAppList(appList, "Dota")
}
