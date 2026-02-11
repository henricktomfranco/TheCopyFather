package windows

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// ClipboardManager handles clipboard operations
type ClipboardManager struct {
	lastText   string
	monitoring bool
	stopChan   chan bool
	callback   func(string)
	mu         sync.RWMutex // Protects lastText and monitoring
}

// NewClipboardManager creates a new clipboard manager
func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{
		stopChan: make(chan bool),
	}
}

// GetText retrieves text from the clipboard
func (c *ClipboardManager) GetText() (string, error) {
	ret, _, err := openClipboard.Call(0)
	if ret == 0 {
		return "", fmt.Errorf("failed to open clipboard: %v", err)
	}
	defer closeClipboard.Call()

	ret, _, _ = getClipboardData.Call(uintptr(CF_UNICODETEXT))
	if ret == 0 {
		return "", fmt.Errorf("no text in clipboard")
	}

	handle := ret

	ret, _, _ = globalLock.Call(handle)
	if ret == 0 {
		return "", fmt.Errorf("failed to lock clipboard data")
	}
	defer globalUnlock.Call(handle)

	ptr := (*uint16)(unsafe.Pointer(ret))

	// Calculate length
	length, _, _ := lstrlenW.Call(ret)

	return syscall.UTF16ToString((*[1 << 30]uint16)(unsafe.Pointer(ptr))[:length:length]), nil
}

// SetText sets text in the clipboard
func (c *ClipboardManager) SetText(text string) error {
	ret, _, err := openClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard: %v", err)
	}
	defer closeClipboard.Call()

	emptyClipboard.Call()

	// Convert string to UTF16
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return fmt.Errorf("failed to convert string: %v", err)
	}

	// Allocate memory
	size := uintptr(len(utf16) * 2)
	ret, _, _ = globalAlloc.Call(uintptr(GMEM_MOVEABLE), size)
	if ret == 0 {
		return fmt.Errorf("failed to allocate memory")
	}

	handle := ret

	// Lock and copy
	ret, _, _ = globalLock.Call(handle)
	if ret == 0 {
		globalFree.Call(handle)
		return fmt.Errorf("failed to lock memory")
	}

	// Copy data
	dest := (*[1 << 30]uint16)(unsafe.Pointer(ret))
	src := utf16
	for i := 0; i < len(src); i++ {
		dest[i] = src[i]
	}

	globalUnlock.Call(handle)

	// Set clipboard data
	ret, _, _ = setClipboardData.Call(uintptr(CF_UNICODETEXT), handle)
	if ret == 0 {
		globalFree.Call(handle)
		return fmt.Errorf("failed to set clipboard data")
	}

	return nil
}

// Start begins monitoring the clipboard for changes
func (c *ClipboardManager) Start(callback func(string)) {
	c.callback = callback
	c.monitoring = true

	go c.monitor()
}

// Stop stops monitoring the clipboard
func (c *ClipboardManager) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.monitoring {
		return
	}

	c.monitoring = false
	select {
	case <-c.stopChan:
		// Already closed
	default:
		close(c.stopChan)
	}
}

func (c *ClipboardManager) monitor() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		c.mu.RLock()
		monitoring := c.monitoring
		c.mu.RUnlock()

		if !monitoring {
			return
		}

		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			text, err := c.GetText()
			if err != nil {
				continue
			}

			c.mu.Lock()
			if text != c.lastText && text != "" {
				c.lastText = text
				callback := c.callback
				c.mu.Unlock()
				if callback != nil {
					go callback(text)
				}
			} else {
				c.mu.Unlock()
			}
		}
	}
}

// SimulateCopy simulates a Ctrl+C copy operation to get selected text
func (c *ClipboardManager) SimulateCopy() error {
	// Store current clipboard
	oldText, _ := c.GetText()

	// Simulate Ctrl+C would go here using SendInput
	// For now, we assume text is already copied

	// Small delay to let the copy operation complete
	time.Sleep(100 * time.Millisecond)

	// Get new clipboard content
	newText, err := c.GetText()
	if err != nil {
		// Restore old clipboard
		c.SetText(oldText)
		return err
	}

	// Restore old clipboard
	c.SetText(oldText)

	if newText != "" {
		c.lastText = newText
		if c.callback != nil {
			c.callback(newText)
		}
	}

	return nil
}

// markdownToHTML converts markdown text to HTML
func markdownToHTML(text string) string {
	// Convert bold text: **text** -> <strong>text</strong>
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = boldRegex.ReplaceAllString(text, "<strong>$1</strong>")

	// Convert italic text: *text* -> <em>text</em>
	italicRegex := regexp.MustCompile(`\*([^\s*][^*]*[^\s*])\*`)
	text = italicRegex.ReplaceAllString(text, "<em>$1</em>")

	// Convert bullet lists
	lines := strings.Split(text, "\n")
	var result []string
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if line is a bullet
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "• ") {
			if !inList {
				result = append(result, "<ul>")
				inList = true
			}
			// Remove bullet marker and wrap in <li>
			content := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* "), "• ")
			result = append(result, fmt.Sprintf("<li>%s</li>", content))
		} else if inList && trimmed == "" {
			// Empty line ends the list
			result = append(result, "</ul>")
			inList = false
		} else {
			if inList {
				result = append(result, "</ul>")
				inList = false
			}
			// Convert newlines to <br> or <p> tags
			if trimmed != "" {
				result = append(result, fmt.Sprintf("<p>%s</p>", line))
			} else {
				result = append(result, "<br>")
			}
		}
	}

	if inList {
		result = append(result, "</ul>")
	}

	html := strings.Join(result, "")
	// Clean up double <br> tags
	html = strings.ReplaceAll(html, "<br><br>", "</p><p>")
	html = strings.ReplaceAll(html, "<p></p>", "")

	return html
}

// SetRichText sets both plain text and HTML to the clipboard
// If htmlText is the same as plainText, it will auto-convert markdown to HTML
func (c *ClipboardManager) SetRichText(plainText, htmlText string) error {
	// If htmlText is the same as plainText, convert markdown to HTML
	if plainText == htmlText {
		htmlText = markdownToHTML(plainText)
	}
	ret, _, err := openClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard: %v", err)
	}
	defer closeClipboard.Call()

	emptyClipboard.Call()

	// Set plain text (Unicode)
	utf16, err := syscall.UTF16FromString(plainText)
	if err != nil {
		return fmt.Errorf("failed to convert string: %v", err)
	}

	size := uintptr(len(utf16) * 2)
	ret, _, _ = globalAlloc.Call(uintptr(GMEM_MOVEABLE), size)
	if ret == 0 {
		return fmt.Errorf("failed to allocate memory for text")
	}
	textHandle := ret

	ret, _, _ = globalLock.Call(textHandle)
	if ret == 0 {
		globalFree.Call(textHandle)
		return fmt.Errorf("failed to lock memory")
	}

	dest := (*[1 << 30]uint16)(unsafe.Pointer(ret))
	for i := 0; i < len(utf16); i++ {
		dest[i] = utf16[i]
	}
	globalUnlock.Call(textHandle)

	// Register HTML format
	cfHTML, _, _ := registerClipboardFormatW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("HTML Format"))))

	if cfHTML != 0 {
		// Create HTML clipboard format
		htmlWithHeader := createHTMLClipboardFormat(htmlText)
		htmlUTF8 := []byte(htmlWithHeader)

		htmlSize := uintptr(len(htmlUTF8))
		ret, _, _ = globalAlloc.Call(uintptr(GMEM_MOVEABLE), htmlSize+1)
		if ret == 0 {
			globalFree.Call(textHandle)
			return fmt.Errorf("failed to allocate memory for HTML")
		}
		htmlHandle := ret

		ret, _, _ = globalLock.Call(htmlHandle)
		if ret == 0 {
			globalFree.Call(htmlHandle)
			globalFree.Call(textHandle)
			return fmt.Errorf("failed to lock HTML memory")
		}

		htmlDest := (*[1 << 30]byte)(unsafe.Pointer(ret))
		for i := 0; i < len(htmlUTF8); i++ {
			htmlDest[i] = htmlUTF8[i]
		}
		htmlDest[len(htmlUTF8)] = 0
		globalUnlock.Call(htmlHandle)

		// Set HTML format
		ret, _, _ = setClipboardData.Call(cfHTML, htmlHandle)
		if ret == 0 {
			globalFree.Call(htmlHandle)
		}
	}

	// Set plain text
	ret, _, _ = setClipboardData.Call(uintptr(CF_UNICODETEXT), textHandle)
	if ret == 0 {
		globalFree.Call(textHandle)
		return fmt.Errorf("failed to set clipboard data")
	}

	return nil
}

// createHTMLClipboardFormat creates the HTML clipboard format header
func createHTMLClipboardFormat(html string) string {
	htmlFragment := fmt.Sprintf("<html><body>%s</body></html>", html)
	startHTML := 0
	endHTML := len(htmlFragment) + 105 // Offset for header
	startFragment := 105
	endFragment := startFragment + len(htmlFragment)

	return fmt.Sprintf("Version:0.9\r\nStartHTML:%d\r\nEndHTML:%d\r\nStartFragment:%d\r\nEndFragment:%d\r\n<%s",
		startHTML, endHTML, startFragment, endFragment, htmlFragment)
}
