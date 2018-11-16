package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/scottfrazer/running/strava"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app    = kingpin.New("running", "Scott's Running Manager")
		login  = app.Command("login", "Strava login")
		list   = app.Command("list", "List all runs")
		load   = app.Command("load", "Load Strava data")
		poster = app.Command("poster", "Create PNG image of runs within a timeframe")
		stats  = app.Command("stats", "Stats")
	)

	googleMapsKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	stravaClientId := os.Getenv("STRAVA_CLIENT_ID")
	stravaSecretKey := os.Getenv("STRAVA_SECRET_KEY")
	headless := false

	store, err := strava.NewPostgresDataStore("dbname=website sslmode=disable")
	check(err)

	var client *strava.StravaClient
	if session, err := store.GetSession(); err != nil {
		log.Fatalf("error loading session: %v", err)
	} else if session == nil {
		if headless {
			log.Printf("session is expired (headless mode)")
			os.Exit(1)
		} else {
			client, err = strava.NewStravaClientFromBrowserBasedLogin(stravaClientId, stravaSecretKey, store)
			check(err)
		}
	} else {
		client, err = strava.NewStravaClientFromSession(*session)
		check(err)
	}

	homeDir, err := os.UserHomeDir()
	check(err)

	ctx := context.Background()

	/////
	activitiesBytes, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".gorun", "activities.json"))
	check(err)
	activities := []strava.SummaryActivity{}
	check(json.Unmarshal(activitiesBytes, &activities))
	check(store.Save(activities))

	lapsDir := filepath.Join(os.Getenv("HOME"), ".gorun", "laps")

	err = filepath.Walk(lapsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", path)
		activityId, err := strconv.ParseInt(strings.TrimSuffix(filepath.Base(path), ".json"), 10, 64)
		if err != nil {
			fmt.Printf("... skip\n")
			return nil
		}

		if !info.IsDir() {
			// Read and parse JSON file
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var laps []strava.ActivityLap
			if err := json.Unmarshal(data, &laps); err != nil {
				return err
			}

			check(store.SaveLaps(activityId, laps))
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
	/////

	check(client.Sync(ctx, store))
	activities, err = store.Load(strava.ActivityFilter{})
	check(err)
	for _, activity := range activities {
		fmt.Printf("%+v\n", activity)
	}

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch command {
	case login.FullCommand():
	case load.FullCommand():
	case stats.FullCommand():
		type streak struct {
			streakType string // run,rest
			start      *time.Time
			end        *time.Time
			days       int
		}

		activitiesByDate := map[time.Time]*strava.SummaryActivity{}
		start := time.Now()
		end := time.Now()
		for i, activity := range activities {
			if activity.Date().Before(start) {
				start = activity.Date()
			}
			activitiesByDate[activity.Date().Truncate(time.Hour*24)] = &activities[i]
		}

		dateIterator := func(s, e time.Time, f func(d time.Time)) {
			s = s.Truncate(time.Hour * 24)
			e = e.Truncate(time.Hour * 24)
			for s.Before(e) {
				f(s)
				s = s.Add(time.Hour * 24)
			}
			f(s)
		}

		currentStreak := streak{}
		streaksByType := map[string][]streak{}
		dateIterator(start, end, func(date time.Time) {
			fmt.Printf("date: %s\n", date.Format("2006-01-02"))
			activity := activitiesByDate[date]

			streakType := "rest"
			if activity != nil && activity.Type == "Run" {
				streakType = "run"
			}

			resetStreak := false

			if currentStreak.streakType == "" {
				resetStreak = true
			} else if currentStreak.streakType != streakType {
				end := date.Add(-time.Hour * 24)
				currentStreak.end = &end
				streaksByType[currentStreak.streakType] = append(streaksByType[currentStreak.streakType], currentStreak)
				resetStreak = true
			} else {
				currentStreak.days += 1
			}

			if resetStreak {
				currentStreak.start = &date
				currentStreak.end = nil
				currentStreak.days = 1
				currentStreak.streakType = streakType
			}
		})

		sort.SliceStable(streaksByType["run"], func(i, j int) bool {
			return streaksByType["run"][i].days < streaksByType["run"][j].days
		})

		sort.SliceStable(streaksByType["rest"], func(i, j int) bool {
			return streaksByType["rest"][i].days < streaksByType["rest"][j].days
		})

		for _, streak := range streaksByType["run"] {
			if streak.days < 20 {
				continue
			}

			fmt.Printf(
				"run streak: %s to %s (%d days)\n",
				streak.start.Format("2006-01-02"),
				streak.end.Format("2006-01-02"),
				streak.days,
			)
		}

		for _, streak := range streaksByType["rest"] {
			if streak.days < 20 {
				continue
			}

			fmt.Printf(
				"rest streak: %s to %s (%d days)\n",
				streak.start.Format("2006-01-02"),
				streak.end.Format("2006-01-02"),
				streak.days,
			)
		}

	case list.FullCommand():
		for _, a := range activities {
			if a.WorkoutType == 1 || true {
				fmt.Printf("%s : %s, %s, type=%d -- %s\n", a.Date().Format("01/02/2006 15:04:05"), a.DistanceString(), a.MovingTimeString(), a.WorkoutType, a.Name)
			}
		}
	case poster.FullCommand():
		var activities2 []strava.SummaryActivity

		mapID := "9dc6d0f8e5ac205b" // retro

		for _, activity := range activities {
			if !strings.HasPrefix(activity.DateString, "2022") {
				continue
			}
			if len(activity.Map.Polyline) == 0 {
				continue
			}
			activities2 = append(activities2, activity)
			mapsDir := path.Join(homeDir, ".gorun", "maps")
			os.MkdirAll(mapsDir, 0755)
			mapPath := path.Join(mapsDir, fmt.Sprintf("%d_%s.png", activity.Id, mapID))
			if _, err := os.Stat(mapPath); os.IsNotExist(err) {
				fmt.Printf("Generating image %s for activity: %s\n", mapPath, activity.Name)
				pngBytes, err := APIGoogleStaticMaps(googleMapsKey, activity.Map.Polyline, mapID)
				check(err)
				err = os.WriteFile(mapPath, pngBytes, 0644)
				check(err)
			}
		}

		mapPath := func(activityId int64) string {
			mapsDir := path.Join(homeDir, ".gorun", "maps")
			return path.Join(mapsDir, fmt.Sprintf("%d_%s.png", activityId, mapID))
		}

		sort.Sort(strava.SummaryActivityDateSort(activities2))

		poster := NewTilePoster(activities2, 18.0/24.0, 1280, 1280, 3, 50, mapPath)
		os.WriteFile("output.png", poster.Generate(), 0755)
		fmt.Println("================================")
	}
}
