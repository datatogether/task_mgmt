package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func updateKiwixSources(db *sql.DB) error {
	zims, err := FetchKiwixZims()
	if err != nil {
		return err
	}

	sources, err := ReadSources(db, "created DESC", 100, 0)
	if err != nil {
		return err
	}

	for _, s := range sources {
		for _, z := range zims {
			if s.Url == z.Url {
				if err := z.FetchMd5(); err != nil {
					return err
				}

				if z.Md5 != s.Checksum {
					s.Title = z.Title()
					s.Checksum = z.Md5
					s.Save(db)
				}
			}
		}
	}

	return nil
}

// url to scrape
const kiwixContentListUrl = "http://wiki.kiwix.org/wiki/Content_in_all_languages"

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
	// url to get MD5 entry
	Md5Url string
	// md5 checksum for download
	Md5 string
	// download url
	Url string
}

func (z *KiwixZim) Title() string {
	return fmt.Sprintf("%s - %s", z.Project, z.Language)
}

func (z *KiwixZim) FetchMd5() error {
	res, err := http.Get(z.Md5Url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	md5 := make([]byte, 32)
	res.Body.Read(md5)
	z.Md5 = string(md5)
	return nil
}

func FetchKiwixZims() ([]*KiwixZim, error) {
	zims := []*KiwixZim{}

	doc, err := goquery.NewDocument(kiwixContentListUrl)
	if err != nil {
		return nil, err
	}

	// relatively brittle scraping code, will need to be kept in sync kiwixContentListUrl
	doc.Find("#zimtable tbody tr").Each(func(i int, el *goquery.Selection) {
		if i == 0 {
			// skip header row
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
				zim.Url, _ = el.Find("a").Attr("href")
				md5Url, _ := el.Find("sup small a").Attr("href")
				if md5Url != "" {
					zim.Md5Url = md5Url
				}
			}
		})

		zims = append(zims, zim)
	})

	return zims, nil
}
