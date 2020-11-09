package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
//DEBUG = 0 // enable debug mode
)

var enableDebug = true

func main() {

	url := "https://gpdb.docs.pivotal.io/43330/relnotes/GPDB_43latest_README.html"
	crawlWebsite(url)
}

func crawlWebsite(url string) {

	allResolvedIssueMap := make(map[string]map[string]string)

	plog("DEBUG", "Reading content from url:"+url)

	// Get the full content of the page
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Done, Processing the content...")
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var curIssueID string
	// find out all div with class = tablenoborder
	doc.Find("div.tablenoborder").Each(func(i int, allDiv *goquery.Selection) {

		// find out all element within the div with id = topic20__, this is the id for resolved issue
		allDiv.Find("[id^=topic20__]").Each(func(j int, allStuffHaveID *goquery.Selection) {

			// find out all td within the div, that what we need
			allStuffHaveID.Find("td.entry").Each(func(id int, allTd *goquery.Selection) {
				hash := make(map[string]string)
				targetColumnNo := id % 4
				if targetColumnNo == 0 {

					curIssueID = allTd.Text()
					plog("ERROR", "Find Issue ID: "+curIssueID)
					// allResolvedIssueMap["curIssueID"] = ""
				} else {
					switch targetColumnNo {
					case 1:
						hash["category"] = allTd.Text()
						allResolvedIssueMap[curIssueID] = hash
					case 2:
						hash["resolved"] = allTd.Text()
						allResolvedIssueMap[curIssueID] = hash
					case 3:
						hash["description"] = allTd.Text()
						allResolvedIssueMap[curIssueID] = hash
					}
				}
			})
		})
	})

	b, _ := json.MarshalIndent(allResolvedIssueMap, "", "  ")
	plog("DEBUG", string(b))
}

// colorful output 23333..
func plog(logLevel string, message string) {

	// define the color code here:
	lightRed := "\033[38;5;9m"
	red := "\033[38;5;1m"
	green := "\033[38;5;2m"
	yellow := "\033[38;5;3m"
	cyan := "\033[38;5;14m"
	//darkBlue := "\033[38;5;25m"
	normal := "\033[39;49m"

	var colorCode string
	var errorOut = 0

	switch logLevel {
	case "INFO":
		colorCode = green
	case "WARN":
		colorCode = yellow
	case "ERROR":
		colorCode = lightRed
	case "FATAL":
		colorCode = red
		errorOut = 1
	case "DEBUG":
		if enableDebug == true {
			colorCode = cyan
		} else {
			return
		}
	default:
		colorCode = normal
	}
	curTime := time.Now()
	fmt.Printf("%s"+curTime.Format("2006-01-02 15:04:05")+" [%s] %s\n", colorCode, logLevel, message)
	fmt.Printf("%s", normal)
	if errorOut == 1 {
		os.Exit(1)
	}
}
