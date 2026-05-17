package sender

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"attackEvilPerson/internal/message"
)

type Config struct {
	BaiduURL    string
	TelNumber   string
	TelName     string
	Proxy       string
	BrowserPath string
	Templates   message.Templates
}

func Run(apiPath string, cfg Config) error {
	urls, err := readURLs(apiPath)
	if err != nil {
		return fmt.Errorf("read api.txt: %w", err)
	}
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})
	total := len(urls)
	success := 0

	browser, err := launchBrowser(cfg)
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: cfg.BaiduURL})
	if err != nil {
		return fmt.Errorf("open baidu: %w", err)
	}
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("wait baidu load: %w", err)
	}

	for i, u := range urls {
		fmt.Printf("%d\t", i+1)
		ok := processURL(page, u, cfg)
		if ok {
			success++
		}
	}
	fmt.Printf("完成！成功: %d/%d\n", success, total)
	return nil
}

func launchBrowser(cfg Config) (*rod.Browser, error) {
	l := launcher.New()
	if cfg.BrowserPath != "" {
		l = l.Bin(cfg.BrowserPath)
	}
	if cfg.Proxy != "" {
		l = l.Proxy(cfg.Proxy)
	}
	l = l.Headless(false).NoSandbox(true)
	controlURL, err := l.Launch()
	if err != nil {
		fmt.Printf("启动浏览器失败: %v，尝试自动发现浏览器...\n", err)
		l = launcher.New().Headless(false).NoSandbox(true)
		if cfg.Proxy != "" {
			l = l.Proxy(cfg.Proxy)
		}
		controlURL, err = l.Launch()
		if err != nil {
			return nil, fmt.Errorf("启动浏览器失败: %w", err)
		}
	}
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect browser: %w", err)
	}
	return browser, nil
}

func processURL(page *rod.Page, u string, cfg Config) bool {
	if err := page.Navigate(u); err != nil {
		fmt.Printf("导航失败: %s, err=%v\n", u, err)
		return false
	}
	if err := page.WaitLoad(); err != nil {
		fmt.Printf("等待加载失败: %s, err=%v\n", u, err)
		return false
	}

	el, err := page.Element(".imlp-component-typebox-input")
	if err != nil || el == nil {
		fmt.Printf("未找到输入框: %s\n", u)
		return false
	}

	msg := message.Build(cfg.Templates, cfg.TelNumber, cfg.TelName)
	fmt.Println(msg)
	if err := el.Input(msg); err != nil {
		fmt.Printf("输入留言失败: %v\n", err)
		return false
	}

	sendBtn, err := page.Element(".imlp-component-typebox-send-btn")
	if err != nil || sendBtn == nil {
		fmt.Printf("未找到发送按钮: %s\n", u)
		return false
	}

	if _, err := sendBtn.Eval("() => this.click()"); err != nil {
		fmt.Printf("点击发送失败: %v\n", err)
		return false
	}

	time.Sleep(500 * time.Millisecond)
	return true
}

func readURLs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}
	return urls, scanner.Err()
}
