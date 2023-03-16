package main

import (
	"github.com/labstack/gommon/color"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xnile/muxwaf/internal/config"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/graceful"
	"github.com/xnile/muxwaf/router"
	"log"
	"net/http"
	"time"
)

const (
	banner = `
   _____                 __      __         _____ 
  /     \  __ _____  ___/  \    /  \_____ _/ ____\
 /  \ /  \|  |  \  \/  /\   \/\/   /\__  \\   __\ 
/    Y    \  |  />    <  \        /  / __ \|  |   
\____|__  /____//__/\_ \  \__/\  /  (____  /__|   
        \/            \/       \/        \/
`
)

var cfg = pflag.StringP("config", "c", "", "muxwaf config file path")

func main() {
	pflag.Parse()
	// init config
	if err := config.Init(*cfg); err != nil {
		panic(err)
	}

	host := viper.GetString("host")
	port := viper.GetString("port")

	service.Setup()
	app := router.Init()
	srv := &http.Server{
		Addr:           host + ":" + port,
		Handler:        app,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to listen: ", err.Error())
		}
	}()
	color.Println(color.Green(banner))
	graceful.Stop(srv)
}
