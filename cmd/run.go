package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nhirsama/onePushBot/config"
	appruntime "github.com/nhirsama/onePushBot/internal/runtime"
)

type process interface {
	Start(context.Context) error
	Close(context.Context) error
}

func Run() {
	log.SetOutput(io.MultiWriter(os.Stderr, appruntime.LogBuffer()))
	log.SetFlags(log.LstdFlags)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, config.OpenStore, config.CloseStore, func() process {
		return appruntime.NewManager()
	}); err != nil {
		log.Fatalf("%v", err)
	}
}

func run(ctx context.Context, openStore func() error, closeStore func() error, newProcess func() process) error {
	if err := openStore(); err != nil {
		return fmt.Errorf("打开配置存储失败: %w", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("关闭配置存储失败: %v", err)
		}
	}()

	proc := newProcess()
	if err := proc.Start(ctx); err != nil {
		return fmt.Errorf("启动程序失败: %w", err)
	}

	<-ctx.Done()

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := proc.Close(closeCtx); err != nil {
		log.Printf("关闭应用失败: %v", err)
	}

	log.Println("程序正常结束，正在保存与释放资源")
	return nil
}
