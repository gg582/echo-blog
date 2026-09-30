package cmd

import (
	"log"
	"os"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/server"
	"github.com/spf13/cobra"
)

func Execute() {
	root := &cobra.Command{
		Use:   "chi-blog",
		Short: "Run the chi-based personal blog backend",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			if err := server.Serve(cfg, server.NewRouter(cfg)); err != nil {
				log.Println(err)
				os.Exit(1)
			}
		},
	}
	root.AddCommand(newInitCommand())

	if err := root.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
