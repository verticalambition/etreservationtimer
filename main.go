package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	//"github.com/robfig/cron"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	fmt.Println("Starting Application")
	cronTimer := cron.New(cron.WithSeconds())
	cronTimer.AddFunc("5 0 6,8,10,12,14,16,18,20 * * *", func() {
		hour, amPm := parseCurrentTime()
		processValidRequests(hour, amPm)
	})

	cronTimer.Start()

	r := mux.NewRouter()
	r.HandleFunc("/et/reservation", processReservationRequest).Methods(http.MethodPost)
	http.Handle("/", r)
	log.Fatalln(http.ListenAndServe(":80", nil))
	//Set Cron job to run every day 1 in the afternoon

}

func parseCurrentTime() (int, string) {
	t := time.Now()
	fmt.Println("Executing job at " + t.String())
	mst, err := time.LoadLocation("America/Denver")
	//Check if valid location
	if err != nil {
		fmt.Println(err)
	}
	localTime := t.In(mst)
	fmt.Println(localTime.Format("2006-01-02 3:4:5 PM"))
	hour := localTime.Hour()
	//Account for military time which website doesn't use
	if hour != 12 {
		hour = hour % 12
	}
	var amPm string
	formattedTime := localTime.Format("2006-01-02 3:4:5 PM")
	if strings.Contains(formattedTime, "PM") {
		amPm = "PM"
	} else {
		amPm = "AM"
	}
	return hour, amPm
}

func processReservationRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received Reservation Request")
	var reservationRequest ReservationRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reservationRequest); err != nil {
		w.Write([]byte(err.Error()))
	}

	defer r.Body.Close()
	writeRequestToFile(reservationRequest)
}

func writeRequestToFile(request ReservationRequest) {
	f, err := os.OpenFile("reservations.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer f.Close()
	if _, err := f.WriteString(strconv.Itoa(request.Time) + "," + request.Ampm + "," + strconv.Itoa(request.Week) + "," + strconv.Itoa(request.Day) + "\n"); err != nil {
		log.Println(err)
	}
	//Test code
	//processValidRequests(request.Time, request.Ampm)
}

func processValidRequests(time int, ampm string) {
	f, err := os.Open("reservations.txt")
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		wholeTime := strings.Split(line, ",")
		hour, _ := strconv.Atoi(wholeTime[0])
		if hour == time && wholeTime[1] == ampm {
			fmt.Println("Going to submit " + line + " for processing")
		}
	}
}

type ReservationRequest struct {
	Time int    `json:"time"`
	Ampm string `json:"ampm"`
	Week int    `json:"week"`
	Day  int    `json:"day"`
}
