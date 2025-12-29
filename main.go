package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/sys/unix"
)

var (
	logFilePath = "/var/log/golang/todo_tui.log"
	logFile     *os.File
	ws          *unix.Winsize
)

// 定義 Model 結構體，儲存狀態
type model struct {
	title string // 方框頂部標題
	// content string // 方框內的文字
	textInput textinput.Model
	// inputValue string
}

// 初始化 Model
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "group name ...."
	ti.Focus()
	ti.Width = int(ws.Col) - 2
	return model{
		title: "group", // 你的標題
		// content: "這是方框內的內容！\n你可以放多行文字~~",
		textInput: ti,
		// inputValue: "",
	}
}

// Init 方法：啟動時執行的命令（這裡無需）
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update 方法：處理訊息（如按鍵）
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Ctrl+C 退出
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			// m.inputValue = m.textInput.Value()
			m.textInput.SetValue("")
			return m, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View 方法：渲染 UI，使用 Lipgloss 繪製帶標題的方框
func (m model) View() string {
	// logFile
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("Can't open the logFile %s: %v\n", logFilePath, err)
	} else {
		log.SetOutput(logFile)
	}

	// 更新(萬一有賤人換大小) 當前視窗大小
	ws, err = unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		fmt.Print("")
	}

	// 定義內容方框樣式：寬度 40、無頂邊框、藍色邊框、居中對齊
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(false). // 關閉頂邊框，讓標題行取代
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderForeground(lipgloss.Color("#598064")). // 邊框顏色（藍色）
		Width(int(ws.Col) - 2).                      // 方框寬度
		Padding(1).                                  // 方框跟文字間的距離
		Align(lipgloss.Center)                       // 內容居中

	// 渲染內容方框（無頂邊框）
	contentBox := contentStyle.Render(m.textInput.View())

	// 定義邊框樣式
	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#598064"))

	// 定義標題樣式
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#74A382")).
		Bold(true)

	// 計算標題內容長度（標題 + 兩側空格）
	titleContentLen := lipgloss.Width(m.title) + 2 // 注意：這裡用 m.title 的寬度（無樣式）

	// 計算內部可用寬度
	innerWidth := contentStyle.GetWidth()

	// 計算左右水平線的「─」數量（對稱）
	dashTotalLen := innerWidth - titleContentLen

	rightDashLen := dashTotalLen - 3
	leftDashLen := dashTotalLen - rightDashLen

	// 分別渲染左邊、標題和右邊部分
	leftPart := borderStyle.Render("┌" + strings.Repeat("─", leftDashLen))
	titlePart := titleStyle.Render(" " + m.title + " ")
	rightPart := borderStyle.Render(strings.Repeat("─", rightDashLen) + "┐")

	// 水平組合標題行
	titleRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPart, titlePart, rightPart)

	// 垂直組合標題行和內容方框
	fullBox := lipgloss.JoinVertical(lipgloss.Top, titleRow, contentBox)

	// 添加說明文字
	return fmt.Sprintf("%s\n\n 'ctrl + c' Exit", fullBox)
}

func main() {
	// 當前視窗大小初始化
	var err error
	ws, err = unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		fmt.Printf("想辦法找瞜XD 錯誤無法取得終端大小： %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("錯誤:", err)
		os.Exit(1)
	}

	defer func() {
		if logFile != nil {
			if err := logFile.Close(); err != nil {
				log.Printf("關閉日誌檔案時發生錯誤: %v\n", err)
			}
		}
	}()
}
