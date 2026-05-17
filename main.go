package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/robfig/cron/v3"

	"attackEvilPerson/internal/collector"
	"attackEvilPerson/internal/config"
	"attackEvilPerson/internal/message"
	"attackEvilPerson/internal/sender"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: attackEvilPerson <collect|send|cron>")
		os.Exit(1)
	}

	cfgPath := "config.yaml"
	if len(os.Args) >= 3 && os.Args[2] != "" {
		cfgPath = os.Args[2]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cmd := os.Args[1]
	switch cmd {
	case "collect":
		runCollect(cfg)
	case "send":
		runSend(cfg)
	case "cron":
		runCron(cfg)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		fmt.Println("用法: attackEvilPerson <collect|send|cron>")
		os.Exit(1)
	}
}

func runCollect(cfg *config.Config) {
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("创建data目录失败: %v", err)
	}
	keywords, err := collector.LoadKeywords("data")
	if err != nil {
		log.Fatalf("加载关键词失败: %v", err)
	}
	fmt.Printf("已加载 %d 个搜索关键词\n", len(keywords))

	if err := collector.Run(keywords, "api.txt", collector.Config{
		Proxy: cfg.Proxy,
	}); err != nil {
		log.Fatalf("采集失败: %v", err)
	}
}

func runSend(cfg *config.Config) {
	_ = sender.Run("api.txt", sender.Config{
		BaiduURL:    cfg.BaiduURL,
		TelNumber:   cfg.TelNumber,
		TelName:     cfg.TelName,
		Proxy:       cfg.Proxy,
		BrowserPath: cfg.BrowserPath,
		Templates: message.Templates{
			Titles:         cfg.Titles,
			Relatives:      cfg.Relatives,
			Situations:     cfg.Situations,
			ContactMethods: cfg.ContactMethods,
			Greetings:      cfg.Greetings,
		},
	})
}

func runCron(cfg *config.Config) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	c := cron.New()
	c.AddFunc("0 9 * * *", func() {
		fmt.Println("定时任务触发，开始执行发送...")
		runSend(cfg)
	})
	c.Start()

	fmt.Println("cron 模式已启动，每天 9:00 执行发送。按 Ctrl+C 退出。")

	<-sigCh
	fmt.Println("\n正在退出...")
	ctx := c.Stop()
	<-ctx.Done()
	fmt.Println("已退出。")
}
