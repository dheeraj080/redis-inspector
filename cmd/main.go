package main

import (
    "crypto/tls"
    "flag"
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/redis/go-redis/v9"

    "redis-inspector/internal/ui"
)

func main() {
    addr := flag.String("addr", "localhost:6379", "Redis server address (host:port)")
    password := flag.String("auth", "", "Redis password (-auth=\"secret\")")
    username := flag.String("user", "", "Redis ACL username (for Redis 6+)")
    db := flag.Int("db", 0, "Redis database index (0-15)")
    useTLS := flag.Bool("tls", false, "Enable TLS connection")
    skipVerify := flag.Bool("tls-skip-verify", false, "Skip TLS certificate verification")

    flag.Parse()

    opts := &redis.Options{
        Addr:     *addr,
        Username: *username,
        Password: *password,
        DB:       *db,
    }

    if *useTLS {
        opts.TLSConfig = &tls.Config{
            InsecureSkipVerify: *skipVerify,
        }
    }

    rdb := redis.NewClient(opts)

    model := ui.NewModel(rdb)
    p := tea.NewProgram(model, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running program: %v\n", err)
        os.Exit(1)
    }
}
