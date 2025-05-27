package ui

import (
	"strings"

	"github.com/kaolengmian7/v2ex-tui/internal/model"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ReplyPage struct {
	mainReply model.Reply   // 当前查看的评论
	replies   []model.Reply // 所有相关回复
	table     table.Model
	selected  int
}

func NewReplyPage() *ReplyPage {
	columns := []table.Column{
		{Title: "序号", Width: 10},
		{Title: "作者", Width: 15},
		{Title: "内容", Width: 80},
		{Title: "时间", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s1 := table.DefaultStyles()
	s1.Header = s1.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s1.Selected = s1.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s1)

	return &ReplyPage{
		table:    t,
		selected: 0,
	}
}

func (r *ReplyPage) LoadReply(reply model.Reply, allReplies []model.Reply) {
	r.mainReply = reply
	r.replies = nil

	// 查找所有回复当前评论的评论
	for _, rep := range allReplies {
		if strings.HasPrefix(strings.TrimSpace(rep.Content), "@"+r.mainReply.Author) {
			r.replies = append(r.replies, rep)
		}
	}

	// 更新表格数据
	var rows []table.Row
	for _, rep := range r.replies {
		rows = append(
			rows, table.Row{
				rep.Number,
				rep.Author,
				rep.Content,
				rep.Time,
			},
		)
	}
	r.table.SetRows(rows)
}

func (r *ReplyPage) Update(msg tea.Msg) (*ReplyPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down":
			r.table, _ = r.table.Update(msg)
			r.selected = r.table.Cursor()
		}
	}
	return r, nil
}

func (r *ReplyPage) View() string {
	var s strings.Builder

	// 原评论区域
	s.WriteString(
		sectionStyle.Render(
			titleStyle.Render(IconContent+" 原评论 (#"+r.mainReply.Number+")")+"\n"+
				subtitleStyle.Render(IconAuthor+" 作者: "+r.mainReply.Author)+"\n"+
				subtitleStyle.Render(IconTime+" 时间: "+r.mainReply.Time)+"\n"+
				contentStyle.Render(r.mainReply.Content),
		) + "\n",
	)

	// 回复列表区域
	if len(r.replies) > 0 {
		s.WriteString(titleStyle.Render(IconComments+" 回复列表:") + "\n")
		s.WriteString(tableStyle.Render(r.table.View()))
	} else {
		s.WriteString(subtitleStyle.Render("暂无回复"))
	}

	s.WriteString("\n" + subtitleStyle.Render(IconBack+" esc 返回评论列表 | q 退出\n"))
	return s.String()
}
