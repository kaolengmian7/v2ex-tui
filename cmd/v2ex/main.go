package main

import (
	"flag"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"v2ex-tui/internal/crawler"
	"v2ex-tui/internal/ui"
)

type page int

var proxyURL string

const (
	homeView page = iota
	detailView
)

type model struct {
	currentPage  page
	homePage     *ui.HomePage
	detailPage   *ui.DetailPage
	mouseEnabled bool
}

func initialModel() model {
	var client *crawler.Crawler
	if proxyURL != "" {
		client = crawler.New(
			crawler.WithProxy(proxyURL),
		)
	} else {
		client = crawler.New()
	}
	return model{
		currentPage: homeView,
		homePage:    ui.NewHomePage(client),
		detailPage:  ui.NewDetailPage(client),
	}
}

func (m model) Init() tea.Cmd {
	return m.homePage.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "backspace", " ":
			if m.currentPage == detailView {
				m.currentPage = homeView
				return m, nil
			}
		case "enter":
			if m.currentPage == homeView {
				if topic := m.homePage.GetSelectedTopic(); topic != nil {
					m.currentPage = detailView
					return m, m.detailPage.LoadTopic(*topic)
				}
			}
		case "m": // 假设使用 "m" 键切换鼠标支持
			m.mouseEnabled = !m.mouseEnabled
			if m.mouseEnabled {
				return m, tea.EnableMouseCellMotion
			} else {
				return m, tea.DisableMouse
			}
		}

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseLeft:
			if m.currentPage == homeView {
				if topic := m.homePage.GetSelectedTopic(); topic != nil {
					m.currentPage = detailView
					return m, m.detailPage.LoadTopic(*topic)
				}
			}
		}
	}

	var cmd tea.Cmd
	switch m.currentPage {
	case homeView:
		m.homePage, cmd = m.homePage.Update(msg)
	case detailView:
		m.detailPage, cmd = m.detailPage.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	switch m.currentPage {
	case homeView:
		return m.homePage.View()
	case detailView:
		return m.detailPage.View()
	default:
		return "Unknown view"
	}
}

func main() {
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL (e.g., http://localhost:8080 or socks5://localhost:1080)")
	flag.Parse()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		return
	}
}
