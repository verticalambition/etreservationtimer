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
	fmt.Println("Starting Reservation Application")
	l, _ := time.LoadLocation("America/Denver")
	cronTimer := cron.New(cron.WithLocation(l), cron.WithSeconds())
	cronTimer.AddFunc("5 0 6,8,10,12,14,16,18,20 * * *", func() {
		hour, amPm := parseCurrentTime()
		ProcessValidRequests(hour, amPm)
	})

	cronTimer.Start()

	r := mux.NewRouter()
	r.HandleFunc("/et/reservation", processReservationRequest).Methods(http.MethodPost)
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Received Reservation Request"))
	})

	http.Handle("/", r)
	log.Fatalln(http.ListenAndServe(":8080", nil))
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
	w.Write([]byte("Reservation Successfully Processed"))
}

func writeRequestToFile(request ReservationRequest) {
	f, err := os.OpenFile("reservations.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer f.Close()
	if _, err := f.WriteString(request.Time + "," +
		request.Ampm + "," +
		request.Week + "," + request.Day +
		"\n"); err != nil {
		log.Println(err)
	}
}

func ProcessValidRequests(time int, ampm string) {

	f, err := os.Open("reservations.txt")
	var newFile string
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		wholeTime := strings.Split(line, ",")
		hour, _ := strconv.Atoi(wholeTime[0])
		if hour == time && wholeTime[1] == ampm {
			fmt.Println("Going to submit " + line + " for processing")
			resp, err := http.Get("http://172.17.0.3:8082/et1/" + strconv.Itoa(hour) + "/" + ampm + "/" + wholeTime[2] + "/" + wholeTime[3])
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Response from Reservation request was " + resp.Status)
			fmt.Println("And then remove entry " + line + " from file")
		} else {
			newFile += line + "\n"
		}
	}
	fmt.Println("Contents of New File are " + newFile)
	f.Close()
	replaceFile, err := os.OpenFile("reservations.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	replaceFile.Write([]byte(newFile))
	replaceFile.Close()
}

type ReservationRequest struct {
	Time        string `json:"time"`
	Ampm        string `json:"ampm"`
	Week        string `json:"week"`
	Day         string `json:"day"`
	UserDetails struct {
		FirstName   string `json:"firstName"`
		MiddleName  string `json:"middleName"`
		BirthMonth  string `json:"birthMonth"`
		BirthYear   string `json:"birthYear"`
		BirthDay    string `json:"birthDay"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	} `json:"userDetails"`
}
