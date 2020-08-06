package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/pkg/browser"
)

type Film struct {
	Slug  string
	Image string
	Name  string
}

const url = "https://letterboxd.com/ajax/poster"
const urlEnd = "menu/linked/125x187/"
const site = "https://letterboxd.com"

var wg sync.WaitGroup

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "please provide letterboxd username")
		os.Exit(1)
	}
	siteToVisit := site + "/" + args[0] + "/watchlist"
	rand.Seed(time.Now().Unix())
	var posters []Film
	ajc := colly.NewCollector()
	ajc.OnHTML("div.film-poster", func(e *colly.HTMLElement) {
		name := e.Attr("data-film-name")
		slug := e.Attr("data-target-link")
		img := e.ChildAttr("img", "src")
		tempfilm := Film{
			Slug:  (site + slug),
			Image: img,
			Name:  name,
		}
		posters = append(posters, tempfilm)
		wg.Done()
	})
	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 50})
	c.OnHTML(".poster-container", func(e *colly.HTMLElement) {
		e.ForEach("div.film-poster", func(i int, ein *colly.HTMLElement) {
			slug := ein.Attr("data-film-slug")
			wg.Add(1)
			go ajc.Visit(url + slug + urlEnd)
		})
		wg.Wait()

	})
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "watchlist/page") {
			e.Request.Visit(e.Request.AbsoluteURL(link))
		}
	})

	c.Visit(siteToVisit)

	n := rand.Int() % len(posters)
	browser.OpenURL(posters[n].Slug)

}
