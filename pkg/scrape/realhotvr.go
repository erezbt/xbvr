package scrape

import (
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/mozillazg/go-slugify"
	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr/pkg/models"
)

func RealHotVR(wg *sync.WaitGroup, updateSite bool, knownScenes []string, out chan<- models.ScrapedScene) error {
	defer wg.Done()
	scraperID := "realhotvr"
	siteID := "RealHot VR"
	logScrapeStart(scraperID, siteID)

	siteCollector := createCollector("realhotvr.com")
	sceneCollector := createCollector("realhotvr.com")

	sceneCollector.OnHTML(`html`, func(e *colly.HTMLElement) {
		sc := e.Request.Ctx.GetAny("scene").(models.ScrapedScene)
		sc.SceneType = "VR"
		sc.Studio = "RealHotVR"
		sc.Site = siteID

		// Scene ID
		sc.SiteID = e.Request.Ctx.Get("id")
		sc.SceneID = slugify.Slugify(sc.Site) + "-" + sc.SiteID

		// Cast
		e.ForEach(`.featuring`, func(id int, e *colly.HTMLElement) {
			c := e.ChildText(`.label`)
			// Cast
			if strings.Contains(c, "Featuring:") {
				e.ForEach(`a`, func(id int, ce *colly.HTMLElement) {
					sc.Cast = append(sc.Cast, strings.TrimSpace(ce.Text))
				})
			} else {
				// Tags
				if strings.Contains(c, "Tags:") {
					e.ForEach(`a`, func(id int, ce *colly.HTMLElement) {
						sc.Tags = append(sc.Tags, strings.TrimSpace(ce.Text))
					})
				}
			}
		})

		// Synopsis
		sc.Synopsis = strings.TrimSpace(e.ChildText(`div.videoDetails p`))

		out <- sc
	})

	siteCollector.OnHTML(`div.pagination li a`, func(e *colly.HTMLElement) {
		pageURL := e.Request.AbsoluteURL(e.Attr("href"))
		siteCollector.Visit(pageURL)
	})

	siteCollector.OnHTML(`div.item-video`, func(e *colly.HTMLElement) {
		sceneURL := e.Request.AbsoluteURL(e.ChildAttr("div.item-thumb a", "href"))
		sc := models.ScrapedScene{}

		if !funk.ContainsString(knownScenes, sceneURL) {
			ctx := colly.NewContext()
			tmpCover := e.Request.AbsoluteURL(e.ChildAttr(`img`, "src0_3x"))
			sc.Covers = append(sc.Covers, tmpCover)
			for _, imgSrc := range []string{"src1_3x", "src2_3x", "src3_3x", "src4_3x", "src5_3x"} {
				sc.Gallery = append(sc.Gallery, e.Request.AbsoluteURL(e.ChildAttr(`img`, imgSrc)))
			}
			sc.HomepageURL = sceneURL
			sc.Title = strings.TrimSpace(e.ChildText(`h4`))
			sc.Released = strings.TrimSpace(e.ChildText(`div.date`))
			content := strings.Split(strings.Split(e.ChildText("div.time"), ", ")[1], ":")[0]
			tmpDuration, err := strconv.Atoi(content)
			if err == nil {
				sc.Duration = tmpDuration
			}

			ctx.Put("scene", sc)

			sceneCollector.Request("GET", sceneURL, nil, ctx, nil)
		}
	})

	siteCollector.Visit("https://realhotvr.com/categories/movies/1/latest/")

	if updateSite {
		updateSiteLastUpdate(scraperID)
	}
	logScrapeFinished(scraperID, siteID)
	return nil
}

func init() {
	registerScraper("realhotvr", "RealHotVR", "https://images.povr.com/assets/logos/channels/0/3/3835/200.svg", RealHotVR)
}
