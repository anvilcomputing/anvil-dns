// cmd/web/main.go
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"anvil-dns/internal/cloudflare"
)

type PageData struct {
	Records  []RecordView
	TargetIP string
	Error    string
	Success  string
}

type RecordView struct {
	Name    string
	Type    string
	Content string
	Proxied string
}

func main() {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	cfClient, err := cloudflare.NewClient(token)
	if err != nil {
		log.Fatalf("Error initializing Cloudflare client: %v", err)
	}

	defaultIP := os.Getenv("TARGET_IP")
	if defaultIP == "" {
		defaultIP = "204.152.223.10"
	}

	tmpl := template.Must(template.ParseFiles("cmd/web/index.html"))

	// Route: GET / (Dashboard)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()
		zoneName := "anvilcomputing.com"

		records, _, err := cfClient.ListRecords(ctx, zoneName, 1, 100)
		if err != nil {
			http.Error(w, "Failed to fetch records", http.StatusInternalServerError)
			return
		}

		var recordViews []RecordView
		for _, rec := range records {
			proxied := "No"
			if rec.Proxied != nil && *rec.Proxied {
				proxied = "Yes"
			}
			recordViews = append(recordViews, RecordView{
				Name:    rec.Name,
				Type:    rec.Type,
				Content: rec.Content,
				Proxied: proxied,
			})
		}

		data := PageData{
			Records:  recordViews,
			TargetIP: defaultIP,
			// Read feedback messages from the URL parameters
			Error:    r.URL.Query().Get("error"),
			Success:  r.URL.Query().Get("success"),
		}

		tmpl.Execute(w, data)
	})

	// Route: POST /create
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		username := r.FormValue("username")
		targetIP := r.FormValue("target_ip")
		zoneName := "anvilcomputing.com"
		recordName := fmt.Sprintf("%s.%s", username, zoneName)

		ctx := context.Background()
		
		// Prevent overwriting
		exists, _, _ := cfClient.CheckRecord(ctx, zoneName, recordName)
		if exists {
			http.Redirect(w, r, "/?error=Record already exists", http.StatusSeeOther)
			return
		}

		err := cfClient.CreateRecord(ctx, zoneName, recordName, targetIP)
		if err != nil {
			http.Redirect(w, r, "/?error=Failed to create record", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/?success=Record created successfully", http.StatusSeeOther)
	})

	// Route: POST /delete
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		recordName := r.FormValue("recordName")
		zoneName := "anvilcomputing.com"

		ctx := context.Background()
		err := cfClient.DeleteRecord(ctx, zoneName, recordName)
		if err != nil {
			http.Redirect(w, r, "/?error=Failed to delete record", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/?success=Record deleted successfully", http.StatusSeeOther)
	})

	port := "8081"
	fmt.Printf("Starting Anvil DNS Web Admin on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
