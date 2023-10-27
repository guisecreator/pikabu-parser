package parser

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Parser struct {
	Url        string
	Tags       []string
	ID         string
	AppINFO    string
	Posts      []map[string]interface{}
	HTTPClient *http.Client
	Blocks     []*goquery.Selection
	EntryTree  *goquery.Document
}

func New(Url string) *Parser {
	return &Parser{Url: Url}
}

func (p *Parser) GetPosts() []map[string]interface{} {
	resp, err := http.Get(p.Url)
	if err != nil {
		fmt.Println("Error:", err)
		return p.Posts
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return p.Posts
	}

	_, name, _ := charset.DetermineEncoding(body, resp.Header.Get("Content-Type"))
	fmt.Printf("Charset: %s\n", name)
	if name == "windows-1251" {
		decoder := charmap.Windows1251.NewDecoder()
		toUtf8, errDecoder := decoder.Bytes(body)
		if errDecoder != nil {
			fmt.Println("decoder err", err)
			return p.Posts
		}
		body = toUtf8
	}

	pageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return p.Posts
	}

	pageReader := bytes.NewReader(pageBytes)
	doc, err := goquery.NewDocumentFromReader(pageReader)
	if err != nil {
		log.Println("doc err", err)
		return p.Posts
	}

	doc.Find(".story").Each(func(i int, s *goquery.Selection) {
		//TODO: fix post id and delete client in this function
		client := http.Client{}

		postID := p.getPostID(s, client)
		if p.existPost(postID) {
			return
		}

		postLink := p.getPostLink(s)
		postDoc, docError := goquery.NewDocument(postLink)
		if docError != nil {
			log.Println("doc error", docError)
			return
		}

		if p.ignorePost(postDoc.Selection) {
			return
		}

		PostDate := p.getPostDate(postDoc.Selection)
		PostTitle := p.getPostTitle(postDoc.Selection)
		PostTags := p.getPostTags(postDoc.Selection)
		PostID := p.getPostID(postDoc.Selection, client)

		post := map[string]interface{}{
			"PostDate":  PostDate,
			"PostTitle": PostTitle,
			"PostTags":  PostTags,
			"PostID":    PostID,
			"PostLink":  postLink,
		}

		p.Posts = append(p.Posts, post)
	})

	return p.Posts
}

func (p *Parser) getPostID(blockTree *goquery.Selection, client http.Client) string {
	href := blockTree.
		Find(".story__title-link").
		First().
		AttrOr("href", "")
	fmt.Printf("href: %s\n", href)

	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}

	data, err := client.Get(href)
	if err != nil {
		log.Printf("Error get Attr0r: %v", err)
		return ""
	}

	dataBytes, err := io.ReadAll(data.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return ""
	}

	data.Body.Close()

	dataDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(dataBytes))
	if err != nil {
		log.Printf("Error doc: %v", err)
		return ""
	}

	return dataDoc.
		Find("a").
		First().
		AttrOr("href", "")
}

func (p *Parser) getPostText(text *goquery.Selection) string {
	return text.Text()
}

func (p *Parser) getPostLink(link *goquery.Selection) string {
	attr, _ := link.
		Find(".story__title-link").
		Attr("href")
	if attr == "" {
		fmt.Printf("no link \n")
	} else {
		fmt.Printf("Post link: %s\n", attr)
	}

	toStr := attr
	return toStr
}

func (p *Parser) ignorePost(ignore *goquery.Selection) bool {
	for _, tag := range p.Tags {
		if ignore.Find(tag).Length() > 0 {
			return true
		}
	}
	return false
}

func (p *Parser) getPostDate(date *goquery.Selection) string {
	postdate, _ := date.Find(".story__datetime[datetime]").Attr("datetime")
	if postdate == "" {
		fmt.Printf("no post date \n")
	} else {
		fmt.Printf("post date: %s\n", postdate)
	}

	return postdate
}

func (p *Parser) getPostTitle(title *goquery.Selection) string {
	postTitle := title.Find(".story__title-link").Text()
	if postTitle == "" {
		fmt.Printf("no post title \n")
	} else {
		fmt.Printf("post title: %s\n", postTitle)
	}

	//htmlContent := title.Find(".story__title-link").Text()
	//fmt.Printf("HTML content: %s\n", htmlContent)

	return postTitle
}

func (p *Parser) getPostTags(tag *goquery.Selection) []string {
	var tags []string

	tagsNodes := tag.Find(".story__tags .tags__tag")
	tagsNodes.Each(func(i int, tag *goquery.Selection) {
		tags = append(tags, tag.Text())
	})

	return tags
}

func (p *Parser) existPost(postID string) bool {
	if postID == "" {
		return true
	}

	return false
}

func (p *Parser) Timer(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}
