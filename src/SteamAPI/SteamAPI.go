package SteamAPI

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const apiURL = "https://api.steampowered.com/ISteamApps/GetAppList/v2/"

type SteamApp struct {
	AppID uint64
	Name  string
}

type AppListResponse struct {
	AppList struct {
		Apps []SteamApp
	}
}

func GetAppListV2() (AppListResponse, error) {
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
	return appList, err
}

func SearchGameFromAppList(appList AppListResponse, searchTitle string) []SteamApp {
	var matchedApps []SteamApp
	for _, app := range appList.AppList.Apps {
		if strings.Contains(strings.ToLower(app.Name), strings.ToLower(searchTitle)) {
			//fmt.Printf("Found game: %s (AppID: %d)\n", app.Name, app.AppID)
			matchedApps = append(matchedApps, app)
		}
	}
	return matchedApps
}

//const BaseSteamAPIURLProduction = "https://api.steampowered.com"
//
//var BaseSteamAPIURL = BaseSteamAPIURLProduction
//
//type SteamMethod string
//
//func GetAppList() ([]SteamApp, error) {
//	getAppList := NewSteamMethod("ISteamApps", "GetAppList", 2)
//	var resp appListJson
//	err := getAppList.Request(nil, &resp)
//	if err != nil {
//		return nil, err
//	}
//	return resp.AppList.Apps, nil
//}
//
//func NewSteamMethod(interf, method string, version int) SteamMethod {
//	m := fmt.Sprintf("%v/%v/%v/v%v/", BaseSteamAPIURL, interf, method, strconv.Itoa(version))
//	return SteamMethod(m)
//}
//func (s SteamMethod) Request(data url.Values, v interface{}) error {
//	url := string(s)
//	if data != nil {
//		url += "?" + data.Encode()
//	}
//	resp, err := http.Get(url)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != 200 {
//		return fmt.Errorf("steamapi %s Status code %d", s, resp.StatusCode)
//	}
//
//	d := json.NewDecoder(resp.Body)
//
//	return d.Decode(&v)
