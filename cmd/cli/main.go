// cmd/cli/main.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"anvil-dns/internal/cloudflare"
)

var (
	targetIP string
)

func initConfig() {
	// Search for config in home directory with name ".anvil-dns" (without extension).
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".anvil-dns")

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}

func main() {
	cobra.OnInitialize(initConfig)

	var rootCmd = &cobra.Command{
		Use:   "anvil-dns",
		Short: "CLI to manage anvilcomputing.com DNS records",
	}

	// Subcommand: check (Same as before)
	var checkCmd = &cobra.Command{
		Use:   "check [username]",
		Short: "Check if a DNS record exists for a username",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]
			zoneName := "anvilcomputing.com"
			recordName := fmt.Sprintf("%s.%s", username, zoneName)

			client, err := cloudflare.NewClient(os.Getenv("CLOUDFLARE_API_TOKEN"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			exists, ip, err := client.CheckRecord(context.Background(), zoneName, recordName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if exists {
				fmt.Printf("❌ Record '%s' already exists -> %s\n", recordName, ip)
			} else {
				fmt.Printf("✅ Record '%s' is available!\n", recordName)
			}
		},
	}

	// Subcommand: create
	var createCmd = &cobra.Command{
		Use:   "create [username]",
		Short: "Create a new DNS record for a username",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]
			zoneName := "anvilcomputing.com"
			recordName := fmt.Sprintf("%s.%s", username, zoneName)

			// 1. Determine the target IP using Viper
			ipToUse := viper.GetString("TARGET_IP")
			if ipToUse == "" {
				fmt.Println("Error: Target IP is required. Provide it via --target-ip, TARGET_IP env var, or save it in ~/.anvil-dns.yaml")
				os.Exit(1)
			}

			client, err := cloudflare.NewClient(os.Getenv("CLOUDFLARE_API_TOKEN"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()

			// 2. Check if it already exists to prevent overwriting
			fmt.Printf("Checking availability for '%s'...\n", recordName)
			exists, existingIP, err := client.CheckRecord(ctx, zoneName, recordName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking record: %v\n", err)
				os.Exit(1)
			}
			if exists {
				fmt.Printf("❌ Cannot create: Record already exists pointing to %s\n", existingIP)
				os.Exit(1)
			}

			// 3. Create the record
			fmt.Printf("Provisioning record -> %s\n", ipToUse)
			err = client.CreateRecord(ctx, zoneName, recordName, ipToUse)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to create record: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Successfully created %s pointing to %s!\n", recordName, ipToUse)

			// 4. Save the IP to config so we don't have to type it next time
			viper.Set("TARGET_IP", ipToUse)
			if err := viper.WriteConfigAs(filepath.Join(os.Getenv("HOME"), ".anvil-dns.yaml")); err == nil {
				fmt.Println("(Saved target IP as default in ~/.anvil-dns.yaml)")
			}
		},
	}

	// Wire up the flags for the create command
	createCmd.Flags().StringVar(&targetIP, "target-ip", "", "The IP address to point the DNS record to")
	viper.BindPFlag("TARGET_IP", createCmd.Flags().Lookup("target-ip"))

	// Subcommand: list
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List and page through all DNS records in the zone",
		Run: func(cmd *cobra.Command, args []string) {
			zoneName := "anvilcomputing.com"

			client, err := cloudflare.NewClient(os.Getenv("CLOUDFLARE_API_TOKEN"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			page := 1
			perPage := 20 // Show 20 records at a time
			reader := bufio.NewReader(os.Stdin)

			for {
				records, resultInfo, err := client.ListRecords(ctx, zoneName, page, perPage)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching records: %v\n", err)
					os.Exit(1)
				}

				if len(records) == 0 && page == 1 {
					fmt.Println("No DNS records found in this zone.")
					return
				}

				// Setup a tabwriter to create nicely aligned columns
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
				fmt.Fprintln(w, "NAME\tTYPE\tCONTENT\tPROXIED\t")
				fmt.Fprintln(w, "----\t----\t-------\t-------\t")

				for _, r := range records {
					proxied := "No"
					if r.Proxied != nil && *r.Proxied {
						proxied = "Yes"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", r.Name, r.Type, r.Content, proxied)
				}
				w.Flush()

				// Pagination logic
				if page >= resultInfo.TotalPages {
					fmt.Printf("\n--- End of records (Total: %d) ---\n", resultInfo.Total)
					break
				}

				fmt.Printf("\nPage %d of %d. Press [Enter] for next page, or 'q' to quit: ", page, resultInfo.TotalPages)

				input, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(input)) == "q" {
					break
				}

				page++
			}
		},
	}

	// Subcommand: delete
	var deleteCmd = &cobra.Command{
		Use:   "delete [username]",
		Short: "Delete a DNS record for a username",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]
			zoneName := "anvilcomputing.com"
			recordName := fmt.Sprintf("%s.%s", username, zoneName)

			client, err := cloudflare.NewClient(os.Getenv("CLOUDFLARE_API_TOKEN"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()

			// 1. Verify the record exists before prompting
			fmt.Printf("Locating record '%s'...\n", recordName)
			exists, ip, err := client.CheckRecord(ctx, zoneName, recordName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if !exists {
				fmt.Printf("❌ Record '%s' does not exist.\n", recordName)
				os.Exit(1)
			}

			// 2. Prompt for confirmation
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("⚠️  Are you sure you want to delete '%s' pointing to %s? [y/N]: ", recordName, ip)
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))

			if confirm != "y" && confirm != "yes" {
				fmt.Println("Deletion cancelled.")
				return
			}

			// 3. Delete the record
			fmt.Println("Deleting record from Cloudflare...")
			err = client.DeleteRecord(ctx, zoneName, recordName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to delete record: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Successfully deleted %s!\n", recordName)
		},
	}


	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
