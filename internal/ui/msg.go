package ui

import (
    "time"

    rclient "redis-inspector/internal/redis"
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

type KeyDeletedMsg struct {
    Key string
}

type KeyDeleteErrMsg struct {
    Err error
}

type DataSeededMsg struct{}

type DataSeedErrMsg struct {
    Err error
}
