package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jfcote87/google-api-go-client/batch"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/calendar.events"},
	Endpoint:     google.Endpoint,
}

func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)
	u := googleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

// Basically just random string generator
// plus setting a cookie to client
func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	oauthstate, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthstate.Value {
		log.Println("Invalid oauth google state!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := addCalendarEvents(r.FormValue("code"), r)

	if err != nil {
		log.Println(err.Error())
		json.NewEncoder(w).Encode(&Stat{Status: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(&Stat{Status: "success"})
}

func findClosestDateOfDay(date string) time.Time {
	var days map[string]int
	days = map[string]int{}
	days["Senin,"] = 1
	days["Selasa,"] = 2
	days["Rabu,"] = 3
	days["Kamis,"] = 4
	days["Jumat,"] = 5
	days["Sabtu,"] = 6
	days["Minggu,"] = 0

	day := days[date]

	dayNow := int(time.Now().Weekday())
	diff := day - dayNow
	var diffAbs int
	var addDate int

	if diff < 0 {
		diffAbs *= -1
	} else {
		diffAbs = diff
	}

	if diffAbs < 5 {
		addDate = diff
	} else {
		addDate = 7 - diffAbs
	}

	return time.Now().AddDate(0, 0, addDate)
}

func addCalendarEvents(authCode string, r *http.Request) error {
	tok, err := googleOauthConfig.Exchange(context.TODO(), authCode)

	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	client := googleOauthConfig.Client(context.Background(), tok)
	bsv := batch.Service{Client: client}
	calsv, err := calendar.New(batch.BatchClient)

	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	calendarID := "primary"
	var days map[string]string
	days = map[string]string{}
	days["Senin,"] = "MO"
	days["Selasa,"] = "TU"
	days["Rabu,"] = "WE"
	days["Kamis,"] = "TH"
	days["Jumat,"] = "FR"
	days["Sabtu,"] = "SA"
	days["Minggu,"] = "SU"

	sessions, err := store.Get(r, "parser-data")
	if err != nil {
		return err
	}

	data := sessions.Values["matkul"]

	Matkuls, ok := data.(*[]MK)

	if !ok {
		return errors.New("Failed to retrieve matkul")
	}

	recurrences, ok1 := sessions.Values["recurrence"].(string)
	if !ok1 {
		return errors.New("Failed to retrieve recurrence")
	}

	for _, ev := range *Matkuls {
		// create Event
		startEnd := strings.Split(ev.Jadwal, " ")
		day := startEnd[0]
		startEnd = strings.Split(startEnd[1], "-")
		start := startEnd[0]
		end := startEnd[1]
		date := findClosestDateOfDay(day)
		dateEv := strings.Split(date.String(), " ")[0]

		event := &calendar.Event{
			Summary:     ev.Nama + " (" + ev.Kode + ")",
			Description: ev.Dosen,
			Location:    ev.Ruang,
			Start:       &calendar.EventDateTime{DateTime: dateEv + "T" + start + ":00+07:00", TimeZone: "Asia/Jakarta"},
			End:         &calendar.EventDateTime{DateTime: dateEv + "T" + end + ":00+07:00", TimeZone: "Asia/Jakarta"},
			Reminders:   &calendar.EventReminders{UseDefault: true},
			Recurrence:  []string{"RRULE:FREQ=WEEKLY;COUNT=" + recurrences + ";BYDAY=" + days[day]},
		}

		event, err := calsv.Events.Insert(calendarID, event).Do()

		// queue new request in batch service
		err = bsv.AddRequest(err,
			batch.SetResult(&event),
			batch.SetTag(ev),
		)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	responses, err := bsv.DoCtx(context.Background())
	if err != nil {
		log.Println(err)
		return err
	}

	for _, r := range responses {
		if r.Err != nil {
			log.Println(r.Err)
			err = r.Err
			continue
		}
	}

	return err
}
