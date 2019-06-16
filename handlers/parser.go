package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// MK matakuliah
type MK struct {
	Kode   string `json:"kode"`
	Nama   string `json:"nama"`
	Jadwal string `json:"jadwal"`
	Ruang  string `json:"ruang"`
	Dosen  string `json:"dosen"`
	Sks    int    `json:"sks"`
}

// Stat is return status
type Stat struct {
	Status string `json:"status"`
}

func parse(str string) []MK {
	var Matkul []MK
	tokenizer := html.NewTokenizer(strings.NewReader(str))
	counter := 1
	var newMK MK
	for tokenType := tokenizer.Next(); tokenType != html.ErrorToken; {
		if tokenType == html.StartTagToken {
			token := tokenizer.Token()

			if token.Data == "b" {
				tokenizer.Next()
				token = tokenizer.Token()
				if len(token.Data) > 3 {
					newMK.Nama = token.Data
				}
			}

			if token.Data == "td" {

				tokenType = tokenizer.Next()
				if tokenType == html.TextToken {
					token = tokenizer.Token()
					data := token.Data
					// fmt.Println(data)
					switch counter {
					case 2:
						newMK.Kode = data
						break
					case 4:
						newMK.Sks, _ = strconv.Atoi(data)
						break
					case 6:
						newMK.Dosen = data
						break
					case 7:
						tmp := strings.Split(data, "Ruang ")
						newMK.Jadwal = tmp[0]
						newMK.Ruang = tmp[1]
						Matkul = append(Matkul, newMK)
						counter = 0
						break
					}
				}
				counter++
			}
		}
		tokenType = tokenizer.Next()
	}

	return Matkul
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	str := r.FormValue("html")

	sessions, err := store.Get(r, "parser-data")
	if err != nil {
		json.NewEncoder(w).Encode(&Stat{Status: err.Error()})
		return
	}
	tmp := parse(str)
	sessions.Values["matkul"] = tmp
	sessions.Values["recurrence"] = r.FormValue("recurrence")
	err = sessions.Save(r, w)

	if len(tmp) > 0 && err == nil {
		http.Redirect(w, r, "/auth/google/login", http.StatusFound)
		return
	}
	json.NewEncoder(w).Encode(&Stat{Status: err.Error()})
}
