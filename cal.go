package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetCalendarService() (*calendar.Service, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope, calendar.CalendarScope)
	if err != nil {
		return nil, err
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

type RunDay struct {
	Day         int
	Miles       int
	Description string
}

var pfitz1855 string = `rest,0
Lactate threshold 8mi w/4 mi @15k to HM pace,8
rest,0
General aerobic 9mi,9
rest,0
Recover 4mi,4
Medium-long run 12mi,12

rest,0
General aerobic + speed 8mi w/10x100m strides,8
rest,0
General Aerobic 10mi,10
rest,0
Recovery 5mi,5
Marathon-pace run 13mi w/8mi @ marathon race pace,13

rest,0
General aerobic 10mi,10
Recover 4mi,4
Lactate threshold 8mi w/4mi @15k to half marathon race pace,8
rest,0
Recover 4mi,4
Medium-long run 14mi,14

rest,0
General aerobic + speed 8mi w/10x100m strides,8
Recovery 5mi,5
General aerobic 10mi,10
rest,0
Recover 4mi,4
Medium-long run 15mi,15

rest,0
Lactate threshold 9mi w/5mi @15k to half marathon race pace,9
Recover 5mi,5
General aerobic 10mi,10
rest,0
Recover 5mi,5
Marathon-pace run 16mi w/10mi @ marathon race pace,16

rest,0
General aerobic + speed 8mi w/ 10x100m strides,8
Recover 5mi,5
General aerobic 8mi,8
rest,0
Recovery 4mi,4
Medium-long run 12mi,12

rest,0
Lactate threshold 10mi w/5mi @15k to half marathon pace,10
Recovery 4mi,4
Medium-long run 11mi,11
rest,0
General aerobic + speed 7mi w/ 8x100m strides p.m.,7
Long run 18mi,18

rest,0
Recovery + speed 7mi w/ 6x100m strides,7
Medium-long run 12mi,12
rest,0
Lactate threshold 10mi w/ 6mi @15k to half marathon race pace,10
Recovery 5mi,5
Long run 20mi,20

rest,0
Recovery 6mi,6
Medium-long run 14mi,14
Recovery 6mi,6
rest,0
Recovery + speed 6mi w/ 6x100m strides,6
Marathon-pace run 16mi w/12mi @ marathon race pace,16

rest,0
General aerobic 8mi,8
VO2max 8mi w/ 5x800m @5k race pace; jog 50 to 90% interval time between,8
Recovery 5mi,5
rest,0
General aerobic + speed 8mi w/8x100m strides,8
Medium-long run 14mi,14

rest,0
Recovery + speed 7mi w/6x100m strides,7
Lactate threshold 11mi w/ 7mi @ 15k to half marathon race pace,11
rest,0
Medium-long run 12mi,12
Recovery 5mi,5
Long run 20mi,20

rest,0
VO2max 8mi w/5x600m @ 5k race pace; jog 50 to 90% interval time between,8
Medium-long run 12mi,12
rest,0
Recovery + speed 5mi w/ 6x100m strides,5
8k-15k tune-up race,9
Long run 17mi,17

rest,0
General aerobic 8mi,8
VO2max 9mi w/ 5x1000m @ 5k race pace; jog 50 to 90% interval time between,9
rest,0
Medium-long run 12mi,12
Recovery 5mi,5
Marathon-pace run 18mi w/ 14mi @ marathon race pace,18

rest,0
VO2max 8mi w/5x600m @ 5k race pace; jog 50 to 90% interval time between,8
Medium-long run 11mi,11
rest,0
Recovery + speed 4mi w/ 6x100m strides,4
8k-15k tune-up race,9
Long run 17mi,17

rest,0
Recovery + speed 7mi w/ 6x100m strides,7
VO2max 10mi w/ 4x1200m @ 5k race pace;jog 50 to 90% interval time between,10
rest,0
Medium-long run 11mi,1
Recovery 4mi,4
Long run 20mi,20

rest,0
VO2max 8mi w/ 5x600m @ 5k race pace; jog 50 to 90% interval time between,8
Recovery 6mi,6
rest,0
Recovery + speed 4mi w/ 6x100m strides,4
8k-10k tune-up race,9
Long run 16mi,16

rest,0
General aerobic + speed 7mi w/ 8x100m strides,7
VO2max 8mi w/ 3x1600m @ 5k race pace; jog 50 to 90% interval time between,8
rest,0
Recovery + speed 5mi w/6x100m strides,5
rest,0
Medium-long run 12mi,12

rest,0
Recovery 6mi,6
Dress rehearsal 7mi w/ 2mi @marathon race pace,7
rest,0
Recovery + speed 5mi w/ 6x100m strides,5
Recovery 4mi,4
MARATHON,26`

type Run struct {
	Day         int
	Description string
	Miles       int
}

type TrainingPlan struct {
	Runs []*Run
}

func (tp *TrainingPlan) String(start time.Time) string {
	s := ""
	weeklyMileage := 0
	week := 1
	for _, run := range tp.Runs {
		var day string
		if start.IsZero() {
			day = fmt.Sprintf("Day %d", run.Day)
		} else {
			runDate := start.Add(time.Hour * time.Duration((run.Day-1)*24))
			day = fmt.Sprintf("Day %d (%s)", run.Day, runDate.Format("Mon Jan 2, 2006"))
		}
		s = s + fmt.Sprintf("%s: %s (%dmi)\n", day, run.Description, run.Miles)
		weeklyMileage += run.Miles
		if run.Day%7 == 0 {
			s = s + fmt.Sprintf("Week %d volume: %dmi\n\n", week, weeklyMileage)
			weeklyMileage = 0
			week += 1
		}
	}
	return s
}

func (tp *TrainingPlan) StringRelDate() string {
	return tp.String(time.Time{})
}

func LoadTrainingPlan(name string) *TrainingPlan {
	var runs []*Run
	if name == "pfitz1855" {
		r := csv.NewReader(strings.NewReader(pfitz1855))

		day := 0
		for {
			day += 1
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			check(err)

			description := record[0]
			miles, err := strconv.Atoi(record[1])
			check(err)

			runs = append(runs, &Run{day, description, miles})
		}
		return &TrainingPlan{runs}
	}
	return nil
}

type CalendarTrainingPlan struct {
	Name   string
	Events []*calendar.Event
}

func FindTrainingPlans(srv *calendar.Service, start, end time.Time) ([]*CalendarTrainingPlan, error) {
	events, err := GetRunningEvents(srv, start, end)
	if err != nil {
		return nil, err
	}

	nameToMetadata := make(map[string]*PlanMetadata)
	for _, event := range events {
		var metadata PlanMetadata
		err := json.Unmarshal([]byte(event.Description), &metadata)
		if err != nil {
			return nil, err
		}

		if _, ok := nameToMetadata[metadata.Name]; !ok {
			nameToMetadata[metadata.Name] = &metadata
		}
	}

	var plans []*CalendarTrainingPlan
	for name, metadata := range nameToMetadata {
		planStart, err := time.Parse(time.RFC3339, metadata.Start)
		if err != nil {
			return nil, err
		}

		planEnd, err := time.Parse(time.RFC3339, metadata.End)
		if err != nil {
			return nil, err
		}

		events, err := GetRunningEvents(srv, planStart, planEnd)
		if err != nil {
			return nil, err
		}

		plans = append(plans, &CalendarTrainingPlan{name, events})
	}
	return plans, nil
}

func DeleteTrainingPlan(srv *calendar.Service, plan *CalendarTrainingPlan) error {
	return DeleteRunningEvents(srv, plan.Events)
}

type PlanMetadata struct {
	Name  string `json:"name"`
	Start string `json:"start"` // RFC3339
	End   string `json:"end"`   // RFC3339
}

func AddTrainingPlan(srv *calendar.Service, name string, start time.Time, plan *TrainingPlan) (*CalendarTrainingPlan, error) {
	var events []*calendar.Event

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	start, err = time.ParseInLocation("Mon Jan 2 2006 15:04:05 MST", fmt.Sprintf("%s 07:00:00 EDT", start.Format("Mon Jan 2 2006")), location)
	end := start
	for _, run := range plan.Runs {
		if run.Day < 0 {
			return nil, errors.New("days cannot be negative")
		}
		runDate := start.Add(time.Duration(run.Day) * 24 * time.Hour)
		if runDate.After(end) {
			end = runDate
		}
	}

	metadata, err := json.Marshal(PlanMetadata{name, start.Format(time.RFC3339), end.Format(time.RFC3339)})
	if err != nil {
		return nil, err
	}

	fmt.Printf("metadata=%s\n", metadata)
	for _, run := range plan.Runs {
		randUuid, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		date := start.Add(time.Duration(run.Day) * 24 * time.Hour)
		eventId := "running" + strings.Replace(randUuid.String(), "-", "", -1)
		event := &calendar.Event{
			Id:          eventId,
			Start:       &calendar.EventDateTime{DateTime: date.Format(time.RFC3339)},
			End:         &calendar.EventDateTime{DateTime: date.Format(time.RFC3339)},
			Summary:     run.Description,
			Description: string(metadata),
		}

		events = append(events, event)
	}

	err = InsertRunningEvents(srv, events)
	if err != nil {
		return nil, err
	}

	return &CalendarTrainingPlan{name, events}, nil
}

func InsertRunningEvents(srv *calendar.Service, events []*calendar.Event) error {
	for _, event := range events {
		_, err := srv.Events.Insert("scott.d.frazer@gmail.com", event).Do()
		fmt.Printf("Inserted event %v\n", event)
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteRunningEvents(srv *calendar.Service, events []*calendar.Event) error {
	for _, event := range events {
		err := srv.Events.Delete("scott.d.frazer@gmail.com", event.Id).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func GetRunningEvents(srv *calendar.Service, start, end time.Time) ([]*calendar.Event, error) {
	events, err := srv.Events.
		List("scott.d.frazer@gmail.com").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		MaxResults(250).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, err
	}

	var runningEvents []*calendar.Event

	for _, event := range events.Items {
		if strings.HasPrefix(event.Id, "running") {
			runningEvents = append(runningEvents, event)
		}
	}

	return runningEvents, nil
}
