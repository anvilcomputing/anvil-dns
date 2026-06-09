// internal/cloudflare/dns.go
package cloudflare

import (
	"context"
	"errors"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go"
)

// Client wraps the Cloudflare API client
type Client struct {
	api *cf.API
}

// NewClient creates a new Cloudflare client using an API token
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("CLOUDFLARE_API_TOKEN is required")
	}
	
	api, err := cf.NewWithAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudflare client: %w", err)
	}
	
	return &Client{api: api}, nil
}

// CheckRecord queries Cloudflare to see if an A record exists for the given name.
// Returns: exists (bool), targetIP (string), error
func (c *Client) CheckRecord(ctx context.Context, zoneName, recordName string) (bool, string, error) {
	// 1. Resolve the Zone ID for "anvilcomputing.com"
	zoneID, err := c.api.ZoneIDByName(zoneName)
	if err != nil {
		return false, "", fmt.Errorf("failed to find zone '%s': %w", zoneName, err)
	}

	// 2. Query for the specific DNS record (filtering by A records)
	params := cf.ListDNSRecordsParams{
		Name: recordName,
		Type: "A", 
	}
	
	// cf.ZoneIdentifier is a helper to format the ID correctly for the API
	records, _, err := c.api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID), params)
	if err != nil {
		return false, "", fmt.Errorf("failed to list DNS records: %w", err)
	}

	// 3. Evaluate results
	if len(records) > 0 {
		// Record exists! Return the IP of the first match.
		return true, records[0].Content, nil
	}

	// No record found
	return false, "", nil
}

// CreateRecord provisions a new A record in Cloudflare
func (c *Client) CreateRecord(ctx context.Context, zoneName, recordName, targetIP string) error {
	// 1. Resolve the Zone ID
	zoneID, err := c.api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("failed to find zone '%s': %w", zoneName, err)
	}

	// 2. Define the record parameters
	proxied := false // Set to false as requested (DNS-only, no orange cloud)

	params := cf.CreateDNSRecordParams{
		Type:    "A",
		Name:    recordName,
		Content: targetIP,
		Proxied: &proxied,
		TTL:     1, // 1 means 'Automatic' TTL in Cloudflare
	}

	// 3. Make the API call to create the record
	_, err = c.api.CreateDNSRecord(ctx, cf.ZoneIdentifier(zoneID), params)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	return nil
}

// ListRecords retrieves a paginated list of all DNS records in the zone.
// We return the slice of records, the pagination info, and any error.
func (c *Client) ListRecords(ctx context.Context, zoneName string, page int, perPage int) ([]cf.DNSRecord, *cf.ResultInfo, error) {
	// 1. Resolve the Zone ID
	zoneID, err := c.api.ZoneIDByName(zoneName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find zone '%s': %w", zoneName, err)
	}

	// 2. Set up the pagination parameters
	params := cf.ListDNSRecordsParams{
		ResultInfo: cf.ResultInfo{
			Page:    page,
			PerPage: perPage,
		},
	}

	// 3. Make the API call
	records, resultInfo, err := c.api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID), params)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list records: %w", err)
	}

	return records, resultInfo, nil
}

// DeleteRecord searches for an A record by name, resolves its ID, and deletes it.
func (c *Client) DeleteRecord(ctx context.Context, zoneName, recordName string) error {
	// 1. Resolve the Zone ID
	zoneID, err := c.api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("failed to find zone '%s': %w", zoneName, err)
	}

	// 2. Search for the record to obtain its unique ID
	params := cf.ListDNSRecordsParams{
		Name: recordName,
		Type: "A",
	}
	records, _, err := c.api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID), params)
	if err != nil {
		return fmt.Errorf("failed to find record for deletion lookup: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no A record found with the name '%s'", recordName)
	}

	// 3. Delete the record using its retrieved ID
	recordID := records[0].ID
	err = c.api.DeleteDNSRecord(ctx, cf.ZoneIdentifier(zoneID), recordID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}
