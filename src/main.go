package main

import (
	"SteamDB/src/SteamAPI"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Result     []SteamAPI.SteamApp
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

	resp, err := SteamAPI.GetAppListV2()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	matched := SteamAPI.SearchGameFromAppList(resp, search.SearchKey)
	search.TotalPages = len(matched)

	//for _, game := range matched {
	//	fmt.Printf("Found game: %s (AppID: %d)\n", game.Name, game.AppID)
	//}

	search.Result = matched

	err = tpl.Execute(w, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//fmt.Println("Search Query is: ", searchKey)
	//fmt.Println("Results page is: ", page)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
	//w.Header().Set("Content-Type", "text/html")
	//db, err := SteamAPI.GetAppList()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("Get App List")
	//for _, v := range db[36:46] {
	//	fmt.Fprintf(w, "<a href=\"https://store.steampowered.com/app/%d\" target=\"_blank\"><img src=\"https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg\" alt=\"%s\"></a><br/>\n", v.AppId, v.AppId, v.Name)
	//}
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", mux)
}
