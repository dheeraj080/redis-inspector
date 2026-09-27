package ui

import (
    "fmt"

    "github.com/charmbracelet/lipgloss"

    rclient "redis-inspector/internal/redis"
    "redis-inspector/internal/utils"
)

func (m Model) View() string {
    if m.err != nil {
        return fmt.Sprintf("\n  Error: %v\n\n  Press 'q' to quit.", m.err)
    }

    titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Padding(0, 1)
    boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#5A56E0")).Padding(1, 2)
    headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
    selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Background(lipgloss.Color("#3C3836"))
    normalStyle := lipgloss.NewStyle()
    dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
    tabActiveStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Underline(true)
    statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
    sparkColorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87"))

    header := titleStyle.Render(fmt.Sprintf("⚡ REDIS MEMORY INSPECTOR  [DB%d]", m.currentDB))

    sparkline := utils.RenderSparkline(m.memHistory)
    if sparkline == "" {
        sparkline = "collecting..."
    }
    renderedSparkline := sparkColorStyle.Render(sparkline)

    overviewText := fmt.Sprintf(
        "Active DB:   db%-10d Peak Memory: %s\nUsed Memory: %-12s Frag Ratio:  %s\nTrend (30s): [%s]",
        m.currentDB, m.peakMem, m.usedMem, m.fragRatio, renderedSparkline,
    )
    overviewBox := boxStyle.Render(overviewText)

    var keysBox string
    var footer string

    if m.showEditModal {
        editBoxStyle := boxStyle.Copy().BorderForeground(lipgloss.Color("#00FF87"))
        editHeader := headerStyle.Render("✏️ EDIT KEY VALUE")

        promptText := fmt.Sprintf(
            "Key: %s (%s)\n\nEnter new content below:\n\n %s",
            m.activeDetail.Key, m.activeDetail.Type, m.editorInput.View(),
        )

        keysBox = editBoxStyle.Render(fmt.Sprintf("%s\n\n%s", editHeader, promptText))
        footer = "  Press 'Enter' to save update, or 'Esc' to cancel."

    } else if m.showDBModal {
        dbBoxStyle := boxStyle.Copy().BorderForeground(lipgloss.Color("#FF007F"))
        dbHeader := headerStyle.Render("🗄️ SELECT REDIS DATABASE (0-15)")

        var dbRows string
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

            rowStr := fmt.Sprintf("%sDB%02d: %-8d keys (%d with expiry)", activeMarker, i, keysCount, expiresCount)

            if i == m.dbSelected {
                dbRows += selectedStyle.Render("👉 "+rowStr) + "\n"
            } else {
                dbRows += normalStyle.Render("   "+rowStr) + "\n"
            }
        }

        keysBox = dbBoxStyle.Render(fmt.Sprintf("%s\n\n%s", dbHeader, dbRows))
        footer = "  '↑/↓' navigate  •  'Enter' select DB  •  'Esc' cancel"

    } else if m.showValueViewer {
        valBoxStyle := boxStyle.Copy().BorderForeground(lipgloss.Color("#FF007F"))
        valHeader := headerStyle.Render("📄 RAW VALUE VIEWER")

        keysBox = valBoxStyle.Render(fmt.Sprintf("%s\n\n%s", valHeader, m.activeValue))
        footer = "  Press 'Esc' or 'q' to return."

    } else if m.showTTLModal {
        ttlBoxStyle := boxStyle.Copy().BorderForeground(lipgloss.Color("#FFD700"))
        ttlHeader := headerStyle.Render("⏱️ SET KEY TTL (EXPIRE)")

        promptText := fmt.Sprintf(
            "Key: %s\nCurrent TTL: %s\n\nEnter new TTL in seconds (e.g., 3600 for 1h, -1 to PERSIST):\n\n %s",
            m.activeDetail.Key, m.activeDetail.TTL, m.ttlInput.View(),
        )

        keysBox = ttlBoxStyle.Render(fmt.Sprintf("%s\n\n%s", ttlHeader, promptText))
        footer = "  Press 'Enter' to confirm, or 'Esc' to cancel."

    } else if m.showDetails {
        detailBoxStyle := boxStyle.Copy().BorderForeground(lipgloss.Color("#00FFFF"))
        detailHeader := headerStyle.Render("🔍 KEY INSPECTOR")

        countLabel := "Length / Items:"
        if m.activeDetail.Type == "string" {
            countLabel = "String Length (bytes):"
        }

        detailText := fmt.Sprintf(
            "Key Name:      %s\n"+
                "Data Type:     %s\n"+
                "Encoding:      %s\n"+
                "Memory Usage:  %s (%d bytes)\n"+
                "Time to Live:  %s\n"+
                "%s %d",
            m.activeDetail.Key,
            m.activeDetail.Type,
            m.activeDetail.Encoding,
            utils.FormatBytes(m.activeDetail.Bytes), m.activeDetail.Bytes,
            m.activeDetail.TTL,
            countLabel, m.activeDetail.Elements,
        )

        keysBox = detailBoxStyle.Render(fmt.Sprintf("%s\n\n%s", detailHeader, detailText))
        footer = "  'v' view content  •  'e' edit  •  't' TTL  •  'Esc' back"

    } else {
        tab1 := dimStyle.Render("1. Top Keys")
        tab2 := dimStyle.Render("2. Namespaces")
        if m.viewMode == 0 {
            tab1 = tabActiveStyle.Render("1. Top Keys")
        } else {
            tab2 = tabActiveStyle.Render("2. Namespaces")
        }
        tabs := fmt.Sprintf("View Mode:  %s  |  %s", tab1, tab2)

        var listRows string
        var titleSection string

        if m.viewMode == 0 {
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

            for i, k := range pageItems {
                truncatedKey := utils.TruncateString(k.Key, 28)
                formattedMem := utils.FormatBytes(k.Bytes)
                rowStr := fmt.Sprintf(" %2d. %-28s %-10s %10s ", start+i+1, truncatedKey, fmt.Sprintf("[%s]", k.Type), formattedMem)

                if i == m.selected {
                    listRows += selectedStyle.Render("👉 "+rowStr) + "\n"
                } else {
                    listRows += normalStyle.Render("   "+rowStr) + "\n"
                }
            }

            if len(pageItems) == 0 {
                listRows = "  No matching keys found in DB" + fmt.Sprintf("%d", m.currentDB) + ". Press 's' to seed mock test data.\n"
            }

            pageInfo := dimStyle.Render(fmt.Sprintf("Page %d of %d (Total Keys: %d)", m.page+1, m.maxPages(), len(fk)))
            titleSection = fmt.Sprintf("%s   %s", headerStyle.Render("Top Memory Keys:"), pageInfo)

        } else {
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

            for i, ns := range pageItems {
                truncatedPrefix := utils.TruncateString(ns.Prefix, 26)
                totalMem := utils.FormatBytes(ns.Bytes)
                avgMem := "0 B"
                if ns.Count > 0 {
                    avgMem = utils.FormatBytes(ns.Bytes / int64(ns.Count))
                }
                keysCountStr := fmt.Sprintf("(%d keys)", ns.Count)

                rowStr := fmt.Sprintf(" %2d. %-26s %-10s %10s  [%s/key] ", start+i+1, truncatedPrefix, keysCountStr, totalMem, avgMem)

                if i == m.selected {
                    listRows += selectedStyle.Render("👉 "+rowStr) + "\n"
                } else {
                    listRows += normalStyle.Render("   "+rowStr) + "\n"
                }
            }

            if len(pageItems) == 0 {
                listRows = "  No matching namespaces found in DB" + fmt.Sprintf("%d", m.currentDB) + ". Press 's' to seed mock test data.\n"
            }

            pageInfo := dimStyle.Render(fmt.Sprintf("Page %d of %d (Namespaces: %d)", m.page+1, m.maxPages(), len(fn)))
            titleSection = fmt.Sprintf("%s   %s", headerStyle.Render("Aggregated Namespaces:"), pageInfo)
        }

        searchBar := fmt.Sprintf(" Filter / Scan: %s", m.searchInput.View())

        keysBox = boxStyle.Render(
            fmt.Sprintf("%s\n\n%s\n%s\n\n%s", tabs, titleSection, searchBar, listRows),
        )

        if m.confirmBulkDelete {
            warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
            footer = fmt.Sprintf("  %s Are you sure you want to BULK UNLINK all keys in '%s'? (y/N)", warnStyle.Render("⚠️  CONFIRM BULK DELETE:"), m.patternToDelete)
        } else if m.confirmDelete {
            warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
            footer = fmt.Sprintf("  %s Are you sure you want to UNLINK '%s'? (y/N)", warnStyle.Render("⚠️  CONFIRM:"), m.keyToDelete)
        } else if m.isSearching {
            footer = "  Type filter. Press 'Enter' to SCAN Redis server, or 'Esc' to cancel."
        } else if m.viewMode == 1 {
            footer = "  'b' select DB  •  'Tab' switch view  •  '/' filter  •  'd' bulk delete  •  's' seed  •  'q' quit"
        } else {
            footer = "  'b' select DB  •  'Tab' switch view  •  '/' filter  •  'v' view content  •  'd' delete  •  'q' quit"
        }
    }

    statusBar := ""
    if m.statusMsg != "" {
        statusBar = fmt.Sprintf("\n%s\n", statusStyle.Render("  🔔 "+m.statusMsg))
    }

    return fmt.Sprintf("\n%s\n\n%s\n\n%s%s\n\n%s\n", header, overviewBox, keysBox, statusBar, footer)
}
