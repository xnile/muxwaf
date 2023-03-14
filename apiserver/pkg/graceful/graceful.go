package graceful

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"
)

// gracefulStop 优雅退出
// 等待中断信号以超时 5 秒正常关闭服务器
// 官方说明：https://github.com/gin-gonic/gin#graceful-restart-or-stop
func Stop(srv *http.Server) {
	quit := make(chan os.Signal)
	// kill 命令发送信号 syscall.SIGTERM
	// kill -2 命令发送信号 syscall.SIGINT
	// kill -9 命令发送信号 syscall.SIGKILL
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// 5 秒后捕获 ctx.Done() 信号
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	default:
	}
	log.Println("Server exiting")
}