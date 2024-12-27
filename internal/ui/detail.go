package ui

import (
	"fmt"
	"os"
	"strings"

	"v2ex-tui/internal/crawler"
	"v2ex-tui/internal/model"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const (
	ReserveSpaceHeight  = 5
	ReserveSpaceWidth   = 0
	ReserveSpaceContent = 10
)

type DetailPage struct {
	Topic    model.Topic
	loading  bool
	err      error
	spinner  spinner.Model
	crawler  *crawler.Crawler
	selected int
	viewport viewport.Model
}

func NewDetailPage() *DetailPage {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 100
		height = 30
	}
	vp := viewport.New(width-ReserveSpaceWidth, height-ReserveSpaceHeight)
	vp.Style = lipgloss.NewStyle().Padding(1, 2)

	return &DetailPage{
		loading:  true,
		spinner:  s,
		crawler:  crawler.New(),
		selected: 0,
		viewport: vp,
	}
}

func (d *DetailPage) LoadTopic(topic model.Topic) tea.Cmd {
	d.Topic = topic
	d.loading = true
	return d.fetchTopicDetail
}

func (d *DetailPage) fetchTopicDetail() tea.Msg {
	topic, err := d.crawler.FetchTopicDetail(d.Topic.URL)
	if err != nil {
		return errMsg{err}
	}

	// 统计每条评论的回复数
	for i := range topic.Replies {
		replyCount := 0
		currentAuthor := topic.Replies[i].Author
		// 遍历后续的评论来统计回复数
		for _, reply := range topic.Replies[i+1:] {
			if strings.HasPrefix(strings.TrimSpace(reply.Content), "@"+currentAuthor) {
				replyCount++
			}
		}
		topic.Replies[i].ReplyCount = replyCount
	}

	return topicDetailMsg{*topic}
}

type topicDetailMsg struct {
	topic model.Topic
}

func (d *DetailPage) Update(msg tea.Msg) (*DetailPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			d.viewport.LineUp(1)
			return d, nil
		case "down", "j":
			d.viewport.LineDown(1)
			return d, nil
		case "f":
			err := clipboard.WriteAll(d.Topic.URL)
			if err != nil {
				d.err = fmt.Errorf("failed to copy URL: %w", err)
			}
			return d, nil
		}

	case topicDetailMsg:
		d.loading = false
		d.Topic = msg.topic
		return d, nil

	case errMsg:
		d.err = msg
		d.loading = false
		return d, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd

	case tea.WindowSizeMsg:
		vp := viewport.New(msg.Width-ReserveSpaceWidth, msg.Height-ReserveSpaceHeight)
		vp.Style = lipgloss.NewStyle().Padding(1, 2)
		d.viewport = vp
		return d, nil

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			d.viewport.LineUp(3) // Scroll up by 3 lines
		case tea.MouseWheelDown:
			d.viewport.LineDown(3) // Scroll down by 3 lines
		}
		return d, nil
	}

	return d, nil
}

func (d *DetailPage) View() string {
	if d.loading {
		return titleStyle.Render("话题详情") + "\n" +
			d.spinner.View() + " 加载中...\n"
	}

	if d.err != nil {
		return errorStyle.Render("Notification: "+d.err.Error()) + "\n"
	}

	content := strings.Builder{}

	// 标题区域
	content.WriteString(
		titleStyle.Render(IconTitle+"话题: "+d.Topic.Title) + "\n" +
			subtitleStyle.Render(IconAuthor+"楼主: "+d.Topic.Author) + "\n" +
			subtitleStyle.Render(IconTime+"活跃时间: "+d.Topic.Time) + "\n")

	// 内容区域
	content.WriteString(
		titleStyle.Render(IconContent+"内容:") + "\n" +
			contentStyle.Render(d.wrapText(d.Topic.Content)) + "\n",
	)

	// 评论区域
	content.WriteString(titleStyle.Render(IconComments+"评论:") + "\n")

	// 格式化展示每条评论
	if len(d.Topic.Replies) > 0 {
		for _, reply := range d.Topic.Replies {
			content.WriteString(subtitleStyle.Render(fmt.Sprintf("%s%s 于 %s 回复:",
				IconAuthor,
				reply.Author,
				reply.Time,
			)) + "\n")

			wrappedContent := d.wrapText(reply.Content)
			content.WriteString(contentStyle.Render(wrappedContent) + "\n")
		}
	} else {
		content.WriteString(contentStyle.Render("暂无评论") + "\n")
	}

	d.viewport.SetContent(content.String())

	return d.viewport.View() + "\n" +
		subtitleStyle.Render(IconBack+" 「空格」返回 | ↑↓ 滚动 ｜ f 复制链接 | q 退出 | "+IconMouse+" 支持鼠标操作(按 m 退出鼠标模式后可以选中文本)\n")
}

func (d *DetailPage) wrapText(text string) string {
	// 处理空字符串情况
	if text == "" {
		return ""
	}

	// 设置实际可用宽度（考虑 padding 等）
	maxWidth := d.viewport.Width - ReserveSpaceContent // 减去 padding 和一些边距
	if maxWidth <= 0 {
		maxWidth = 80 // 默认宽度
	}

	var lines []string
	var currentLine string

	// 按 UTF-8 字符分割，而不是按字节分割
	chars := []rune(text)

	for _, char := range chars {
		currentWidth := lipgloss.Width(currentLine + string(char))
		if currentWidth > maxWidth {
			lines = append(lines, strings.TrimSpace(currentLine))
			currentLine = string(char)
		} else {
			currentLine += string(char)
		}
	}

	if currentLine != "" {
		lines = append(lines, strings.TrimSpace(currentLine))
	}

	return strings.Join(lines, "\n")
}
