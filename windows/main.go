// DWG Viewer за Windows: малък локален сървър + прозорец на Edge/Chrome в "app" режим.
// Работи изцяло офлайн — всичко е вградено в този .exe, файловете не напускат компютъра.
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"
)

//go:embed web
var webFS embed.FS

const port = 47831 // фиксиран порт => историята на файловете се пази между стартиранията

var lastHit atomic.Int64

func main() {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		os.Exit(1)
	}
	files := http.FileServer(http.FS(sub))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastHit.Store(time.Now().Unix())
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	})
	url := fmt.Sprintf("http://127.0.0.1:%d/", port)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		// вероятно вече е стартиран — отваряме още един прозорец към него
		openWindow(url)
		return
	}
	lastHit.Store(time.Now().Unix())
	go http.Serve(ln, handler)

	if runtime.GOOS != "windows" {
		fmt.Println("DWG Viewer:", url)
		select {}
	}

	started := time.Now()
	cmd := openWindow(url)
	if cmd != nil {
		cmd.Wait()
		if time.Since(started) > 5*time.Second {
			return // прозорецът е затворен — спираме сървъра
		}
	}
	// браузърът е предал прозореца на друг процес: чакаме докато няма заявки 30 мин.
	for time.Since(time.Unix(lastHit.Load(), 0)) < 30*time.Minute {
		time.Sleep(10 * time.Second)
	}
}

func openWindow(url string) *exec.Cmd {
	if runtime.GOOS != "windows" {
		return nil
	}
	profile := filepath.Join(firstEnv("LOCALAPPDATA", "APPDATA"), "DWGViewer", "profile")
	os.MkdirAll(profile, 0o755)
	for _, b := range browsers() {
		if _, err := os.Stat(b); err != nil {
			continue
		}
		cmd := exec.Command(b, "--app="+url, "--user-data-dir="+profile,
			"--window-size=1200,820", "--no-first-run", "--no-default-browser-check")
		if cmd.Start() == nil {
			return cmd
		}
	}
	// без Edge/Chrome: обикновения браузър по подразбиране
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	return nil
}

func browsers() []string {
	pf, pf86, local := os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LOCALAPPDATA")
	var out []string
	for _, base := range []string{pf86, pf} {
		if base != "" {
			out = append(out, filepath.Join(base, `Microsoft\Edge\Application\msedge.exe`))
		}
	}
	for _, base := range []string{pf, pf86, local} {
		if base != "" {
			out = append(out, filepath.Join(base, `Google\Chrome\Application\chrome.exe`))
		}
	}
	return out
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return os.TempDir()
}
