package repl

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type History struct {
	entries  []string
	filePath string
}

func NewHistory() *History {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".es-cli")
	os.MkdirAll(dir, 0755)

	h := &History{
		filePath: filepath.Join(dir, "history"),
	}
	h.load()
	return h
}

func (h *History) Entries() []string {
	return h.entries
}

func (h *History) Add(entry string) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		return
	}
	h.entries = append(h.entries, entry)
	h.save(entry)
}

func (h *History) load() {
	f, err := os.Open(h.filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
}

func (h *History) save(entry string) {
	f, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(entry + "\n")
}
