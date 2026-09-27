package ui

import (
    "time"

    "github.com/redis/go-redis/v9"

    rclient "github.com/dheeraj080/redis-inspector/internal/redis"
)

type TickMsg time.Time

type MemoryDataMsg struct {
    Stats rclient.MemoryStats
    Err   error
}

type KeyDetailMsg struct {
    Detail rclient.KeyDetail
    Err    error
}

type ScannedKeysMsg struct {
    Keys    []rclient.KeyMem
    Pattern string
    Err     error
}

type KeySavedMsg struct {
    Err error
}

type KeyDeletedMsg struct {
    Key string
}

type KeyDeleteErrMsg struct {
    Err error
}

type NamespaceDeletedMsg struct {
    Count   int64
    Pattern string
    Err     error
}

type DBSwitchedMsg struct {
    DB     int
    Client *redis.Client
    Err    error
}

type DataSeededMsg struct{}

type DataSeedErrMsg struct {
    Err error
}
