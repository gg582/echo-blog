package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/server"
)

func Execute() {
	root := &cobra.Command{
		Use:   "echo-blog",
		Short: "Run the Echo-based personal blog backend",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd.Context()); err != nil {
				log.Println(err)
				os.Exit(1)
			}
		},
	}
	root.AddCommand(newInitCommand())

	// SIGINT/SIGTERM (e.g. systemctl stop) cancel the context and trigger a
	// graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := root.ExecuteContext(ctx); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	app, err := server.NewApp(config.Load())
	if err != nil {
		return err
	}
	defer app.Close()
	log.Println("Database loaded.")

	return app.Run(ctx)
}
