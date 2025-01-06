package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"v2ex-tui/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	socks5Addr string
	client     *http.Client
)

type page int

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
	flag.StringVar(&socks5Addr, "proxy", "", "use socks5 proxy")
	flag.Parse()
	if socks5Addr != "" {
		proxyFun := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(socks5Addr)
		}
		transport := &http.Transport{
			Proxy: proxyFun,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client = &http.Client{
			Transport: transport,
		}
	} else {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		return
	}
}
