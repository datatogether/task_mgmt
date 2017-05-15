package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var zims = []*KiwixZim{}

func init() {
	data, err := ioutil.ReadFile("kiwix.json")
	if err != nil {
		logger.Println(err.Error())
		return
	}
	if err := json.Unmarshal(data, &zims); err != nil {
		logger.Println(err.Error())
	}
}

// url to scrape
const contentListUrl = "http://wiki.kiwix.org/wiki/Content_in_all_languages"

// KiwixZim represents a Zim package on the Kiwix service
type KiwixZim struct {
	// Name of the project that's been archived
	Project string
	// ISO Language code
	Language string
	// Size String (eg: "11M" or "5.4G")
	Size string
	// Created Date in form YYYY-MM
	Created string
	// Weather it's all content or not
	FullContent bool
	// md5 checksum for download
	Md5 string
	// download url
	Url string
}

func scanKiwix() {
	archives := []*KiwixZim{}

	doc, err := goquery.NewDocument(contentListUrl)
	if err != nil {
		log.Fatal(err.Error())
	}

	// relatively brittle scraping code, will need to be kept in sync contentListUrl
	doc.Find("#zimtable tbody tr").Each(func(i int, el *goquery.Selection) {
		md5Url := ""
		if i == 0 {
			return
		}

		zim := &KiwixZim{}

		el.ChildrenFiltered("td").Each(func(j int, cell *goquery.Selection) {
			switch j {
			case 0:
				zim.Project = strings.TrimSpace(cell.Text())
			case 1:
				zim.Language = strings.TrimSpace(cell.Text())
			case 2:
				zim.Size = strings.TrimSpace(cell.Text())
			case 3:
				zim.Created = strings.TrimSpace(cell.Text())
			case 4:
				zim.FullContent = strings.TrimSpace(cell.Text()) == "all"
			case 5:
				log.Println(el.Find("a").Attr("href"))
				zim.Url, _ = el.Find("a").Attr("href")
				md5Url, _ = el.Find("sup small a").Attr("href")
				if md5Url != "" {
					res, err := http.Get(md5Url)
					if err != nil {
						log.Println(err.Error())
					}
					defer res.Body.Close()

					md5 := make([]byte, 32)
					res.Body.Read(md5)
					zim.Md5 = string(md5)
				}
			}
		})

		archives = append(archives, zim)
		time.Sleep(time.Second)
	})

	// write json file
	data, err := json.MarshalIndent(archives, "", "  ")
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := ioutil.WriteFile("kiwix.json", data, os.ModePerm); err != nil {
		log.Fatal(err.Error())
	}
}
