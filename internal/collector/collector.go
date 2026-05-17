package collector

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	workers  = 10
	maxPage  = 2
	baiduURL = "https://www.baidu.com/s"
)

var adPattern = regexp.MustCompile(`https?://ada\.baidu\.com/site/[\w.-]+/xyl\?imid=[\w-]+`)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
}

type Config struct {
	Proxy string
}

func LoadKeywords(dataDir string) ([]string, error) {
	cityPath := filepath.Join(dataDir, "kw_city.txt")
	hospitalPath := filepath.Join(dataDir, "kw_hospital.txt")

	cities, err := readLines(cityPath)
	if err != nil {
		return nil, fmt.Errorf("read cities: %w", err)
	}
	hospitals, err := readLines(hospitalPath)
	if err != nil {
		return nil, fmt.Errorf("read hospitals: %w", err)
	}

	keywords := make([]string, 0, len(cities)*len(hospitals))
	for _, city := range cities {
		for _, hospital := range hospitals {
			keywords = append(keywords, city+hospital)
		}
	}
	return keywords, nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func Run(keywords []string, outputPath string, cfg Config) error {
	jobs := make(chan string, len(keywords))
	results := make(chan []string, len(keywords))

	var wg sync.WaitGroup
	var client *http.Client

	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return fmt.Errorf("parse proxy: %w", err)
		}
		client = &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
			Timeout:   30 * time.Second,
		}
	} else {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(&wg, client, jobs, results)
	}

	for _, kw := range keywords {
		jobs <- kw
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	seen := make(map[string]bool)
	for urls := range results {
		for _, u := range urls {
			seen[u] = true
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	for u := range seen {
		fmt.Fprintln(f, u)
	}
	fmt.Printf("完成去重，共 %d 条唯一 URL 写入 %s\n", len(seen), outputPath)
	return nil
}

func worker(wg *sync.WaitGroup, client *http.Client, jobs <-chan string, results chan<- []string) {
	defer wg.Done()
	for keyword := range jobs {
		urls, err := fetch(client, keyword)
		if err != nil {
			fmt.Printf("关键字 \"%s\" 获取失败: %v\n", keyword, err)
			continue
		}
		if len(urls) > 0 {
			fmt.Printf("成功提取 %d 条 url: %s\n", len(urls), strings.Join(urls, "  "))
			results <- urls
		} else {
			fmt.Printf("当前关键字: %s \t 未查询到匹配结果\n", keyword)
		}
	}
}

func fetch(client *http.Client, keyword string) ([]string, error) {
	delay := time.Duration(500+rand.Intn(2000)) * time.Millisecond
	fmt.Printf("当前关键字: %s \t %.1f秒后开始提取\n", keyword, delay.Seconds())
	time.Sleep(delay)

	var allResults []string
	for page := 0; page < maxPage; page++ {
		matches, err := fetchPage(client, keyword, page)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, matches...)
	}
	if len(allResults) > 0 {
		return dedupe(allResults), nil
	}
	return nil, nil
}

func fetchPage(client *http.Client, keyword string, page int) ([]string, error) {
	reqURL := fmt.Sprintf("%s?wd=%s&pn=%d", baiduURL, url.QueryEscape(keyword), page*10)
	fmt.Println(reqURL)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return adPattern.FindAllString(string(body), -1), nil
}

func dedupe(s []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
