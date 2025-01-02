package crawler

import (
	"net/http"
	"strings"
	"v2ex-tui/internal/config"

	"v2ex-tui/internal/model"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	client *http.Client
}

func New() *Crawler {
	proxyConfig, err := config.LoadConfig()
	var proxyAddr string
	if err == nil {
		proxyAddr = proxyConfig.Proxy.Socks5
	}
	client, err := createHttpClient(proxyAddr)
	if err != nil {
		return nil
	}
	return &Crawler{
		client: client,
	}
}

func (c *Crawler) FetchTopics() ([]model.Topic, error) {
	resp, err := c.client.Get("https://www.v2ex.com/?tab=all")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var topics []model.Topic
	doc.Find(".cell.item").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".item_title a").Text()
		url, _ := s.Find(".item_title a").Attr("href")
		author := s.Find("strong a").First().Text()
		comments := s.Find(".count_livid").Text()
		if comments == "" {
			comments = "0"
		}
		timeStr := s.Find(".topic_info").Text()
		timeStr = strings.Split(timeStr, "•")[2]
		timeStr = strings.TrimSpace(timeStr)

		if !strings.HasPrefix(url, "http") {
			url = "https://www.v2ex.com" + url
		}

		topics = append(topics, model.Topic{
			Title:    title,
			Author:   author,
			Comments: comments,
			Time:     timeStr,
			URL:      url,
		})
	})

	return topics, nil
}

func (c *Crawler) FetchTopicDetail(url string) (*model.Topic, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	topic := &model.Topic{
		Title:   doc.Find("h1").Text(),
		Author:  doc.Find(".header small a").First().Text(),
		Time:    doc.Find(".header small span").First().AttrOr("title", ""),
		Content: doc.Find(".topic_content").Text(),
		URL:     url,
	}

	// 获取评论，并识别被回复的评论
	var replies []model.Reply
	doc.Find(".cell[id^='r_']").Each(func(i int, s *goquery.Selection) {
		content := s.Find(".reply_content").Text()
		replyTo := ""
		if strings.HasPrefix(strings.TrimSpace(content), "@") {
			replyTo = strings.Split(content, " ")[0][1:]
		}

		reply := model.Reply{
			Author:  s.Find("strong a").Text(),
			Time:    s.Find(".ago").Text(),
			Content: content,
			Number:  s.Find(".no").Text(),
			ReplyTo: replyTo,
		}
		replies = append(replies, reply)
	})

	// 计算每条评论被回复的次数
	for i := range replies {
		count := 0
		for _, r := range replies {
			if r.ReplyTo == replies[i].Author {
				count++
			}
		}
		replies[i].ReplyCount = count
	}

	topic.Replies = replies
	return topic, nil
}
