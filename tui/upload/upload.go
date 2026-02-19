package upload

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"schneider.vip/share/tui"
)

// ProgressReader wraps an io.Reader and reports progress to a Bubble Tea program.
type ProgressReader struct {
	reader  io.Reader
	total   int64
	read    int64
	program *tea.Program
	mu      sync.Mutex
}

type progressMsg struct {
	percent   float64
	bytesRead int64
}

type uploadDoneMsg struct {
	fileID string
	err    error
}

// NewProgressReader creates a reader that sends progress updates to the program.
func NewProgressReader(r io.Reader, size int64, p *tea.Program) *ProgressReader {
	return &ProgressReader{
		reader:  r,
		total:   size,
		program: p,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.mu.Lock()
		pr.read += int64(n)
		pct := float64(pr.read) / float64(pr.total)
		read := pr.read
		pr.mu.Unlock()
		pr.program.Send(progressMsg{percent: pct, bytesRead: read})
	}
	return n, err
}

// Model is the Bubble Tea model for the upload progress screen.
type Model struct {
	progress  progress.Model
	filename  string
	filesize  int64
	percent   float64
	bytesRead int64
	speed     float64 // bytes per second
	lastBytes int64
	lastTime  time.Time
	done      bool
	fileID    string
	err       error
	url       string
	qr        string
}

// NewModel creates a new upload progress model.
func NewModel(filename string, filesize int64) Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	return Model{
		progress: p,
		filename: filename,
		filesize: filesize,
		lastTime: time.Now(),
	}
}

// SetResult sets the final URL and QR code after upload completes.
func (m *Model) SetResult(url, qr string) {
	m.url = url
	m.qr = qr
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case progressMsg:
		m.percent = msg.percent
		if m.percent >= 1.0 {
			m.percent = 1.0
		}
		m.bytesRead = msg.bytesRead

		now := time.Now()
		dt := now.Sub(m.lastTime).Seconds()
		if dt >= 0.5 {
			delta := m.bytesRead - m.lastBytes
			m.speed = float64(delta) / dt
			m.lastBytes = m.bytesRead
			m.lastTime = now
		}

		return m, m.progress.SetPercent(m.percent)
	case uploadDoneMsg:
		m.done = true
		m.fileID = msg.fileID
		m.err = msg.err
		return m, tea.Quit
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	if m.done {
		if m.err != nil {
			b.WriteString(tui.Error.Render("Upload failed: " + m.err.Error()))
			b.WriteString("\n")
		} else {
			b.WriteString(tui.Success.Render("Upload complete!"))
			b.WriteString("\n")
			if m.qr != "" {
				b.WriteString(m.qr)
			}
			if m.url != "" {
				fmt.Fprintf(&b, "\n%s %s\n", tui.Title.Render("URL:"), m.url)
			}
		}
		return b.String()
	}

	b.WriteString(tui.Title.Render("Uploading"))
	b.WriteString(" ")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.filename))
	b.WriteString("\n\n")
	b.WriteString(m.progress.View())
	fmt.Fprintf(&b, "\n%s / %s",
		humanBytes(m.bytesRead),
		humanBytes(m.filesize))
	if m.speed > 0 {
		fmt.Fprintf(&b, "  %s/s", humanBytes(int64(m.speed)))
	}
	b.WriteString("\n")

	return b.String()
}

// SendDone sends an uploadDoneMsg to the program.
func SendDone(p *tea.Program, fileID string, err error) {
	p.Send(uploadDoneMsg{fileID: fileID, err: err})
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
