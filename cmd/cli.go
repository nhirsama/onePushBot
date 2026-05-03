package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/internal/admin"
	"github.com/spf13/viper"
)

func Cli() {
	log.SetOutput(io.MultiWriter(os.Stderr, logBuffer))
	log.SetFlags(log.LstdFlags)

	if err := config.OpenDefaultStore(); err != nil {
		log.Fatalf("打开配置存储失败: %v", err)
	}
	defer func() {
		if err := config.CloseStore(); err != nil {
			log.Printf("关闭配置存储失败: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtime := newRuntimeController()
	if err := runtime.Start(ctx); err != nil {
		log.Fatalf("启动程序失败: %v", err)
	}

	adminServer := admin.New(viper.GetString("admin.addr"), viper.GetString("admin.token"), runtime)
	log.Printf("管理面板: http://%s/?token=%s", adminServer.Addr(), viper.GetString("admin.token"))
	go func() {
		if err := adminServer.Start(); err != nil {
			log.Printf("管理面板退出: %v", err)
		}
	}()

	<-ctx.Done()

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := adminServer.Close(closeCtx); err != nil {
		log.Printf("关闭管理面板失败: %v", err)
	}
	if err := runtime.Close(closeCtx); err != nil {
		log.Printf("关闭应用失败: %v", err)
	}

	log.Println("程序正常结束，正在保存与释放资源")
}
