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
	// Guard against initial render before WindowSizeMsg is received
	if m.height == 0 {
		return "Initializing UI..."
	}

	// Minimum width guard to prevent broken layout on narrow terminals
	if m.width < 60 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F7768E")).
			Padding(1, 2).
			Render(
				"! Terminal too narrow!\n\n" +
					"Please widen your terminal to at least 60 columns.\n" +
					"Current width: " + fmt.Sprintf("%d", m.width) + " columns\n\n" +
					"Press 'q' to quit.",
			)
	}

	if m.err != nil {
		errBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F7768E")).
			Padding(1, 2)

		errMsg := fmt.Sprintf(
			"! Redis Connection Error:\n\n  %v\n\n  Please verify your Redis server is running and connection parameters are correct.\n  Press 'q' or 'Ctrl+C' to quit.",
			m.err,
		)
		return errBoxStyle.Render(errMsg)
	}

	// --- Color Palette ---
	cPrimary := lipgloss.Color("#7AA2F7")
	cSecondary := lipgloss.Color("#BB9AF7")
	cMint := lipgloss.Color("#73DACA")
	cGreen := lipgloss.Color("#9ECE6A")
	cCoral := lipgloss.Color("#F7768E")
	cAmber := lipgloss.Color("#E0AF68")
	cCyan := lipgloss.Color("#7DCFFF")
	cBorder := lipgloss.Color("#3B4261")
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

	// Dynamically calculate content width to avoid line wraps.
	totalWidth := m.width
	if totalWidth <= 0 {
		totalWidth = 84
	}
	boxInnerWidth := totalWidth - 6
	if boxInnerWidth < 1 {
		boxInnerWidth = 1
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

	// Replaced emoji with ASCII for robust width calculation
	headerLeft := brandStyle.Render("* REDIS INSPECTOR")
	dbPill := dbPillStyle.Render(fmt.Sprintf("DB %d", m.currentDB))
	statusPill := statusPillStyle.Render("* ONLINE")
	hostPill := hostPillStyle.Render(hostAddr)

	header := renderHeader(
		totalWidth,
		headerLeft,
		dbPill,
		statusPill,
		hostPill,
	)

	// --- 2. Live Metrics ---
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
	renderedSparkline := lipgloss.NewStyle().
		Bold(true).
		Foreground(cMint).
		Render(sparkline)

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

	metricLabelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cTextDim)
	metricValStyle := lipgloss.NewStyle().
		Bold(true)
	subLabelStyle := lipgloss.NewStyle().
		Foreground(cTextDim)

	metricColumns := metricColumnCountForHeight(boxInnerWidth, m.height)

	useCompactMetrics := false
	if m.height > 0 {
		switch {
		case m.height < 32:
			useCompactMetrics = true
		case metricColumns == 1 && m.height < 45:
			useCompactMetrics = true
		}
	}

	var statsRow string
	if useCompactMetrics {
		statsRow = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1).
			Render(
				renderCompactMetrics(
					usedMemStr,
					peakMemStr,
					fragStr,
					fragStatus,
					fragColor,
				),
			)
	} else {
		cardGap := 2
		cardWidth := boxInnerWidth
		if metricColumns > 1 {
			cardWidth = (boxInnerWidth - (metricColumns-1)*cardGap) / metricColumns
		}
		if cardWidth < 1 {
			cardWidth = 1
		}

		// Prevent text wrapping inside metric cards by strictly truncating content
		maxContentWidth := cardWidth - 4
		if maxContentWidth < 1 {
			maxContentWidth = 1
		}

		subLabel1 := utils.TruncateString(fmt.Sprintf("%d B", m.currentStats().UsedMemBytes), maxContentWidth)
		subLabel2 := utils.TruncateString("Alloc: "+utils.TruncateString(allocatorStr, 8), maxContentWidth)
		truncSparkline := utils.TruncateString(renderedSparkline, maxContentWidth)
		truncTelemetry := utils.TruncateString("Telemetry", maxContentWidth)

		card1 := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1).
			Width(cardWidth).
			Render(
				fmt.Sprintf(
					"%s\n%s\n%s",
					metricLabelStyle.Render("USED MEM"),
					metricValStyle.Foreground(cCyan).Render(usedMemStr),
					subLabelStyle.Render(subLabel1),
				),
			)
		card2 := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1).
			Width(cardWidth).
			Render(
				fmt.Sprintf(
					"%s\n%s\n%s",
					metricLabelStyle.Render("PEAK MEM"),
					metricValStyle.Foreground(cSecondary).Render(peakMemStr),
					subLabelStyle.Render(subLabel2),
				),
			)
		card3 := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1).
			Width(cardWidth).
			Render(
				fmt.Sprintf(
					"%s\n%s\n%s",
					metricLabelStyle.Render("FRAG RATIO"),
					metricValStyle.Foreground(fragColor).Render(fragStr),
					lipgloss.NewStyle().
						Foreground(fragColor).
						Render("* "+fragStatus),
				),
			)
		card4 := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1).
			Width(cardWidth).
			Render(
				fmt.Sprintf(
					"%s\n%s\n%s",
					metricLabelStyle.Render("30s TREND"),
					truncSparkline,
					subLabelStyle.Render(truncTelemetry),
				),
			)

		statsGap := strings.Repeat(" ", cardGap)
		switch metricColumns {
		case 1:
			statsRow = lipgloss.JoinVertical(
				lipgloss.Left,
				card1,
				card2,
				card3,
				card4,
			)
		case 2:
			firstRow := lipgloss.JoinHorizontal(
				lipgloss.Top,
				card1,
				statsGap,
				card2,
			)
			secondRow := lipgloss.JoinHorizontal(
				lipgloss.Top,
				card3,
				statsGap,
				card4,
			)
			statsRow = lipgloss.JoinVertical(
				lipgloss.Left,
				firstRow,
				secondRow,
			)
		default:
			statsRow = lipgloss.JoinHorizontal(
				lipgloss.Top,
				card1,
				statsGap,
				card2,
				statsGap,
				card3,
				statsGap,
				card4,
			)
		}
	}

	// --- 3. Body Content / Modals ---
	var bodyBox string
	var footer string

	if m.showEditModal {
		editBoxStyle := cardStyle.Copy().
			BorderForeground(cMint).
			Padding(1, 2)
		editHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(cMint).
			Render("Edit KEY VALUE")
		typeBadge := renderTypeBadge(m.activeDetail.Type)
		promptText := fmt.Sprintf(
			"Key: %s  %s\n\nEnter new content below:\n\n%s",
			lipgloss.NewStyle().
				Bold(true).
				Foreground(cTextBright).
				Render(m.activeDetail.Key),
			typeBadge,
			m.editorInput.View(),
		)
		bodyBox = editBoxStyle.Render(
			fmt.Sprintf("%s\n\n%s", editHeader, promptText),
		)
		footer = renderFooterKeys([]keyHelp{
			{"Enter", "Save Update"},
			{"Esc", "Cancel"},
		})
	} else if m.showDBModal {
		dbBoxStyle := cardStyle.Copy().
			BorderForeground(cSecondary).
			Padding(1, 2)
		dbHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(cSecondary).
			Render("DB: SELECT REDIS DATABASE (DB 0 - 15)")
		var col1, col2 strings.Builder
		for i := 0; i < 16; i++ {
			var keysCount int64
			for _, db := range m.databases {
				if db.DB == i {
					keysCount = db.Keys
					break
				}
			}
			activeMarker := "  "
			if i == m.currentDB {
				activeMarker = "* "
			}
			rowText := fmt.Sprintf(
				"%sDB %02d : %6d keys",
				activeMarker,
				i,
				keysCount,
			)
			var line string
			if i == m.dbSelected {
				line = lipgloss.NewStyle().
					Bold(true).
					Foreground(cCyan).
					Background(lipgloss.Color("#283457")).
					Render(" > " + rowText + " ")
			} else if i == m.currentDB {
				line = lipgloss.NewStyle().
					Bold(true).
					Foreground(cMint).
					Render("    " + rowText)
			} else {
				line = lipgloss.NewStyle().
					Foreground(cTextMain).
					Render("    " + rowText)
			}
			if i < 8 {
				col1.WriteString(line + "\n")
			} else {
				col2.WriteString(line + "\n")
			}
		}
		dbGrid := lipgloss.JoinHorizontal(
			lipgloss.Top,
			col1.String(),
			"   ",
			col2.String(),
		)
		bodyBox = dbBoxStyle.Render(
			fmt.Sprintf("%s\n\n%s", dbHeader, dbGrid),
		)
		footer = renderFooterKeys([]keyHelp{
			{"Up/Down/Left/Right", "Navigate DBs"},
			{"Enter", "Switch DB"},
			{"Esc", "Cancel"},
		})
	} else if m.showValueViewer {
		valBoxStyle := cardStyle.Copy().
			BorderForeground(cCyan).
			Padding(1, 2)
		typeBadge := renderTypeBadge(m.activeDetail.Type)
		valHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(cCyan).
			Render(
				fmt.Sprintf(
					"Val: RAW VALUE VIEWER: %s %s",
					m.activeDetail.Key,
					typeBadge,
				),
			)
		bodyBox = valBoxStyle.Render(
			fmt.Sprintf("%s\n\n%s", valHeader, m.activeValue),
		)
		footer = renderFooterKeys([]keyHelp{
			{"Esc / q", "Return to List"},
		})
	} else if m.showTTLModal {
		ttlBoxStyle := cardStyle.Copy().
			BorderForeground(cAmber).
			Padding(1, 2)
		ttlHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(cAmber).
			Render("TTL: CONFIGURE KEY EXPIRATION")
		promptText := fmt.Sprintf(
			"Key: %s\nCurrent TTL: %s\n\nEnter new TTL in seconds (-1 to PERSIST):\n\n%s",
			lipgloss.NewStyle().
				Bold(true).
				Foreground(cTextBright).
				Render(m.activeDetail.Key),
			lipgloss.NewStyle().
				Foreground(cCyan).
				Render(m.activeDetail.TTL),
			m.ttlInput.View(),
		)
		bodyBox = ttlBoxStyle.Render(
			fmt.Sprintf("%s\n\n%s", ttlHeader, promptText),
		)
		footer = renderFooterKeys([]keyHelp{
			{"Enter", "Confirm TTL"},
			{"Esc", "Cancel"},
		})
	} else if m.showDetails {
		detailBoxStyle := cardStyle.Copy().
			BorderForeground(cPrimary).
			Padding(1, 2)
		detailHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(cPrimary).
			Render("Key: KEY INSPECTOR")
		countLabel := "Items / Elements:"
		if m.activeDetail.Type == "string" {
			countLabel = "String Length (bytes):"
		}
		typeBadge := renderTypeBadge(m.activeDetail.Type)
		labelStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(cTextDim).
			Width(22)
		valStyle := lipgloss.NewStyle().
			Foreground(cTextBright)
		grid := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s (%d bytes)\n%s %s\n%s %d",
			labelStyle.Render("Name: Key Name:"),
			valStyle.Render(m.activeDetail.Key),
			labelStyle.Render("Type: Data Type:"),
			typeBadge,
			labelStyle.Render("Enc: Encoding:"),
			valStyle.Render(m.activeDetail.Encoding),
			labelStyle.Render("Mem: Memory Footprint:"),
			lipgloss.NewStyle().
				Bold(true).
				Foreground(cCyan).
				Render(utils.FormatBytes(m.activeDetail.Bytes)),
			m.activeDetail.Bytes,
			labelStyle.Render("TTL: Time to Live:"),
			lipgloss.NewStyle().
				Foreground(cAmber).
				Render(m.activeDetail.TTL),
			labelStyle.Render("Cnt: "+countLabel),
			m.activeDetail.Elements,
		)
		bodyBox = detailBoxStyle.Render(
			fmt.Sprintf("%s\n\n%s", detailHeader, grid),
		)
		actions := []keyHelp{
			{"v", "Raw Value"},
			{"t", "Set TTL"},
		}
		if m.activeDetail.Type == "string" ||
			m.activeDetail.Type == "hash" {
			actions = append(
				actions,
				keyHelp{"e", "Edit Value"},
			)
		}
		actions = append(actions, keyHelp{"Esc", "Back"})
		footer = renderFooterKeys(actions)
	} else {
		// --- Tab Bar ---
		tab1Text := fmt.Sprintf(
			"1. Top Keys (%d)",
			len(m.filteredKeys()),
		)
		tab2Text := fmt.Sprintf(
			"2. Namespaces (%d)",
			len(m.filteredNamespaces()),
		)
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
		tabsBar := fmt.Sprintf(
			" %s  %s",
			tab1,
			tab2,
		)

		// --- Tables ---
		var tableContent string
		keyColWidth := tableKeyColumnWidth(boxInnerWidth)
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
			var maxBytes int64 = 1
			if len(pageItems) > 0 &&
				pageItems[0].Bytes > 0 {
				maxBytes = pageItems[0].Bytes
			}
			thStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(cTextDim)
			var th string
			if boxInnerWidth < 45 {
				th = fmt.Sprintf(
					" %-*s %10s",
					keyColWidth,
					thStyle.Render("KEY NAME"),
					thStyle.Render("MEMORY"),
				)
			} else if boxInnerWidth < 70 {
				th = fmt.Sprintf(
					" %-*s %-6s %10s",
					keyColWidth,
					thStyle.Render("KEY NAME"),
					thStyle.Render("TYPE"),
					thStyle.Render("MEMORY"),
				)
			} else {
				th = fmt.Sprintf(
					"   %-4s %-*s %-6s %10s  %-8s",
					thStyle.Render("#"),
					keyColWidth,
					thStyle.Render("KEY NAME"),
					thStyle.Render("TYPE"),
					thStyle.Render("MEMORY"),
					thStyle.Render("BAR"),
				)
			}
			divider := lipgloss.NewStyle().
				Foreground(cBorder).
				Render(
					"  " + strings.Repeat(
						"-",
						maxInt(1, boxInnerWidth-4),
					),
				)
			var rows strings.Builder
			for i, k := range pageItems {
				idxStr := fmt.Sprintf(
					"%2d.",
					start+i+1,
				)
				truncKey := utils.TruncateString(
					k.Key,
					keyColWidth,
				)
				typeBadge := renderTypeBadge(k.Type)
				memStr := utils.FormatBytes(k.Bytes)
				miniBar := renderMiniBar(
					k.Bytes,
					maxBytes,
					8,
				)
				var rowStr string
				if boxInnerWidth < 45 {
					rowStr = fmt.Sprintf(
						" %-*s %10s",
						keyColWidth,
						truncKey,
						memStr,
					)
				} else if boxInnerWidth < 70 {
					rowStr = fmt.Sprintf(
						" %-*s %s  %10s",
						keyColWidth,
						truncKey,
						typeBadge,
						memStr,
					)
				} else {
					rowStr = fmt.Sprintf(
						" %-4s %-*s %s  %10s  %s",
						idxStr,
						keyColWidth,
						truncKey,
						typeBadge,
						memStr,
						miniBar,
					)
				}
				if i == m.selected {
					selStyle := lipgloss.NewStyle().
						Bold(true).
						Foreground(cTextBright).
						Background(lipgloss.Color("#283457"))
					rows.WriteString(
						"  > " +
							selStyle.Render(rowStr) +
							"\n",
					)
				} else {
					rows.WriteString(
						"    " +
							lipgloss.NewStyle().
								Foreground(cTextMain).
								Render(rowStr) +
							"\n",
					)
				}
			}
			if len(pageItems) == 0 {
				rows.WriteString(
					"\n   " +
						lipgloss.NewStyle().
							Foreground(cTextDim).
							Render("No matching keys. Press 's' to seed test data.") +
						"\n\n",
				)
			}
			pageInfo := lipgloss.NewStyle().
				Foreground(cTextDim).
				Render(
					fmt.Sprintf(
						"Page %d of %d  (Keys: %d)",
						m.page+1,
						m.maxPages(),
						len(fk),
					),
				)
			if m.truncated {
				pageInfo += " " +
					lipgloss.NewStyle().
						Bold(true).
						Foreground(cAmber).
						Render("! Sampled top 500")
			}
			tableHeader := fmt.Sprintf(
				"  %s   %s",
				lipgloss.NewStyle().
					Bold(true).
					Foreground(cCyan).
					Render("Top Memory Consumers"),
				pageInfo,
			)
			tableContent = fmt.Sprintf(
				"%s\n%s\n%s\n%s",
				tableHeader,
				th,
				divider,
				rows.String(),
			)
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
			thStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(cTextDim)
			var th string
			if boxInnerWidth < 55 {
				th = fmt.Sprintf(
					" %-*s %10s",
					keyColWidth,
					thStyle.Render("NAMESPACE"),
					thStyle.Render("TOTAL MEM"),
				)
			} else {
				th = fmt.Sprintf(
					"   %-4s %-*s %-10s %10s  %10s",
					thStyle.Render("#"),
					keyColWidth,
					thStyle.Render("NAMESPACE"),
					thStyle.Render("COUNT"),
					thStyle.Render("TOTAL MEM"),
					thStyle.Render("AVG/KEY"),
				)
			}
			divider := lipgloss.NewStyle().
				Foreground(cBorder).
				Render(
					"  " + strings.Repeat(
						"-",
						maxInt(1, boxInnerWidth-4),
					),
				)
			var rows strings.Builder
			for i, ns := range pageItems {
				idxStr := fmt.Sprintf(
					"%2d.",
					start+i+1,
				)
				truncPrefix := utils.TruncateString(
					ns.Prefix,
					keyColWidth,
				)
				countStr := fmt.Sprintf(
					"%d keys",
					ns.Count,
				)
				totalMem := utils.FormatBytes(ns.Bytes)
				avgMem := "0 B"
				if ns.Count > 0 {
					avgMem = utils.FormatBytes(
						ns.Bytes / int64(ns.Count),
					)
				}
				var rowStr string
				if boxInnerWidth < 55 {
					rowStr = fmt.Sprintf(
						" %-*s %10s",
						keyColWidth,
						truncPrefix,
						totalMem,
					)
				} else {
					rowStr = fmt.Sprintf(
						" %-4s %-*s %-10s %10s  %10s",
						idxStr,
						keyColWidth,
						truncPrefix,
						countStr,
						totalMem,
						avgMem,
					)
				}
				if i == m.selected {
					selStyle := lipgloss.NewStyle().
						Bold(true).
						Foreground(cTextBright).
						Background(lipgloss.Color("#283457"))
					rows.WriteString(
						"  > " +
							selStyle.Render(rowStr) +
							"\n",
					)
				} else {
					rows.WriteString(
						"    " +
							lipgloss.NewStyle().
								Foreground(cTextMain).
								Render(rowStr) +
							"\n",
					)
				}
			}
			if len(pageItems) == 0 {
				rows.WriteString(
					"\n   " +
						lipgloss.NewStyle().
							Foreground(cTextDim).
							Render("No namespaces. Press 's' to seed test data.") +
						"\n\n",
				)
			}
			pageInfo := lipgloss.NewStyle().
				Foreground(cTextDim).
				Render(
					fmt.Sprintf(
						"Page %d of %d  (Namespaces: %d)",
						m.page+1,
						m.maxPages(),
						len(fn),
					),
				)
			tableHeader := fmt.Sprintf(
				"  %s   %s",
				lipgloss.NewStyle().
					Bold(true).
					Foreground(cSecondary).
					Render("Aggregated Namespaces"),
				pageInfo,
			)
			tableContent = fmt.Sprintf(
				"%s\n%s\n%s\n%s",
				tableHeader,
				th,
				divider,
				rows.String(),
			)
		}

		searchIcon := lipgloss.NewStyle().
			Foreground(cPrimary).
			Render("> ")
		searchBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1)
		if m.isSearching {
			searchBoxStyle = searchBoxStyle.
				BorderForeground(cCyan)
		}
		searchContainer := searchBoxStyle.Render(
			fmt.Sprintf(
				"%s%s",
				searchIcon,
				m.searchInput.View(),
			),
		)
		contentBlock := fmt.Sprintf(
			"%s\n%s\n%s",
			tabsBar,
			searchContainer,
			tableContent,
		)
		bodyBox = cardStyle.Copy().
			BorderForeground(cBorderActive).
			Render(contentBlock)

		if m.confirmBulkDelete {
			warnStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(cTextBright).
				Background(cCoral).
				Padding(0, 1)
			footer = fmt.Sprintf(
				"  %s  UNLINK all keys in '%s'? Press 'y' to confirm, or 'n' / 'Esc' to cancel.",
				warnStyle.Render("! BULK DELETE"),
				m.patternToDelete,
			)
		} else if m.confirmDelete {
			warnStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(cTextBright).
				Background(cCoral).
				Padding(0, 1)
			footer = fmt.Sprintf(
				"  %s  UNLINK '%s'? Press 'y' to confirm, or 'n' / 'Esc' to cancel.",
				warnStyle.Render("! DELETE KEY"),
				m.keyToDelete,
			)
		} else if m.isSearching {
			footer = renderFooterKeys([]keyHelp{
				{"Enter", "SCAN Pattern"},
				{"Esc", "Cancel"},
			})
		} else if m.viewMode == 1 {
			footer = renderFooterKeys([]keyHelp{
				{"Tab", "Top Keys"},
				{"Enter", "Filter"},
				{"b", "DB"},
				{"/", "Search"},
				{"d", "Bulk Delete"},
				{"q", "Quit"},
			})
		} else {
			footer = renderFooterKeys([]keyHelp{
				{"Tab", "Namespaces"},
				{"Enter", "Inspect"},
				{"v", "Value"},
				{"b", "DB"},
				{"/", "Search"},
				{"d", "Delete"},
				{"q", "Quit"},
			})
		}
	}

	// --- Toast Notification Bar ---
	var statusBar string
	if m.statusMsg != "" {
		toastStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(cTextBright).
			Background(cBgPill).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cAmber).
			Padding(0, 1)
		statusBar = fmt.Sprintf(
			" %s\n",
			toastStyle.Render("! "+m.statusMsg),
		)
	}

	output := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n%s%s",
		header,
		statsRow,
		bodyBox,
		statusBar,
		footer,
	)

	output = fitTerminalWidth(output, totalWidth)

	// --- HARD HEIGHT CLAMP WITH SAFETY MARGIN ---
	// Guarantees the output never exceeds terminal height, preventing the
	// terminal from scrolling and hiding the top header/metrics or cutting off the bottom.
	if m.height > 0 {
		lines := strings.Split(output, "\n")
		// Remove trailing empty line artifact from Split if it exists
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		// Leave a 1-line safety margin at the bottom to ensure the footer
		// is never cut off by the terminal emulator's scroll behavior.
		maxLines := m.height
		if maxLines > 1 {
			maxLines--
		}

		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}
		output = strings.Join(lines, "\n")
	}

	return output
}

// Helper Components
func metricColumnCount(width int) int {
	switch {
	case width >= 70:
		return 4
	case width >= 34:
		return 2
	default:
		return 1
	}
}

func metricColumnCountForHeight(width, height int) int {
	columns := metricColumnCount(width)
	if height <= 0 {
		return columns
	}
	if height < 30 && columns == 1 && width >= 34 {
		return 2
	}
	return columns
}

func renderCompactMetrics(
	usedMem string,
	peakMem string,
	frag string,
	fragStatus string,
	fragColor lipgloss.Color,
) string {
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#565F89"))
	valueStyle := lipgloss.NewStyle().
		Bold(true)
	fragStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(fragColor)

	return strings.Join(
		[]string{
			fmt.Sprintf(
				"%s %s",
				labelStyle.Render("MEM"),
				valueStyle.
					Foreground(lipgloss.Color("#7DCFFF")).
					Render(usedMem),
			),
			fmt.Sprintf(
				"%s %s",
				labelStyle.Render("PEAK"),
				valueStyle.
					Foreground(lipgloss.Color("#BB9AF7")).
					Render(peakMem),
			),
			fmt.Sprintf(
				"%s %s",
				labelStyle.Render("FRAG"),
				fragStyle.Render(frag),
			),
			fmt.Sprintf(
				"%s %s",
				labelStyle.Render("STATUS"),
				fragStyle.Render(fragStatus),
			),
		},
		"  ",
	)
}

func tableKeyColumnWidth(width int) int {
	// More conservative width calculation to prevent right-edge truncation
	switch {
	case width >= 70:
		return maxInt(20, width-42)
	case width >= 45:
		return maxInt(8, width-32)
	default:
		return maxInt(4, width-16)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderHeader(width int, parts ...string) string {
	if width <= 0 {
		return strings.Join(parts, "  ")
	}
	for len(parts) > 1 {
		line := " " + strings.Join(parts, "  ")
		if lipgloss.Width(line) <= width {
			return line
		}
		parts = parts[:len(parts)-1]
	}
	return truncatePlain(
		" "+stripANSI(parts[0]),
		width,
	)
}

func truncatePlain(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	if width <= 3 {
		return string(
			runes[:minInt(width, len(runes))],
		)
	}
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-3]) + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func fitTerminalWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			lines[i] = truncatePlain(
				stripANSI(line),
				width,
			)
		}
	}
	return strings.Join(lines, "\n")
}

func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if inEscape {
			if (s[i] >= 'a' && s[i] <= 'z') ||
				(s[i] >= 'A' && s[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		if s[i] == 0x1b {
			inEscape = true
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

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
		part := fmt.Sprintf(
			"%s %s",
			keyPillStyle.Render(h.Key),
			descStyle.Render(h.Desc),
		)
		parts = append(parts, part)
	}
	return "  " + strings.Join(parts, "   ")
}

func renderTypeBadge(kType string) string {
	base := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)
	switch strings.ToLower(kType) {
	case "string":
		return base.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2563EB")).
			Render("STR ")
	case "hash":
		return base.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#D97706")).
			Render("HASH")
	case "list":
		return base.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Render("LIST")
	case "set":
		return base.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#059669")).
			Render("SET ")
	case "zset":
		return base.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#DB2777")).
			Render("ZSET")
	case "stream":
		return base.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0891B2")).
			Render("STRM")
	default:
		return base.
			Foreground(lipgloss.Color("#9CA3AF")).
			Background(lipgloss.Color("#374151")).
			Render("OTHR")
	}
}

func renderMiniBar(bytes, maxBytes int64, width int) string {
	if maxBytes <= 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2E3440")).
			Render(strings.Repeat("-", width))
	}
	filled := int(
		float64(bytes) /
			float64(maxBytes) *
			float64(width),
	)
	if filled < 1 && bytes > 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	fillColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7AA2F7"))
	emptyColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2E3440"))
	return fillColor.Render(
		strings.Repeat("#", filled),
	) + emptyColor.Render(
		strings.Repeat("-", width-filled),
	)
}
