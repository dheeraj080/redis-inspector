package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
)

type KeyMem struct {
    Key   string
    Bytes int64
    Type  string
}

type NamespaceMem struct {
    Prefix string
    Bytes  int64
    Count  int
}

type KeyDetail struct {
    Key      string
    Type     string
    Encoding string
    Bytes    int64
    TTL      string
    Elements int64
}

type MemoryStats struct {
    UsedMem    string         `json:"used_memory_human"`
    PeakMem    string         `json:"peak_memory_human"`
    FragRatio  string         `json:"mem_fragmentation_ratio"`
    Allocator  string         `json:"mem_allocator"`
    TopKeys    []KeyMem       `json:"top_keys"`
    Namespaces []NamespaceMem `json:"namespaces"`
}

func ExtractNamespace(key string) string {
    idx := strings.LastIndex(key, ":")
    if idx != -1 {
        return key[:idx] + ":*"
    }
    return "(root)"
}

func FetchMemoryData(ctx context.Context, rdb *redis.Client) (MemoryStats, error) {
    info, err := rdb.Info(ctx, "memory").Result()
    if err != nil {
        return MemoryStats{}, err
    }

    res := parseInfoMemory(info)

    var cursor uint64
    var keys []string

    for i := 0; i < 10; i++ {
        var fetched []string
        var scanErr error
        fetched, cursor, scanErr = rdb.Scan(ctx, cursor, "*", 100).Result()
        if scanErr != nil {
            break
        }
        keys = append(keys, fetched...)

        if cursor == 0 || len(keys) >= 500 {
            break
        }
    }

    if len(keys) > 0 {
        pipe := rdb.Pipeline()
        cmds := make(map[string]*redis.IntCmd, len(keys))
        typeCmds := make(map[string]*redis.StatusCmd, len(keys))

        for _, k := range keys {
            cmds[k] = pipe.MemoryUsage(ctx, k)
            typeCmds[k] = pipe.Type(ctx, k)
        }
        _, _ = pipe.Exec(ctx)

        parsedKeys := make([]KeyMem, 0, len(keys))
        nsMap := make(map[string]*NamespaceMem)

        for _, k := range keys {
            bytes, err := cmds[k].Result()
            if err != nil {
                bytes = 0
            }
            kType, err := typeCmds[k].Result()
            if err != nil {
                kType = "unknown"
            }
            parsedKeys = append(parsedKeys, KeyMem{Key: k, Bytes: bytes, Type: kType})

            prefix := ExtractNamespace(k)
            if ns, exists := nsMap[prefix]; exists {
                ns.Bytes += bytes
                ns.Count++
            } else {
                nsMap[prefix] = &NamespaceMem{
                    Prefix: prefix,
                    Bytes:  bytes,
                    Count:  1,
                }
            }
        }

        sort.Slice(parsedKeys, func(i, j int) bool {
            return parsedKeys[i].Bytes > parsedKeys[j].Bytes
        })

        namespaces := make([]NamespaceMem, 0, len(nsMap))
        for _, ns := range nsMap {
            namespaces = append(namespaces, *ns)
        }

        sort.Slice(namespaces, func(i, j int) bool {
            return namespaces[i].Bytes > namespaces[j].Bytes
        })

        res.TopKeys = parsedKeys
        res.Namespaces = namespaces
    }

    return res, nil
}

func FetchKeyDetails(ctx context.Context, rdb *redis.Client, key string) (KeyDetail, error) {
    pipe := rdb.Pipeline()
    typeCmd := pipe.Type(ctx, key)
    ttlCmd := pipe.TTL(ctx, key)
    memCmd := pipe.MemoryUsage(ctx, key)
    encCmd := pipe.ObjectEncoding(ctx, key)

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return KeyDetail{}, err
    }

    kType, _ := typeCmd.Result()
    ttl, _ := ttlCmd.Result()
    bytes, _ := memCmd.Result()
    encoding, _ := encCmd.Result()

    var count int64
    switch kType {
    case "string":
        count, _ = rdb.StrLen(ctx, key).Result()
    case "hash":
        count, _ = rdb.HLen(ctx, key).Result()
    case "list":
        count, _ = rdb.LLen(ctx, key).Result()
    case "set":
        count, _ = rdb.SCard(ctx, key).Result()
    case "zset":
        count, _ = rdb.ZCard(ctx, key).Result()
    case "stream":
        count, _ = rdb.XLen(ctx, key).Result()
    }

    ttlStr := "No Expiry"
    if ttl > 0 {
        ttlStr = ttl.String()
    } else if ttl == -2 {
        ttlStr = "Key Expired / Not Found"
    }

    return KeyDetail{
        Key:      key,
        Type:     kType,
        Encoding: encoding,
        Bytes:    bytes,
        TTL:      ttlStr,
        Elements: count,
    }, nil
}

func FetchKeyValue(ctx context.Context, rdb *redis.Client, key, kType string) (string, error) {
    switch kType {
    case "string":
        val, err := rdb.Get(ctx, key).Result()
        if err != nil {
            return "", err
        }
        var js map[string]interface{}
        if json.Unmarshal([]byte(val), &js) == nil {
            pretty, err := json.MarshalIndent(js, "", "  ")
            if err == nil {
                return string(pretty), nil
            }
        }
        return val, nil

    case "hash":
        val, err := rdb.HGetAll(ctx, key).Result()
        if err != nil {
            return "", err
        }
        pretty, _ := json.MarshalIndent(val, "", "  ")
        return string(pretty), nil

    case "list":
        val, err := rdb.LRange(ctx, key, 0, 50).Result()
        if err != nil {
            return "", err
        }
        pretty, _ := json.MarshalIndent(val, "", "  ")
        return string(pretty), nil

    case "set":
        val, err := rdb.SMembers(ctx, key).Result()
        if err != nil {
            return "", err
        }
        pretty, _ := json.MarshalIndent(val, "", "  ")
        return string(pretty), nil

    case "zset":
        val, err := rdb.ZRangeWithScores(ctx, key, 0, 50).Result()
        if err != nil {
            return "", err
        }
        pretty, _ := json.MarshalIndent(val, "", "  ")
        return string(pretty), nil

    default:
        return "(Binary or unsupported type)", nil
    }
}

func SetKeyTTL(ctx context.Context, rdb *redis.Client, key string, seconds int) error {
    if seconds < 0 {
        return rdb.Persist(ctx, key).Err()
    }
    return rdb.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
}

func ExportReport(stats MemoryStats) (string, error) {
    filename := fmt.Sprintf("redis_memory_report_%d.json", time.Now().Unix())
    data, err := json.MarshalIndent(stats, "", "  ")
    if err != nil {
        return "", err
    }
    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        return "", err
    }
    return filename, nil
}

func DeleteKey(ctx context.Context, rdb *redis.Client, key string) error {
    return rdb.Unlink(ctx, key).Err()
}

func SeedMockData(ctx context.Context, rdb *redis.Client) error {
    pipe := rdb.Pipeline()

    for i := 1; i <= 15; i++ {
        key := fmt.Sprintf("user:session:%d", i)
        val := strings.Repeat("x", i*250)
        pipe.Set(ctx, key, val, 0)
    }

    pipe.Set(ctx, "config:app_settings", `{"theme":"dark","timeout":30,"features":["auth","billing"]}`, 0)

    for i := 1; i <= 5; i++ {
        key := fmt.Sprintf("cache:item:%d", i)
        pipe.HSet(ctx, key, map[string]interface{}{
            "id":      i,
            "title":   fmt.Sprintf("Sample Item %d", i),
            "payload": strings.Repeat("data_", i*50),
        })
    }

    for i := 1; i <= 10; i++ {
        pipe.RPush(ctx, "queue:jobs", fmt.Sprintf("job_payload_bytes_chunk_%d", i))
    }

    pipe.SAdd(ctx, "tags:active", "golang", "redis", "bubbletea", "tui", "performance")

    pipe.ZAdd(ctx, "leaderboard:scores",
        redis.Z{Score: 100, Member: "player_one"},
        redis.Z{Score: 250, Member: "player_two"},
        redis.Z{Score: 50, Member: "player_three"},
    )

    _, err := pipe.Exec(ctx)
    return err
}

func parseInfoMemory(info string) MemoryStats {
    res := MemoryStats{}
    lines := strings.Split(info, "\r\n")
    for _, line := range lines {
        parts := strings.Split(line, ":")
        if len(parts) < 2 {
            continue
        }
        switch parts[0] {
        case "used_memory_human":
            res.UsedMem = parts[1]
        case "used_memory_peak_human":
            res.PeakMem = parts[1]
        case "mem_fragmentation_ratio":
            res.FragRatio = parts[1]
        case "mem_allocator":
            res.Allocator = parts[1]
        }
    }
    return res
}
