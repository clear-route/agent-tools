package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

var scopes = []string{
	"User.Read",
	"Calendars.Read",
	"OnlineMeetings.Read",
	"OnlineMeetingTranscript.Read.All",
}

const authRecordFile = ".teams-assistant-auth.json"

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

func Scopes() []string {
	return scopes
}

func NewCredential(clientID, tenantID string) (azcore.TokenCredential, error) {
	record, err := loadRecord()
	if err != nil {
		return nil, fmt.Errorf("loading auth record: %w", err)
	}

	type cacheResult struct {
		cache azidentity.Cache
		err   error
	}
	ch := make(chan cacheResult, 1)
	go func() {
		c, err := cache.New(nil)
		ch <- cacheResult{c, err}
	}()

	var persistentCache azidentity.Cache
	select {
	case res := <-ch:
		if res.err == nil {
			persistentCache = res.cache
		}
	case <-time.After(3 * time.Second):
		fmt.Fprintln(os.Stderr, "warning: keychain cache timed out, using memory-only token cache")
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

	if record == (azidentity.AuthenticationRecord{}) {
		fmt.Fprintln(os.Stderr, "Opening browser for authentication\u2026")
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

	return cred, nil
}
