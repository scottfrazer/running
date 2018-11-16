package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type SummaryActivity struct {
	Id          int64       `json:"id"`
	Name        string      `json:"name"`
	DateString  string      `json:"start_date_local"`
	Distance    float64     `json:"distance"`
	MovingTime  float64     `json:"moving_time"`
	WorkoutType int         `json:"workout_type"`
	Map         ActivityMap `json:"map"`
}

type ActivityMap struct {
	Id            string `json:"id"`
	ResourceState int    `json:"resource_state"`
	Polyline      string `json:"summary_polyline"`
}

func (a *SummaryActivity) Date() time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z", a.DateString)
	check(err)
	return t
}

func (a *SummaryActivity) Miles() float64 {
	return (a.Distance / 1000) * 0.621371
}

func (a *SummaryActivity) DistanceString() string {
	return fmt.Sprintf("%s mi", strconv.FormatFloat(a.Miles(), 'f', 2, 64))
}

func (a *SummaryActivity) MovingTimeString() string {
	d := time.Duration(a.MovingTime) * time.Second
	return d.String()
}

func (a *SummaryActivity) Pace() string {
	d := time.Duration(a.MovingTime/a.Miles()) * time.Second
	_ = d
	return ""
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadActivities() []SummaryActivity {
	client := &http.Client{}
	page := 0
	var activities []SummaryActivity

	for {
		page++

		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?page=%d&per_page=200", page), nil)
		req.Header.Add("Authorization", "Bearer 6d5a66010bd4762fb5e165796ba007209a0d4f9b")
		resp, err := client.Do(req)
		check(err)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		var pageActivities []SummaryActivity
		err = json.Unmarshal(body, &pageActivities)
		check(err)

		if len(pageActivities) == 0 {
			break
		}

		activities = append(activities, pageActivities...)
	}
	return activities
}

func GoogleMapsPolylineURL(polyline string) *url.URL {
	u, err := url.Parse("https://maps.googleapis.com/maps/api/staticmap")
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("size", "640x640")
	q.Set("scale", "2")
	q.Set("maptype", "terrain")
	q.Set("path", "weight:3|color:red|enc:"+polyline)
	q.Set("key", "AIzaSyAeDQVk6UHY1oabDEsSETDQgtXRzDzFW_E")
	u.RawQuery = q.Encode()
	return u
}

func ResizePNG(pngPath string, width, height int) []byte {
	out, err := exec.Command("convert", "-resize", fmt.Sprintf("%dx%d", width, height), pngPath, "-").Output()
	check(err)
	return out
}

func main() {
	var (
		app       = kingpin.New("running", "Running Manager")
		list      = app.Command("list", "List transactions")
		importCmd = app.Command("import", "Import transactions from raw CSV exports")
	)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch command {
	case importCmd.FullCommand():
		activities := loadActivities()
		var activities2018 []SummaryActivity

		for _, activity := range activities {
			mapPath := fmt.Sprintf("maps/%d.png", activity.Id)
			if _, err := os.Stat(mapPath); os.IsNotExist(err) && len(activity.Map.Polyline) > 0 {
				fmt.Printf("Generating image %s for activity: %s\n", mapPath, activity.Name)
				mapUrl := GoogleMapsPolylineURL(activity.Map.Polyline).String()

				client := &http.Client{}
				req, err := http.NewRequest("GET", mapUrl, nil)
				resp, err := client.Do(req)
				check(err)

				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				err = ioutil.WriteFile(mapPath, body, 0644)
				check(err)
			}
			if activity.Date().Format("2006") == "2018" {
				activities2018 = append(activities2018, activity)
				fmt.Printf("Id=%d, Date=%s\n", activity.Id, activity.Date())
			}
		}

		width := 250
		cols := 10
		rows := 30
		col := 0
		row := 0
		rgba := image.NewRGBA(image.Rect(0, 0, cols*width, rows*width))

		for _, activity := range activities2018 {
			if len(activity.Map.Polyline) == 0 {
				continue
			}

			imageFilename := fmt.Sprintf("maps/%d.png", activity.Id)
			imageResizedFilename := fmt.Sprintf("maps/%d.resize.png", activity.Id)
			fmt.Printf("Processing %s...\n", imageFilename)

			var resizedImageFile []byte
			if _, err := os.Stat(imageResizedFilename); os.IsNotExist(err) {
				resizedImageFile = ResizePNG(imageFilename, width, width)
				err = ioutil.WriteFile(imageResizedFilename, resizedImageFile, 0644)
				check(err)
			} else {
				resizedImageFile, err = ioutil.ReadFile(imageResizedFilename)
				check(err)
			}

			activityMap, _, err := image.Decode(bytes.NewReader(resizedImageFile))
			check(err)

			draw.Draw(rgba, image.Rect(col*width, row*width, col*width+width, row*width+width), activityMap, image.Point{0, 0}, draw.Src)

			if col == cols-1 {
				col = 0
				row += 1
			} else {
				col += 1
			}
		}

		out, err := os.Create("output.png")
		check(err)
		png.Encode(out, rgba)
		out.Close()

		/*
			imgFile1, err := os.Open("maps/522205094.png")
			check(err)
			imgFile2, err := os.Open("maps/527507219.png")
			check(err)
			img1, _, err := image.Decode(imgFile1)
			check(err)
			img2, _, err := image.Decode(imgFile2)
			check(err)

			rgba := image.NewRGBA(image.Rect(0, 0, 1280*2, 1280))
			draw.Draw(rgba, image.Rect(0, 0, 1280, 1280), img1, image.Point{0, 0}, draw.Src)
			draw.Draw(rgba, image.Rect(1280, 0, 1280*2, 1280), img2, image.Point{0, 0}, draw.Src)
			out, err := os.Create("/Users/scottfrazer/output.png")
			check(err)
			png.Encode(out, rgba)
			out.Close()

			resized := ResizePNG("/Users/scottfrazer/output.png", 1000, 500)
			err = ioutil.WriteFile("/Users/scottfrazer/resized.png", resized, 0644)
			check(err)
		*/

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Date", "Distance", "Time", "Pace", "Type", "Polyline"})
		for _, a := range activities {
			table.Append([]string{a.Name, a.Date().Format("01/02/2006 15:04:05"), a.DistanceString(), a.MovingTimeString(), a.Pace(), strconv.Itoa(a.WorkoutType), strconv.Itoa(len(a.Map.Polyline))})
		}
		//table.Render()

	case list.FullCommand():
		fmt.Println("list")
	}
}
