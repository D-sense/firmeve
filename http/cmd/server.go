package cmd

import (
	"context"
	"fmt"
	http2 "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/firmeve/firmeve/bootstrap"
	"github.com/firmeve/firmeve/config"
	"github.com/firmeve/firmeve/http"
	logging "github.com/firmeve/firmeve/logger"
	"github.com/spf13/cobra"
)

type cmd struct {
	bootstrap *bootstrap.Bootstrap
	command   *cobra.Command
	logger    logging.Loggable
	config    *config.Config
}

func NewServer(bootstrap *bootstrap.Bootstrap) *cmd {
	return &cmd{
		bootstrap: bootstrap,
		command:   new(cobra.Command),
	}
}

func (c *cmd) Cmd() *cobra.Command {
	c.command.Use = "http:serve"
	c.command.Flags().StringP("host", "H", ":80", "Http serve address")
	c.command.Run = func(cmd *cobra.Command, args []string) {
		c.run(cmd, args)
	}

	return c.command
}

func (c *cmd) run(cmd *cobra.Command, args []string) {
	// bootstrap
	c.bootstrap.Boot()

	// base
	c.config = c.bootstrap.Firmeve.Get(`config`).(*config.Config)
	c.logger = c.bootstrap.Firmeve.Get(`logger`).(logging.Loggable)

	srv := &http2.Server{
		Addr:    cmd.Flag("host").Value.String(),
		Handler: c.bootstrap.Firmeve.Get(`http.router`).(*http.Router),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http2.ErrServerClosed {
			c.logger.Fatal(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	//log.Println("Shutdown Server ...")
	c.logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		c.logger.Fatal("Server Shutdown: ", err)
	}

	c.logger.Info("Server exiting")
}
