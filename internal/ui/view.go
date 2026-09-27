package ui

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/charmbracelet/lipgloss"

    rclient "github.com/dheeraj080/redis-inspector/internal/redis"
    "github.com/dheeraj080/redis-inspector/internal/utils"
)

func (m Model) View() string {
    if m.err != nil {
        errBoxStyle := lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("#F7768E")).
            Padding(1, 2).
            Margin(1, 2)
        errMsg := fmt.Sprintf("⚠️  Redis Connection Error:\n\n  %v\n\n  Please verify your Redis server is running and connection parameters are correct.\n  Press 'q' or 'Ctrl+C' to quit.", m.err)
        return errBoxStyle.Render(errMsg)
    }

    // --- Color Palette (Cyberpunk / Tokyo Night inspired) ---
    cPrimary := lipgloss.Color("#7AA2F7")   // Soft electric blue
    cSecondary := lipgloss.Color("#BB9AF7") // Cyber violet
    cMint := lipgloss.Color("#73DACA")      // Mint green
    cGreen := lipgloss.Color("#9ECE6A")     // Emerald green
    cCoral := lipgloss.Color("#F7768E")     // Coral red (danger)
    cAmber := lipgloss.Color("#E0AF68")     // Warm gold/amber
    cCyan := lipgloss.Color("#7DCFFF")      // Bright cyan
    cBorder := lipgloss.Color("#3B4261")    // Subtle card border
    cBorderActive := lipgloss.Color("#7AA2F7")
    cBgPill := lipgloss.Color("#24283B")
    cTextMain := lipgloss.Color("#C0CAF5")
    cTextDim := lipgloss.Color("#565F89")
    cTextBright := lipgloss.Color("#FFFFFF")

    // --- Base Styles ---
    cardStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(cBorder).
        Padding(0, 1)

    // Calculate content width
    totalWidth := m.width
    if totalWidth <= 0 {
        totalWidth = 84
    }
    boxInnerWidth := totalWidth - 6
    if boxInnerWidth < 74 {
        boxInnerWidth = 74
    }

    // --- 1. Top Header Banner ---
    brandStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(cTextBright).
        Background(cPrimary).
        Padding(0, 1)

    hostAddr := "localhost:6379"
    if m.redisOpts != nil && m.redisOpts.Addr != "" {
        hostAddr = m.redisOpts.Addr
    }

    dbPillStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(cCyan).
        Background(cBgPill).
        Padding(0, 1)

    statusPillStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(cGreen).
        Background(cBgPill).
        Padding(0, 1)

    hostPillStyle := lipgloss.NewStyle().
        Foreground(cTextDim).
        Background(cBgPill).
        Padding(0, 1)

    headerLeft := brandStyle.Render("⚡ REDIS INSPECTOR")
    dbPill := dbPillStyle.Render(fmt.Sprintf("DB %d", m.currentDB))
    statusPill := statusPillStyle.Render("● ONLINE")
    hostPill := hostPillStyle.Render(hostAddr)

    header := fmt.Sprintf(" %s  %s  %s  %s", headerLeft, dbPill, statusPill, hostPill)

    // --- 2. Live Metrics Grid (4 Stats Cards) ---
    fragNum, _ := strconv.ParseFloat(m.fragRatio, 64)
    fragColor := cGreen
    fragStatus := "Optimal"
    if fragNum >= 2.0 {
        fragColor = cCoral
        fragStatus = "High Frag"
    } else if fragNum >= 1.5 {
        fragColor = cAmber
        fragStatus = "Moderate"
    }
    if m.fragRatio == "" {
        fragStatus = "N/A"
    }

    sparkline := utils.RenderSparkline(m.memHistory)
    if sparkline == "" {
        sparkline = "collecting..."
    }
    renderedSparkline := lipgloss.NewStyle().Bold(true).Foreground(cMint).Render(sparkline)

    usedMemStr := m.usedMem
    if usedMemStr == "" {
        usedMemStr = "0 B"
    }
    peakMemStr := m.peakMem
    if peakMemStr == "" {
        peakMemStr = "0 B"
    }
    fragStr := m.fragRatio
    if fragStr == "" {
        fragStr = "1.00"
    }
    allocatorStr := m.allocator
    if allocatorStr == "" {
        allocatorStr = "jemalloc"
    }

    metricLabelStyle := lipgloss.NewStyle().Bold(true).Foreground(cTextDim)
    metricValStyle := lipgloss.NewStyle().Bold(true)
    subLabelStyle := lipgloss.NewStyle().Foreground(cTextDim)

    // 4 Stat Cards
    card1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cBorder).Padding(0, 1).Width(18).Render(
        fmt.Sprintf("%s\n%s\n%s",
            metricLabelStyle.Render("USED MEMORY"),
            metricValStyle.Foreground(cCyan).Render(usedMemStr),
            subLabelStyle.Render(fmt.Sprintf("%d B", m.currentStats().UsedMemBytes)),
        ),
    )

    card2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cBorder).Padding(0, 1).Width(18).Render(
        fmt.Sprintf("%s\n%s\n%s",
            metricLabelStyle.Render("PEAK MEMORY"),
            metricValStyle.Foreground(cSecondary).Render(peakMemStr),
            subLabelStyle.Render("Alloc: "+utils.TruncateString(allocatorStr, 10)),
        ),
    )

    card3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cBorder).Padding(0, 1).Width(18).Render(
        fmt.Sprintf("%s\n%s\n%s",
            metricLabelStyle.Render("FRAG RATIO"),
            metricValStyle.Foreground(fragColor).Render(fragStr),
            lipgloss.NewStyle().Foreground(fragColor).Render("● "+fragStatus),
        ),
    )

    trendWidth := boxInnerWidth - 18*3 - 8
    if trendWidth < 22 {
        trendWidth = 22
    }
    card4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cBorder).Padding(0, 1).Width(trendWidth).Render(
        fmt.Sprintf("%s\n%s\n%s",
            metricLabelStyle.Render("30s MEMORY TREND"),
            renderedSparkline,
            subLabelStyle.Render("Live telemetry window"),
        ),
    )

    statsRow := lipgloss.JoinHorizontal(lipgloss.Top, card1, " ", card2, " ", card3, " ", card4)

    // --- 3. Body Content / Modals ---
    var bodyBox string
    var footer string

    if m.showEditModal {
        editBoxStyle := cardStyle.Copy().BorderForeground(cMint).Padding(1, 2)
        editHeader := lipgloss.NewStyle().Bold(true).Foreground(cMint).Render("✏️  EDIT KEY VALUE")

        typeBadge := renderTypeBadge(m.activeDetail.Type)
        promptText := fmt.Sprintf(
            "Key: %s  %s\n\nEnter new content below (JSON format for hash):\n\n%s",
            lipgloss.NewStyle().Bold(true).Foreground(cTextBright).Render(m.activeDetail.Key),
            typeBadge,
            m.editorInput.View(),
        )

        bodyBox = editBoxStyle.Render(fmt.Sprintf("%s\n\n%s", editHeader, promptText))
        footer = renderFooterKeys([]keyHelp{
            {"Enter", "Save Update"},
            {"Esc", "Cancel"},
        })

    } else if m.showDBModal {
        dbBoxStyle := cardStyle.Copy().BorderForeground(cSecondary).Padding(1, 2)
        dbHeader := lipgloss.NewStyle().Bold(true).Foreground(cSecondary).Render("🗄️  SELECT REDIS DATABASE (DB 0 - 15)")

        // Render DBs in 2 clean columns of 8
        var col1, col2 strings.Builder
        for i := 0; i < 16; i++ {
            var keysCount int64 = 0
            var expiresCount int64 = 0
            for _, db := range m.databases {
                if db.DB == i {
                    keysCount = db.Keys
                    expiresCount = db.Expires
                    break
                }
            }

            activeMarker := "  "
            if i == m.currentDB {
                activeMarker = "★ "
            }

            rowText := fmt.Sprintf("%sDB %02d : %6d keys", activeMarker, i, keysCount)
            if expiresCount > 0 {
                rowText += fmt.Sprintf(" (%d exp)", expiresCount)
            }

            var line string
            if i == m.dbSelected {
                line = lipgloss.NewStyle().Bold(true).Foreground(cCyan).Background(lipgloss.Color("#283457")).Render(" 👉 " + rowText + " ")
            } else if i == m.currentDB {
                line = lipgloss.NewStyle().Bold(true).Foreground(cMint).Render("    " + rowText)
            } else {
                line = lipgloss.NewStyle().Foreground(cTextMain).Render("    " + rowText)
            }

            if i < 8 {
                col1.WriteString(line + "\n")
            } else {
                col2.WriteString(line + "\n")
            }
        }

        dbGrid := lipgloss.JoinHorizontal(lipgloss.Top, col1.String(), "   ", col2.String())
        bodyBox = dbBoxStyle.Render(fmt.Sprintf("%s\n\n%s", dbHeader, dbGrid))
        footer = renderFooterKeys([]keyHelp{
            {"↑/↓/←/→", "Navigate DBs"},
            {"Enter", "Switch DB"},
            {"Esc", "Cancel"},
        })

    } else if m.showValueViewer {
        valBoxStyle := cardStyle.Copy().BorderForeground(cCyan).Padding(1, 2)
        typeBadge := renderTypeBadge(m.activeDetail.Type)
        valHeader := lipgloss.NewStyle().Bold(true).Foreground(cCyan).Render(fmt.Sprintf("📄 RAW VALUE VIEWER: %s %s", m.activeDetail.Key, typeBadge))

        bodyBox = valBoxStyle.Render(fmt.Sprintf("%s\n\n%s", valHeader, m.activeValue))
        footer = renderFooterKeys([]keyHelp{
            {"Esc / q", "Return to List"},
        })

    } else if m.showTTLModal {
        ttlBoxStyle := cardStyle.Copy().BorderForeground(cAmber).Padding(1, 2)
        ttlHeader := lipgloss.NewStyle().Bold(true).Foreground(cAmber).Render("⏱️  CONFIGURE KEY EXPIRATION (TTL)")

        promptText := fmt.Sprintf(
            "Key: %s\nCurrent TTL: %s\n\nEnter new TTL in seconds (e.g. 3600 for 1h, -1 to PERSIST indefinitely):\n\n%s",
            lipgloss.NewStyle().Bold(true).Foreground(cTextBright).Render(m.activeDetail.Key),
            lipgloss.NewStyle().Foreground(cCyan).Render(m.activeDetail.TTL),
            m.ttlInput.View(),
        )

        bodyBox = ttlBoxStyle.Render(fmt.Sprintf("%s\n\n%s", ttlHeader, promptText))
        footer = renderFooterKeys([]keyHelp{
            {"Enter", "Confirm TTL"},
            {"Esc", "Cancel"},
        })

    } else if m.showDetails {
        detailBoxStyle := cardStyle.Copy().BorderForeground(cPrimary).Padding(1, 2)
        detailHeader := lipgloss.NewStyle().Bold(true).Foreground(cPrimary).Render("🔍 KEY INSPECTOR")

        countLabel := "Items / Elements:"
        if m.activeDetail.Type == "string" {
            countLabel = "String Length (bytes):"
        }

        typeBadge := renderTypeBadge(m.activeDetail.Type)
        labelStyle := lipgloss.NewStyle().Bold(true).Foreground(cTextDim).Width(22)
        valStyle := lipgloss.NewStyle().Foreground(cTextBright)

        grid := fmt.Sprintf(
            "%s %s\n"+
                "%s %s\n"+
                "%s %s\n"+
                "%s %s (%d bytes)\n"+
                "%s %s\n"+
                "%s %d",
            labelStyle.Render("🏷️  Key Name:"), valStyle.Render(m.activeDetail.Key),
            labelStyle.Render("📦 Data Type:"), typeBadge,
            labelStyle.Render("⚙️  Encoding:"), valStyle.Render(m.activeDetail.Encoding),
            labelStyle.Render("💾 Memory Footprint:"), lipgloss.NewStyle().Bold(true).Foreground(cCyan).Render(utils.FormatBytes(m.activeDetail.Bytes)), m.activeDetail.Bytes,
            labelStyle.Render("⏱️  Time to Live:"), lipgloss.NewStyle().Foreground(cAmber).Render(m.activeDetail.TTL),
            labelStyle.Render("🔢 "+countLabel), m.activeDetail.Elements,
        )

        bodyBox = detailBoxStyle.Render(fmt.Sprintf("%s\n\n%s", detailHeader, grid))

        actions := []keyHelp{
            {"v", "Raw Value"},
            {"t", "Set TTL"},
        }
        if m.activeDetail.Type == "string" || m.activeDetail.Type == "hash" {
            actions = append(actions, keyHelp{"e", "Edit Value"})
        }
        actions = append(actions, keyHelp{"Esc", "Back"})
        footer = renderFooterKeys(actions)

    } else {
        // --- Tab Bar ---
        tab1Text := fmt.Sprintf("1. Top Keys (%d)", len(m.filteredKeys()))
        tab2Text := fmt.Sprintf("2. Namespaces (%d)", len(m.filteredNamespaces()))

        activeTabStyle := lipgloss.NewStyle().
            Bold(true).
            Foreground(cTextBright).
            Background(cPrimary).
            Padding(0, 1)

        inactiveTabStyle := lipgloss.NewStyle().
            Foreground(cTextDim).
            Background(cBgPill).
            Padding(0, 1)

        var tab1, tab2 string
        if m.viewMode == 0 {
            tab1 = activeTabStyle.Render(tab1Text)
            tab2 = inactiveTabStyle.Render(tab2Text)
        } else {
            tab1 = inactiveTabStyle.Render(tab1Text)
            tab2 = activeTabStyle.Render(tab2Text)
        }
        tabsBar := fmt.Sprintf("  %s  %s", tab1, tab2)

        // --- Table Headers & Rows ---
        var tableContent string

        if m.viewMode == 0 {
            // View Mode 0: Top Keys Table
            fk := m.filteredKeys()
            start := m.page * m.pageSize
            end := start + m.pageSize
            if end > len(fk) {
                end = len(fk)
            }

            pageItems := []rclient.KeyMem{}
            if start < len(fk) {
                pageItems = fk[start:end]
            }

            var maxBytes int64 = 1
            if len(pageItems) > 0 && pageItems[0].Bytes > 0 {
                maxBytes = pageItems[0].Bytes
            }

            // Table Header
            thStyle := lipgloss.NewStyle().Bold(true).Foreground(cTextDim)
            th := fmt.Sprintf("   %-4s %-32s %-8s %11s  %-10s",
                thStyle.Render("#"),
                thStyle.Render("KEY NAME"),
                thStyle.Render("TYPE"),
                thStyle.Render("MEMORY"),
                thStyle.Render("PROPORTION"),
            )
            divider := lipgloss.NewStyle().Foreground(cBorder).Render("  " + strings.Repeat("─", boxInnerWidth-4))

            var rows strings.Builder
            for i, k := range pageItems {
                idxStr := fmt.Sprintf("%2d.", start+i+1)
                truncKey := utils.TruncateString(k.Key, 30)
                typeBadge := renderTypeBadge(k.Type)
                memStr := utils.FormatBytes(k.Bytes)
                miniBar := renderMiniBar(k.Bytes, maxBytes, 8)

                rowStr := fmt.Sprintf(" %-4s %-32s %-8s %11s  %s", idxStr, truncKey, typeBadge, memStr, miniBar)

                if i == m.selected {
                    selStyle := lipgloss.NewStyle().
                        Bold(true).
                        Foreground(cTextBright).
                        Background(lipgloss.Color("#283457"))
                    rows.WriteString(" 👉 " + selStyle.Render(rowStr) + "\n")
                } else {
                    rows.WriteString("    " + lipgloss.NewStyle().Foreground(cTextMain).Render(rowStr) + "\n")
                }
            }

            if len(pageItems) == 0 {
                rows.WriteString("\n   " + lipgloss.NewStyle().Foreground(cTextDim).Render("No matching keys found in database. Press 's' to seed test data.") + "\n\n")
            }

            pageInfo := lipgloss.NewStyle().Foreground(cTextDim).Render(
                fmt.Sprintf("Page %d of %d  (Keys: %d)", m.page+1, m.maxPages(), len(fk)),
            )
            if m.truncated {
                warnStyle := lipgloss.NewStyle().Bold(true).Foreground(cAmber)
                pageInfo += "  " + warnStyle.Render("⚠️ Sampled top 500 keys")
            }

            tableHeader := fmt.Sprintf("  %s   %s", lipgloss.NewStyle().Bold(true).Foreground(cCyan).Render("Top Memory Consumers"), pageInfo)
            tableContent = fmt.Sprintf("%s\n\n%s\n%s\n%s", tableHeader, th, divider, rows.String())

        } else {
            // View Mode 1: Namespaces Table
            fn := m.filteredNamespaces()
            start := m.page * m.pageSize
            end := start + m.pageSize
            if end > len(fn) {
                end = len(fn)
            }

            pageItems := []rclient.NamespaceMem{}
            if start < len(fn) {
                pageItems = fn[start:end]
            }

            thStyle := lipgloss.NewStyle().Bold(true).Foreground(cTextDim)
            th := fmt.Sprintf("   %-4s %-30s %-12s %11s  %12s",
                thStyle.Render("#"),
                thStyle.Render("NAMESPACE PREFIX"),
                thStyle.Render("KEY COUNT"),
                thStyle.Render("TOTAL MEM"),
                thStyle.Render("AVG / KEY"),
            )
            divider := lipgloss.NewStyle().Foreground(cBorder).Render("  " + strings.Repeat("─", boxInnerWidth-4))

            var rows strings.Builder
            for i, ns := range pageItems {
                idxStr := fmt.Sprintf("%2d.", start+i+1)
                truncPrefix := utils.TruncateString(ns.Prefix, 28)
                countStr := fmt.Sprintf("%d keys", ns.Count)
                totalMem := utils.FormatBytes(ns.Bytes)
                avgMem := "0 B"
                if ns.Count > 0 {
                    avgMem = utils.FormatBytes(ns.Bytes / int64(ns.Count))
                }

                countPill := lipgloss.NewStyle().Foreground(cAmber).Render(fmt.Sprintf("%-10s", countStr))
                rowStr := fmt.Sprintf(" %-4s %-30s %-12s %11s  %12s", idxStr, truncPrefix, countPill, totalMem, avgMem)

                if i == m.selected {
                    selStyle := lipgloss.NewStyle().
                        Bold(true).
                        Foreground(cTextBright).
                        Background(lipgloss.Color("#283457"))
                    rows.WriteString(" 👉 " + selStyle.Render(rowStr) + "\n")
                } else {
                    rows.WriteString("    " + lipgloss.NewStyle().Foreground(cTextMain).Render(rowStr) + "\n")
                }
            }

            if len(pageItems) == 0 {
                rows.WriteString("\n   " + lipgloss.NewStyle().Foreground(cTextDim).Render("No namespaces found. Press 's' to seed test data.") + "\n\n")
            }

            pageInfo := lipgloss.NewStyle().Foreground(cTextDim).Render(
                fmt.Sprintf("Page %d of %d  (Namespaces: %d)", m.page+1, m.maxPages(), len(fn)),
            )
            tableHeader := fmt.Sprintf("  %s   %s", lipgloss.NewStyle().Bold(true).Foreground(cSecondary).Render("Aggregated Namespaces"), pageInfo)
            tableContent = fmt.Sprintf("%s\n\n%s\n%s\n%s", tableHeader, th, divider, rows.String())
        }

        // Search Bar Container
        searchIcon := lipgloss.NewStyle().Foreground(cPrimary).Render("🔎 ")
        searchBoxStyle := lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(cBorder).
            Padding(0, 1)

        if m.isSearching {
            searchBoxStyle = searchBoxStyle.BorderForeground(cCyan)
        }
        searchContainer := searchBoxStyle.Render(fmt.Sprintf("%s%s", searchIcon, m.searchInput.View()))

        contentBlock := fmt.Sprintf("%s\n\n%s\n\n%s", tabsBar, searchContainer, tableContent)
        bodyBox = cardStyle.Copy().BorderForeground(cBorderActive).Render(contentBlock)

        if m.confirmBulkDelete {
            warnStyle := lipgloss.NewStyle().Bold(true).Foreground(cTextBright).Background(cCoral).Padding(0, 1)
            footer = fmt.Sprintf("  %s  Are you sure you want to BULK UNLINK all keys in '%s'?  Press 'y' to confirm, or 'n' / 'Esc' to cancel.", warnStyle.Render("⚠️  BULK DELETE"), m.patternToDelete)
        } else if m.confirmDelete {
            warnStyle := lipgloss.NewStyle().Bold(true).Foreground(cTextBright).Background(cCoral).Padding(0, 1)
            footer = fmt.Sprintf("  %s  Are you sure you want to UNLINK '%s'?  Press 'y' to confirm, or 'n' / 'Esc' to cancel.", warnStyle.Render("⚠️  DELETE KEY"), m.keyToDelete)
        } else if m.isSearching {
            footer = renderFooterKeys([]keyHelp{
                {"Enter", "Server SCAN Pattern"},
                {"Esc", "Cancel Filter"},
            })
        } else if m.viewMode == 1 {
            footer = renderFooterKeys([]keyHelp{
                {"Tab", "Top Keys"},
                {"Enter", "Filter Prefix"},
                {"b", "Switch DB"},
                {"/", "Search"},
                {"d", "Bulk Unlink"},
                {"s", "Seed Mock"},
                {"x", "Export JSON"},
                {"q", "Quit"},
            })
        } else {
            footer = renderFooterKeys([]keyHelp{
                {"Tab", "Namespaces"},
                {"Enter", "Inspect Key"},
                {"v", "Raw Value"},
                {"b", "Switch DB"},
                {"/", "Search"},
                {"d", "Unlink Key"},
                {"s", "Seed Mock"},
                {"x", "Export JSON"},
                {"q", "Quit"},
            })
        }
    }

    // --- Status / Toast Notification ---
    var statusBar string
    if m.statusMsg != "" {
        toastStyle := lipgloss.NewStyle().
            Bold(true).
            Foreground(cTextBright).
            Background(cBgPill).
            Border(lipgloss.RoundedBorder()).
            BorderForeground(cAmber).
            Padding(0, 1)
        statusBar = fmt.Sprintf("\n %s\n", toastStyle.Render("🔔 "+m.statusMsg))
    }

    return fmt.Sprintf("\n%s\n\n%s\n\n%s%s\n\n%s\n", header, statsRow, bodyBox, statusBar, footer)
}

// --- Visual Helper Components ---

type keyHelp struct {
    Key  string
    Desc string
}

func renderFooterKeys(helps []keyHelp) string {
    keyPillStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FFFFFF")).
        Background(lipgloss.Color("#24283B")).
        Padding(0, 1)

    descStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#565F89"))

    var parts []string
    for _, h := range helps {
        part := fmt.Sprintf("%s %s", keyPillStyle.Render(h.Key), descStyle.Render(h.Desc))
        parts = append(parts, part)
    }
    return "  " + strings.Join(parts, "   ")
}

func renderTypeBadge(kType string) string {
    base := lipgloss.NewStyle().Bold(true).Padding(0, 1)
    switch strings.ToLower(kType) {
    case "string":
        return base.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#2563EB")).Render("STR")
    case "hash":
        return base.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#D97706")).Render("HASH")
    case "list":
        return base.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7C3AED")).Render("LIST")
    case "set":
        return base.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#059669")).Render("SET")
    case "zset":
        return base.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#DB2777")).Render("ZSET")
    case "stream":
        return base.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0891B2")).Render("STRM")
    default:
        return base.Foreground(lipgloss.Color("#9CA3AF")).Background(lipgloss.Color("#374151")).Render("OTHER")
    }
}

func renderMiniBar(bytes, maxBytes int64, width int) string {
    if maxBytes <= 0 {
        return lipgloss.NewStyle().Foreground(lipgloss.Color("#2E3440")).Render(strings.Repeat("░", width))
    }
    filled := int(float64(bytes) / float64(maxBytes) * float64(width))
    if filled < 1 && bytes > 0 {
        filled = 1
    }
    if filled > width {
        filled = width
    }
    fillColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7"))
    emptyColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#2E3440"))
    return fillColor.Render(strings.Repeat("█", filled)) + emptyColor.Render(strings.Repeat("░", width-filled))
}
