package main

import (
    "crypto/tls"
    "flag"
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/redis/go-redis/v9"

    "github.com/dheeraj080/redis-inspector/internal/ui"
)

func main() {
    addr := flag.String("addr", "localhost:6379", "Redis server address (host:port)")
    auth := flag.String("auth", "", "Redis password (-auth=\"secret\")")
    password := flag.String("password", "", "Redis password (alias for -auth)")
    username := flag.String("user", "", "Redis ACL username (for Redis 6+)")
    db := flag.Int("db", 0, "Redis database index (0-15)")
    useTLS := flag.Bool("tls", false, "Enable TLS connection")
    skipVerify := flag.Bool("tls-skip-verify", false, "Skip TLS certificate verification")

    flag.Parse()

    pass := *auth
    if pass == "" && *password != "" {
        pass = *password
    }

    opts := &redis.Options{
        Addr:     *addr,
        Username: *username,
        Password: pass,
        DB:       *db,
    }

    if *useTLS {
        opts.TLSConfig = &tls.Config{
            InsecureSkipVerify: *skipVerify, //nolint:gosec
        }
    }

    rdb := redis.NewClient(opts)
    defer rdb.Close()

    model := ui.NewModel(rdb, opts)
    p := tea.NewProgram(model, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running program: %v\n", err)
        os.Exit(1)
    }
}
