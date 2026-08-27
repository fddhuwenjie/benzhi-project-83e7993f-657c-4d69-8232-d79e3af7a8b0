package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sherd-proof/internal/app"
	"sherd-proof/internal/store"
	webapi "sherd-proof/internal/web"
)

func main() {
	configuration := config{}
	flag.StringVar(&configuration.address, "addr", addressDefault(), "HTTP 监听地址")
	flag.StringVar(&configuration.database, "data", "data/sherd-proof.db", "SQLite 数据库路径")
	flag.BoolVar(&configuration.selftest, "selftest", false, "执行真实 HTTP 业务自检后退出")
	flag.Parse()
	if err := run(configuration); err != nil {
		log.Printf("服务退出: %v", err)
		os.Exit(1)
	}
}

func run(configuration config) error {
	if err := validateAddress(configuration.address, configuration.selftest); err != nil {
		return err
	}
	if configuration.selftest {
		return runSelftest(configuration.address)
	}
	if err := ensureDatabaseDirectory(configuration.database); err != nil {
		return fmt.Errorf("创建数据目录: %w", err)
	}
	repository, err := store.Open(configuration.database)
	if err != nil {
		return fmt.Errorf("打开存储: %w", err)
	}
	defer repository.Close()
	service := app.NewService(repository)
	server := newHTTPServer(configuration.address, webapi.New(service))
	listener, err := net.Listen("tcp", configuration.address)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", configuration.address, err)
	}
	log.Printf("sherd-proof 已监听 http://%s", listener.Addr())
	errChannel := make(chan error, 1)
	go func() { errChannel <- server.Serve(listener) }()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	select {
	case signal := <-signals:
		log.Printf("收到信号 %s，开始关闭", signal)
	case serveErr := <-errChannel:
		if serveErr != nil && serveErr != http.ErrServerClosed {
			return serveErr
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("优雅关闭: %w", err)
	}
	select {
	case serveErr := <-errChannel:
		if serveErr != nil && serveErr != http.ErrServerClosed {
			return serveErr
		}
	case <-time.After(time.Second):
	}
	return nil
}

func newHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second,
		WriteTimeout: 25 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
}
