package main

import (
    "flag"
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/redis/go-redis/v9"

    "redis-inspector/internal/ui"
)

func main() {
    addr := flag.String("addr", "localhost:6379", "Redis server address")
    pass := flag.String("password", "", "Redis password")
    db := flag.Int("db", 0, "Redis database index")
    flag.Parse()

    rdb := redis.NewClient(&redis.Options{
        Addr:     *addr,
        Password: *pass,
        DB:       *db,
    })
    defer rdb.Close()

    p := tea.NewProgram(ui.NewModel(rdb), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running program: %v\n", err)
        os.Exit(1)
    }
}
