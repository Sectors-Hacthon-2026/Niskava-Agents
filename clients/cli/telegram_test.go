package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
)

func TestTelegramCLICommandRegistered(t *testing.T) {
	foundTelegramCmd := false
	for _, cmd := range RootCmd.Commands() {
		if cmd.Name() == "telegram" {
			foundTelegramCmd = true
			break
		}
	}
	if !foundTelegramCmd {
		t.Errorf("expected 'telegram' subcommand to be registered in RootCmd")
	}

	telegramFlag := serveCmd.Flags().Lookup("telegram")
	if telegramFlag == nil {
		t.Errorf("expected '--telegram' flag to be present on serveCmd")
	}
}

func TestTelegramWhitelistHelpers(t *testing.T) {
	testCfg := &config.Config{
		Telegram: config.TelegramConfig{
			AllowedUsers: []string{"12345678", "user_alpha"},
		},
	}

	// Test Add duplicate
	added, err := AddTelegramAllowedUser(testCfg, "12345678")
	if err != nil || added {
		t.Fatalf("expected duplicate add to return false, got added=%v, err=%v", added, err)
	}

	// Test Add new user (numeric ID)
	added, err = AddTelegramAllowedUser(testCfg, "98765432")
	if err != nil || !added {
		t.Fatalf("expected new user add to return true, got added=%v, err=%v", added, err)
	}
	if len(testCfg.Telegram.AllowedUsers) != 3 {
		t.Fatalf("expected 3 users, got %d", len(testCfg.Telegram.AllowedUsers))
	}

	// Test Add username with @ prefix - should strip @
	added, err = AddTelegramAllowedUser(testCfg, "@new_analyst")
	if err != nil || !added {
		t.Fatalf("expected @new_analyst add to return true, got added=%v, err=%v", added, err)
	}
	if testCfg.Telegram.AllowedUsers[3] != "new_analyst" {
		t.Fatalf("expected stripped username 'new_analyst', got %s", testCfg.Telegram.AllowedUsers[3])
	}

	// Test Remove existing
	removed, err := RemoveTelegramAllowedUser(testCfg, "user_alpha")
	if err != nil || !removed {
		t.Fatalf("expected remove to return true, got removed=%v, err=%v", removed, err)
	}
	if len(testCfg.Telegram.AllowedUsers) != 3 {
		t.Fatalf("expected 3 users remaining, got %d", len(testCfg.Telegram.AllowedUsers))
	}

	// Test Remove non-existent
	removed, err = RemoveTelegramAllowedUser(testCfg, "unknown_user")
	if err != nil || removed {
		t.Fatalf("expected remove non-existent to return false, got %v", removed)
	}
}

func TestFetchTelegramBotInfo_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "invalid") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"ok":false,"error_code":401,"description":"Unauthorized"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"id":99887766,"is_bot":true,"first_name":"Niskava Bot","username":"niskava_test_bot"}}`))
	}))
	defer server.Close()

	// Helper test with custom endpoint URL
	info, err := fetchTelegramBotInfoWithBaseURL("test_token", server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Username != "niskava_test_bot" || info.ID != 99887766 {
		t.Fatalf("unexpected bot info: %+v", info)
	}

	// Invalid token test
	_, err = fetchTelegramBotInfoWithBaseURL("invalid_token", server.Client(), server.URL)
	if err == nil {
		t.Fatalf("expected error for invalid token, got nil")
	}
}

func TestTelegramSubcommandsRegistered(t *testing.T) {
	var hasStatus, hasUser, hasTest bool
	for _, cmd := range telegramCmd.Commands() {
		switch cmd.Name() {
		case "status":
			hasStatus = true
		case "user":
			hasUser = true
			var hasList, hasAdd, hasRemove bool
			for _, sub := range cmd.Commands() {
				if sub.Name() == "list" {
					hasList = true
				}
				if sub.Name() == "add" {
					hasAdd = true
				}
				if sub.Name() == "remove" {
					hasRemove = true
				}
			}
			if !hasList || !hasAdd || !hasRemove {
				t.Errorf("user command missing subcommands: list=%v, add=%v, remove=%v", hasList, hasAdd, hasRemove)
			}
		case "test":
			hasTest = true
		}
	}
	if !hasStatus || !hasUser || !hasTest {
		t.Errorf("expected status, user, and test subcommands under telegramCmd (status=%v, user=%v, test=%v)", hasStatus, hasUser, hasTest)
	}
}
