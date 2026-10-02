package auth

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	sessionmodel "shijibu/internal/model/redis"
)

func testRedis(t *testing.T) (*sessionmodel.Client, *goredis.Client) {
	t.Helper()
	binary, err := exec.LookPath("redis-server")
	if err != nil {
		t.Skip("redis-server is required for real Redis session integration tests")
	}
	dir, err := os.MkdirTemp("", "nl-redis-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})
	socket := filepath.Join(dir, "redis.sock")
	cmd := exec.Command(binary, "--port", "0", "--unixsocket", socket, "--save", "", "--appendonly", "no")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Error(err)
		}
		if err := cmd.Wait(); err != nil {
			t.Error(err)
		}
	})
	client, err := sessionmodel.NewRedis("unix://" + socket)
	if err != nil {
		t.Fatal(err)
	}
	raw := goredis.NewClient(&goredis.Options{Network: "unix", Addr: socket})
	t.Cleanup(func() {
		if err := client.Close(); err != nil && err.Error() != "redis: client is closed" {
			t.Error(err)
		}
		if err := raw.Close(); err != nil {
			t.Error(err)
		}
	})
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := client.Init(context.Background()); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Redis startup timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return client, raw
}

func TestRedisSessionLifecycle(t *testing.T) {
	store, raw := testRedis(t)
	ctx := context.Background()
	manager := NewSessionManager("01234567890123456789012345678901", 2*time.Hour, store)
	pair, err := manager.Issue(ctx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if pair.ExpiresIn != 7200 || pair.RefreshExpiresIn != 7200 {
		t.Fatal("expected two-hour tokens")
	}
	if id, err := manager.Authenticate(ctx, pair.Token); err != nil || id != 42 {
		t.Fatalf("authenticate: %d %v", id, err)
	}
	if _, err := manager.Authenticate(ctx, pair.RefreshToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("refresh token accepted as access token")
	}
	if _, err := manager.Refresh(ctx, pair.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("access token accepted as refresh token")
	}
	other := NewSessionManager("abcdefghijklmnopqrstuvwxyz123456", 2*time.Hour, store)
	if _, err := other.Authenticate(ctx, pair.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("wrong secret accepted")
	}
	claims, err := manager.claims(pair.Token, "access")
	if err != nil {
		t.Fatal(err)
	}
	key := "v1:session:{" + claims.SessionID + "}"
	ttl, err := raw.TTL(ctx, key).Result()
	if err != nil || ttl < 7190*time.Second || ttl > 7200*time.Second {
		t.Fatalf("TTL = %v, %v", ttl, err)
	}
	values, err := raw.HGetAll(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if values["access_hash"] != tokenHash(pair.Token) || values["refresh_hash"] != tokenHash(pair.RefreshToken) {
		t.Fatal("credentials must be stored as hashes")
	}
	var winners atomic.Int32
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.Refresh(ctx, pair.RefreshToken)
			if err == nil {
				winners.Add(1)
			} else if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("refresh: %v", err)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("refresh winners = %d", winners.Load())
	}
	if _, err := manager.Authenticate(ctx, pair.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("old access token accepted after rotation")
	}
	pair, err = manager.Issue(ctx, 42)
	if err != nil {
		t.Fatal(err)
	}
	claims, err = manager.claims(pair.Token, "access")
	if err != nil {
		t.Fatal(err)
	}
	key = "v1:session:{" + claims.SessionID + "}"
	if err := raw.PExpire(ctx, key, time.Millisecond).Err(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := manager.Authenticate(ctx, pair.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("expired Redis session accepted")
	}
	if _, err := manager.Refresh(ctx, pair.RefreshToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("expired Redis refresh accepted")
	}
}

func TestRedisOutageFailsClosed(t *testing.T) {
	store, _ := testRedis(t)
	manager := NewSessionManager("01234567890123456789012345678901", 2*time.Hour, store)
	ctx := context.Background()
	pair, err := manager.Issue(ctx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Authenticate(ctx, pair.Token); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("outage authentication: %v", err)
	}
	if _, err := manager.Refresh(ctx, pair.RefreshToken); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("outage refresh: %v", err)
	}
	if _, err := manager.Issue(ctx, 42); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("outage login: %v", err)
	}
}
