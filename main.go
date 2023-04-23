/*
solar collector downtime monitor for hamphack 23
by ian anderson
*/

/*
email contents:
"These are the two websites showing production:
Amherst solar field (near Atkins): https://mysolarcity.com/Share/992e1fa2-c42f-46aa-9099-238379a01726#/monitoring/historical/day

Hadley solar field: https://mysolarcity.com/Share/a4d56ea3-2b97-42f2-b0b4-704e468b3161#/monitoring/historical/day

Both fields are currently having problems, but by going back a few days, you can see what "normal" conditions look like. For example, the Amherst field for last Thursday:
https://mysolarcity.com/Share/992e1fa2-c42f-46aa-9099-238379a01726#/monitoring/historical/day?date=2023-04-13"
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"
) 
var err error
var hclient = http.Client{}
var beginURL string = "https://mysolarcity.com/solarcity-api/powerguide/v1.0/"
var userID = "f11af9e2-37ba-4177-aeeb-1bf1e99ff7c3"
var amherst string = "992e1fa2-c42f-46aa-9099-238379a01726"
var hadley string = "a4d56ea3-2b97-42f2-b0b4-704e468b3161"
var endURL string = "#/monitoring/historical/day"

var from string = "0"
var pass string = "0"
var to string = "0"
type Panels struct {
	TotalConsumptionInIntervalkWh float64 `json:"TotalConsumptionInIntervalkWh"`
	Consumption                   []struct {
		Timestamp                string  `json:"Timestamp"`
		ConsumptionInIntervalkWh float64 `json:"ConsumptionInIntervalkWh"`
		DataStatus               string  `json:"DataStatus"`
	} `json:"Consumption"`
	Appliances []any `json:"Appliances"`
}

func main() {
	flag.StringVar(&from, "from", "0", "Gmail account to send notifications from")
	flag.StringVar(&pass, "pass", "0", "App password for above account")
	flag.StringVar(&to, "to", "0", "Email account to send notifications to")
	flag.Parse()
	logfile, err := os.OpenFile("Hampshire-Solar-Downtime-Monitor " + time.Now().Format("2006-01-02 15:04:05"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil { log.Fatalln(err) }
	defer logfile.Close()
	log.SetOutput(logfile)
	fmt.Println("Welcome to the Hampshire Solar Downtime Monitor!")
	if from == "0" {
		fmt.Println("Enter the email address to send notifications from:")
		_, err = fmt.Scanln(&from)
		if err != nil { log.Fatalln(err) }
	}
	if pass == "0" {
		fmt.Println("Enter the password to the above email address:")
		_, err = fmt.Scanln(&pass)
		if err != nil { log.Fatalln(err) }
	}
	if to == "0" {
		fmt.Println("Enter the email address to send notifications to:")
		_, err = fmt.Scanln(&to)
		if err != nil { log.Fatalln(err) }
	}
	fmt.Println("Welcome to Hampshire Solar Downtime Monitor")
	for {
		if(time.Now().Minute() < 5) {
			time.Sleep(time.Minute * 5)
		}
		panelRequests(hclient, amherst)
		panelRequests(hclient, hadley)
		time.Sleep(time.Hour)
	}
}
func panelRequests(hclient http.Client, panelID string) {
	town := ""
	if(string(panelID[0]) == "9") {
		town = "Amherst"
	} else {
		town = "Hadley"
	}
	var currentStuff Panels
	currentTime := time.Now()
	currentTimeNum := currentTime.Hour()
	currentTimeStart := strconv.Itoa(currentTimeNum) + ":05:00"
	currentTimeEnd := strconv.Itoa(currentTimeNum + 1) + ":05:00"
	today := currentTime.Format("2006-01-02") + "T"
	currentRequest, err := http.NewRequest("GET", beginURL + "consumption/" + panelID + "?ID=" + userID + "&StartTime=" + today + currentTimeStart + "&EndTime=" + today + currentTimeEnd + "&Period=Hour", nil)
	if err != nil { log.Fatalln(err) }
	currentResponse, err := hclient.Do(currentRequest)
	if err != nil { log.Fatalln(err) }
	currentBody, err := ioutil.ReadAll(currentResponse.Body)
	if err != nil { log.Fatalln(err) }
	err = json.Unmarshal([]byte(string(currentBody)), &currentStuff)
	if err != nil { log.Fatalln(err) }
	for i := 0; i < len(currentStuff.Consumption); i++ {
		num, err := fmt.Println(town + " TIMESTAMP = " + currentStuff.Consumption[i].Timestamp, "KWH = " + fmt.Sprintf("%f", currentStuff.Consumption[i].ConsumptionInIntervalkWh))
		if err != nil { log.Fatalln(err) }
		if(num == 0) {
			// fail condition, send email!
			send(from, pass, to, "The solar panels in " + town + "are down!", town)
		}
	}
}
func send(from string, pass string, to string, body string, town string) {
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + town + " solar panels are down!\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Fatalln("smtp error: ", err)
		return
	}
	log.Println("Email sent to " + to)
}