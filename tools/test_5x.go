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
)

const (
//DEBUG = 0 // enable debug mode
)

var enableDebug = true

func main() {

	mainURL := "https://gpdb.docs.pivotal.io/5280/relnotes/gpdb-5latest-release-notes.html"
	mainVer := "5.28.0"
	allRelNote := findOutAllChildRelNote(mainURL, mainVer)
	for id, url := range allRelNote {
		fmt.Println(id)
		plog("INFO", "Checking url ["+url+"]")
		parseURL5x(url)
	}
}

/* ==== Progress and Summary ====

- Takeing 5.28.x for example:

[DONE] Step1: visit https://gpdb.docs.pivotal.io/5280/relnotes/gpdb-5latest-release-notes.html, get the list of release note of 5.28.x

	<a href="/5280/relnotes/gpdb-5283-release-notes.html">Pivotal Greenplum 5.28.3 Release Notes</a>
	<a href="/5280/relnotes/gpdb-5282-release-notes.html">Pivotal Greenplum 5.28.2 Release Notes</a>
	...

[DONE] Step2: go through each page, find out all resuloved issues
	return a map with below structure:

	ID#1 ------>[IssueID]
			|-->[category]
			|-->[Resolved]
			|-->[Description]
	ID#2 ------>[IssueID]
			|-->[category]
			|-->[Resolved]
			|-->[Description]
	...

[INPROGRESS] Step3: send the result to DB

*/

// ====== start work at here =======

/* find out all the sub release note based by main release note url, example code is like:
	<div class="nav-content">
  	<ul>
      <li class="">
        <a href="/5280/relnotes/../../5280/homenav.html">Pivotal Greenplum® 5.28 Documentation</a>
	  </li>
	...*/
func findOutAllChildRelNote(url string, mainVer string) (c []string) {

	plog("DEBUG", "Collecting sub release note from main url ["+url+"]")
	formatedMainVer := strings.ReplaceAll(mainVer, `.`, "")
	plog("DEBUG", "Processing main version ["+formatedMainVer+"]")

	var childUrls []string

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

	// find out div class="nav-content"
	doc.Find("div.nav-content").Each(func(i int, divWithSubUrl *goquery.Selection) {
		// then findout all the sub URL with <a href=...
		divWithSubUrl.Find("a").Each(func(j int, subUrl *goquery.Selection) {
			plog("DEBUG", "Find sub version strings ["+subUrl.Text()+"]")
			subVer := subUrl.Text()
			findVerReg := regexp.MustCompile(`(\d\.\d+\.\d)`)
			if findVerReg.MatchString(subVer) {
				findVer := findVerReg.FindAllString(subVer, -1)[0]
				formatedFindVer := strings.ReplaceAll(findVer, `.`, "")
				plog("DEBUG", "Find sub version ["+formatedFindVer+"]")
				// fmt.Printf("%q\n", findVerReg.FindAll([]byte(subVer), -1))
				// https://gpdb.docs.pivotal.io/5280/relnotes/gpdb-5284-release-notes.html
				targetURL := "https://gpdb.docs.pivotal.io/" + formatedMainVer + "/relnotes/gpdb-" + formatedFindVer + "-release-notes.html"
				plog("DEBUG", "generated sub release note url ["+targetURL+"]")
				childUrls = append(childUrls, targetURL)
			}
		})
	})
	// fmt.Println(childUrls)

	return childUrls
}

func parseURL5x(url string) {

	//allResolvedIssueMap := make(map[string]map[string]string)

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
	curVersion := "n/a"

	/* get the current version, code: <h1 class="title topictitle1">Pivotal Greenplum 5.28.0 Release Notes</h1> */
	doc.Find("h1.title.topictitle1").Each(func(t int, title *goquery.Selection) {
		plog("DEBUG", "Find title strings ["+title.Text()+"]")
		findVerReg := regexp.MustCompile(`(\d\.\d+\.\d)`)
		if findVerReg.MatchString(title.Text()) {
			curVersion = findVerReg.FindAllString(title.Text(), -1)[0]
		}
	})

	/* Find out the detail of resolved issue, code:

	   <div class="topic nested1" id="topic_cq5_vkf_dbb">
	        <dl class="dl parml">

	             <dt class="dt pt dlterm">30923 - Server Execution, Planner</dt>

	             <dd class="dd pd">Resolved a problem where a query could return incorrect results if segments held a
	               NULL value in an empty set.</dd>

	*/
	issueMap := make(map[int][]string)
	// find all div with class "topic nested1" and with ID topic_cq5_vkf_dbb
	doc.Find("div.topic.nested1#topic_cq5_vkf_dbb").Each(func(i int, allDiv *goquery.Selection) {
		// plog("INFO", allDiv.Text())
		// find out all dt with class "dl parml"

		allDiv.Find("dl.dl.parml").Each(func(j int, allDt *goquery.Selection) {

			// init the hash for puting the temporary result
			//var resultArrary []string

			var count = 0
			// find out all dt with class = "dt pt dlterm", this have issue id and category
			allDt.Find("dt.dt.pt.dlterm").Each(func(dtId int, allChildDt *goquery.Selection) {

				returnedDetail := allChildDt.Text()
				// findout issue id and category
				issueID := strings.Split(returnedDetail, " - ")[0]

				// some issue does not have Category, so we put n/a if null
				issueCategory := "n/a"
				if len(strings.Split(returnedDetail, " - ")) > 1 {
					issueCategory = strings.Split(allChildDt.Text(), " - ")[1]
				}

				plog("DEBUG", "=== ID: "+issueID+"; Category: "+issueCategory+" ===")
				issueMap[count] = append(issueMap[count], issueID, issueCategory)
				//issueMap[count] = resultArrary
				count++
			})
			count = 0
			// find out all dd with class = "dd pd", this have issue description
			allDt.Find("dd.dd.pd").Each(func(ddId int, allChildDd *goquery.Selection) {
				// plog("INFO", allChildDd.Text())
				// resultArrary = append(resultArrary, allChildDd.Text())
				issueMap[count] = append(issueMap[count], allChildDd.Text())
				issueMap[count] = append(issueMap[count], curVersion) // add resolved version here
				count++
			})

		})
	})

	// b, _ := json.MarshalIndent(issueMap, "", "  ")
	// plog("DEBUG", string(b))

	for id := range issueMap {
		b, _ := json.MarshalIndent(issueMap[id], "", "  ")
		plog("DEBUG", string(b))
	}
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
