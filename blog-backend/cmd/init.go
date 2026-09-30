package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gg582/echo-blog/blog-backend/config"
	"github.com/gg582/echo-blog/blog-backend/database"
	"github.com/gg582/echo-blog/blog-backend/utils"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Register the blog admin account",
		Long:  "Register the blog admin account. Requires sqlite3.",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("WARNING: you will add administrator into SQLite. type 'yes' to continue")
			var answer string
			fmt.Scanln(&answer)
			if strings.ToLower(answer) != "yes" {
				log.Println("Quitting without registration...")
				os.Exit(1)
			}

			database.InitDatabase(config.Load().DBPath)

			var count int
			err := database.DB.QueryRow("SELECT COUNT(*) FROM blog_users").Scan(&count)
			if err != nil {
				log.Fatalf("Failed to query users: %v", err)
			}
			if count > 0 {
				log.Println("You can enter only one admin account for this blog.")
				os.Exit(1)
			}

			log.Println("Please enter your username.")
			var username string
			fmt.Scanln(&username)
			log.Println("Please enter your password.")
			var password string
			fmt.Scanln(&password)

			pwHash, err := utils.HashPassword(password)
			if err != nil {
				log.Fatalf("Failed to generate password hash: %v", err)
			}
			if _, err := database.DB.Exec("INSERT INTO blog_users (username, password_hash) values (?,?)", username, pwHash); err != nil {
				log.Fatalf("Failed to insert user info to Database. Please check sqlite3's condition: %v", err)
			}
			log.Printf("Admin user (%v) created", username)
		},
	}
}
