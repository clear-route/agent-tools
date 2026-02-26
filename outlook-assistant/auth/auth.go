package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

var scopes = []string{
	"Mail.ReadWrite",
	"Mail.Send",
	"Calendars.ReadWrite",
	"User.Read",
}

const authRecordFile = ".outlook-assistant-auth.json"

func recordPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, authRecordFile), nil
}

func loadRecord() (azidentity.AuthenticationRecord, error) {
	record := azidentity.AuthenticationRecord{}
	path, err := recordPath()
	if err != nil {
		return record, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		// File not found is expected on first run
		return record, nil
	}
	err = json.Unmarshal(b, &record)
	return record, err
}

func saveRecord(record azidentity.AuthenticationRecord) error {
	path, err := recordPath()
	if err != nil {
		return err
	}
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

// NewGraphClient returns an authenticated Microsoft Graph client.
// On first run the user is prompted to log in via browser; subsequent runs
// reuse the cached token without any browser interaction.
func NewGraphClient(clientID, tenantID string) (*msgraphsdk.GraphServiceClient, error) {
	record, err := loadRecord()
	if err != nil {
		return nil, fmt.Errorf("loading auth record: %w", err)
	}

	persistentCache, err := cache.New(nil)
	if err != nil {
		// Persistent caching unavailable in this environment; fall back to memory-only.
		persistentCache = azidentity.Cache{}
	}

	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		ClientID:             clientID,
		TenantID:             tenantID,
		RedirectURL:          "http://localhost:4321",
		AuthenticationRecord: record,
		Cache:                persistentCache,
	})
	if err != nil {
		return nil, fmt.Errorf("creating credential: %w", err)
	}

	// If no record was stored, authenticate now and save the record so future
	// invocations skip the browser entirely.
	if record == (azidentity.AuthenticationRecord{}) {
		fmt.Fprintln(os.Stderr, "Opening browser for authentication…")
		newRecord, authErr := cred.Authenticate(context.Background(), &policy.TokenRequestOptions{
			Scopes: scopes,
		})
		if authErr != nil {
			return nil, fmt.Errorf("authenticating: %w", authErr)
		}
		if saveErr := saveRecord(newRecord); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save auth record: %v\n", saveErr)
		}
	}

	tokenProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
	if err != nil {
		return nil, fmt.Errorf("creating token provider: %w", err)
	}

	adapter, err := msgraphsdk.NewGraphRequestAdapter(tokenProvider)
	if err != nil {
		return nil, fmt.Errorf("creating graph adapter: %w", err)
	}

	return msgraphsdk.NewGraphServiceClient(adapter), nil
}
