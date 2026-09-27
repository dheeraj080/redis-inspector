package redis

import (
    "testing"
)

func TestExtractNamespace(t *testing.T) {
    tests := []struct {
        key      string
        expected string
    }{
        {"user:session:123", "user:session:*"},
        {"cache:items:recent", "cache:items:*"},
        {"config:app_settings", "config:*"},
        {"standalone_key", "(root)"},
        {"a:b:c:d", "a:b:c:*"},
        {"colon:at:end:", "colon:at:end:*"},
    }

    for _, tt := range tests {
        t.Run(tt.key, func(t *testing.T) {
            got := ExtractNamespace(tt.key)
            if got != tt.expected {
                t.Errorf("ExtractNamespace(%q) = %q; want %q", tt.key, got, tt.expected)
            }
        })
    }
}

func TestParseInfoMemoryAndKeyspace(t *testing.T) {
    mockInfo := `# Memory
used_memory:10485760
used_memory_human:10.00M
used_memory_peak_human:12.50M
mem_fragmentation_ratio:1.15
mem_allocator:jemalloc-5.3.0

# Keyspace
db0:keys=150,expires=10,avg_ttl=3600
db1:keys=25,expires=0,avg_ttl=0
`

    stats := parseInfoMemoryAndKeyspace(mockInfo)

    if stats.UsedMemBytes != 10485760 {
        t.Errorf("UsedMemBytes = %d; want 10485760", stats.UsedMemBytes)
    }
    if stats.UsedMem != "10.00M" {
        t.Errorf("UsedMem = %q; want \"10.00M\"", stats.UsedMem)
    }
    if stats.PeakMem != "12.50M" {
        t.Errorf("PeakMem = %q; want \"12.50M\"", stats.PeakMem)
    }
    if stats.FragRatio != "1.15" {
        t.Errorf("FragRatio = %q; want \"1.15\"", stats.FragRatio)
    }
    if stats.Allocator != "jemalloc-5.3.0" {
        t.Errorf("Allocator = %q; want \"jemalloc-5.3.0\"", stats.Allocator)
    }

    if len(stats.Databases) != 16 {
        t.Fatalf("expected 16 databases, got %d", len(stats.Databases))
    }

    // Check DB0
    db0 := stats.Databases[0]
    if db0.DB != 0 || db0.Keys != 150 || db0.Expires != 10 || db0.AvgTTL != 3600 {
        t.Errorf("DB0 = %+v; want keys=150, expires=10, avg_ttl=3600", db0)
    }

    // Check DB1
    db1 := stats.Databases[1]
    if db1.DB != 1 || db1.Keys != 25 || db1.Expires != 0 || db1.AvgTTL != 0 {
        t.Errorf("DB1 = %+v; want keys=25, expires=0, avg_ttl=0", db1)
    }

    // Check DB2 (empty default)
    db2 := stats.Databases[2]
    if db2.DB != 2 || db2.Keys != 0 || db2.Expires != 0 {
        t.Errorf("DB2 = %+v; want empty", db2)
    }
}
