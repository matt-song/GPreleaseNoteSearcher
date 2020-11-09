/*
Author: Matt Song (matt.song@live.cn)
Date:   Aug 31st, 2020

Steps:
1. get the version list from the latest lin, put them into a map
2. get content of each of link, generate the data for tables

to be done: only get uniuq value of g
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pborman/getopt/v2"
)

var (
	enableDebug      = false     // default disable debug mode, unless use -D
	latestRelaseURLs = []string{ // The link for the latest URL
		"https://gpdb.docs.pivotal.io/43latest/main/index.html",
		"https://gpdb.docs.pivotal.io/5latest/main/index.html",
		"https://gpdb.docs.pivotal.io/6latest/main/index.html",
	}
)

func init() {

	/* get the options */
	optDebug := getopt.Bool('D', "Display debug message") // enable DEBUG mode
	optHelp := getopt.Bool('H', "Help")                   // print the help message

	getopt.Parse()
	enableDebug = *optDebug

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}
}

func main() {

	var allGPVerMap map[string][]string
	allGPVerMap = getReleaseNoteList(latestRelaseURLs)
	for targetGpFamily := range allGPVerMap {
		// fmt.Printf("key[%s] value[%s]\n", targetGpFamily, allGPVerMap[targetGpFamily])
		switch targetGpFamily {
		/* for 4x, the release note url is like:
		https://gpdb.docs.pivotal.io/43latest/relnotes/GPDB_43latest_README.html
		https://gpdb.docs.pivotal.io/43310/relnotes/GPDB_43latest_README.html
		*/
		case "4":
			for _, gp4xVer := range allGPVerMap[targetGpFamily] {
				relNoteURL := "https://gpdb.docs.pivotal.io/" + gp4xVer + "/relnotes/GPDB_43latest_README.html"
				plog("DEBUG", "find release note url: "+relNoteURL)
				parseURL4x(relNoteURL)
			}
		case "5":

		case "6":
		}

	}

}

func parseURL4x(url string) {

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

// get the content of url like curl in linux
/*
func parseURL(url string) {

	plog("DEBUG", "Parsing url ["+url+"]")
	doc, _ := html.Parse(strings.NewReader(url))
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					fmt.Println(a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}
*/

func getReleaseNoteList(latestURLs []string) (m map[string][]string) {

	var gpV4List, gpV5List, gpV6List []string
	gpVerList := make(map[string][]string)

	// loop against each gp family - gpv4,v5,v6...
	plog("INFO", "Getting the GP version list...")
	for _, url := range latestURLs {

		// get the GPDB family
		gpFamily := strings.Split(url, "/")[3]
		plog("INFO", "Checking the GP version list based on url: "+url)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		} else {
			plog("DEBUG", "Done, Processing the content...")
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Get all the href tag in the page, and filter with the keyword we need to find out all aviliable gp version
		doc.Find("a").Each(func(i int, selection *goquery.Selection) {
			qHref, _ := selection.Attr("href")
			for _, link := range strings.Split(qHref, "\n") {
				switch gpFamily {
				case "43latest":
					r, _ := regexp.Compile("43[0-9]+/common/welcome.html")
					if r.MatchString(link) {
						gpVer4x := strings.Split(link, "/")[1]
						gpV4List = append(gpV4List, gpVer4x)
					}
				case "5latest":
					r, _ := regexp.Compile("5[0-9]+/main/index.htm")
					if r.MatchString(link) {
						gpVer5x := strings.Split(link, "/")[1]
						gpV5List = append(gpV5List, gpVer5x)
					}
				case "6latest":
					r, _ := regexp.Compile("6-[0-9]+/main/index.html")
					if r.MatchString(link) {
						gpVer6x := strings.Split(link, "/")[1]
						gpV6List = append(gpV6List, gpVer6x)
					}
				}

			}
		})
	}
	gpVerList["4"] = gpV4List
	gpVerList["5"] = gpV5List
	gpVerList["6"] = gpV6List

	return gpVerList
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
