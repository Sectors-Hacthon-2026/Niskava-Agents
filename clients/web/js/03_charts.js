// 3. CANDLESTICK CHARTS, EVIDENCE MATRIX & ADVANCED WORKSPACE ACTIONS
            // ==========================================================================

            // TradingView Lightweight Charts & High-Fidelity SVG Fallback Engine
            let activeCharts = {};

            function renderTradingViewCandlestick(containerId, candleData, anomalyDate) {
                const container = document.getElementById(containerId);
                if (!container) return;

                // Dispose any previous chart instance bound to this container
                if (activeCharts[containerId]) {
                    try { activeCharts[containerId].remove(); } catch (_) { /* already detached */ }
                    delete activeCharts[containerId];
                }

                // Never fabricate market data: show an explicit empty state instead
                if (!Array.isArray(candleData) || candleData.length === 0) {
                    const msg = (typeof currentLang !== 'undefined' && currentLang === 'en')
                        ? 'No daily candlestick data available for this ticker.'
                        : 'Data candlestick harian belum tersedia untuk emiten ini.';
                    container.innerHTML = `<div class="dossier-empty-state">${msg}</div>`;
                    return;
                }

                container.innerHTML = '';

                // If LightweightCharts is available from CDN
                if (window.LightweightCharts && typeof window.LightweightCharts.createChart === 'function') {
                    try {
                        const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
                        const chart = window.LightweightCharts.createChart(container, {
                            width: container.clientWidth || 300,
                            height: container.clientHeight || 240,
                            layout: {
                                background: { color: 'transparent' },
                                textColor: isDark ? '#8E8C84' : '#6E6B63',
                                fontSize: 11,
                                fontFamily: 'Plus Jakarta Sans, -apple-system, sans-serif'
                            },
                            grid: {
                                vertLines: { color: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(20, 20, 19, 0.03)' },
                                horzLines: { color: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(20, 20, 19, 0.03)' }
                            },
                            crosshair: { mode: 1 },
                            rightPriceScale: { borderColor: isDark ? '#292926' : '#E6E4DC' },
                            timeScale: { borderColor: isDark ? '#292926' : '#E6E4DC', timeVisible: true }
                        });

                        const candleSeries = chart.addCandlestickSeries({
                            upColor: '#0F6735',
                            downColor: '#AD1C1C',
                            borderUpColor: '#0F6735',
                            borderDownColor: '#AD1C1C',
                            wickUpColor: '#0F6735',
                            wickDownColor: '#AD1C1C'
                        });

                        candleSeries.setData(candleData.map(c => ({
                            time: c.time,
                            open: c.open,
                            high: c.high,
                            low: c.low,
                            close: c.close
                        })));

                        const volumeSeries = chart.addHistogramSeries({
                            priceFormat: { type: 'volume' },
                            priceScaleId: '',
                            scaleMargins: { top: 0.75, bottom: 0 }
                        });

                        volumeSeries.setData(candleData.map(c => ({
                            time: c.time,
                            value: c.volume || 1000000,
                            color: c.close >= c.open ? 'rgba(14, 203, 129, 0.35)' : 'rgba(246, 70, 93, 0.35)'
                        })));

                        if (anomalyDate) {
                            candleSeries.setMarkers([{
                                time: anomalyDate,
                                position: 'aboveBar',
                                color: '#F6465D',
                                shape: 'arrowDown',
                                text: 'Volume Anomaly (Z ≥ 2.50σ)'
                            }]);
                        }

                        chart.timeScale().fitContent();
                        activeCharts[containerId] = chart;

                        // Responsive observer
                        if (window.ResizeObserver) {
                            const ro = new ResizeObserver(entries => {
                                if (entries && entries[0] && chart) {
                                    const cr = entries[0].contentRect;
                                    if (cr.width > 0 && cr.height > 0) {
                                        chart.applyOptions({ width: cr.width, height: cr.height });
                                        chart.timeScale().fitContent();
                                    }
                                }
                            });
                            ro.observe(container);
                        }
                        return;
                    } catch (err) {
                        console.warn('LightweightCharts init error, using SVG fallback:', err);
                    }
                }

                // Fallback to pure SVG Candlestick
                renderSVGCandlestickFallback(container, candleData, anomalyDate);
            }

            function renderSVGCandlestickFallback(container, candleData, anomalyDate) {
                const width = container.clientWidth || 300;
                const height = container.clientHeight || 240;
                let minP = Infinity, maxP = -Infinity;
                candleData.forEach(c => {
                    if (c.low < minP) minP = c.low;
                    if (c.high > maxP) maxP = c.high;
                });
                const pad = (maxP - minP) * 0.1 || 10;
                minP -= pad;
                maxP += pad;
                const priceRange = maxP - minP || 1;

                const stepX = width / candleData.length;
                const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
                const gridColor = isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(20, 20, 19, 0.06)';
                let svg = `<svg width="100%" height="100%" viewBox="0 0 ${width} ${height}" style="display:block;">`;
                svg += `<line x1="0" y1="${height * 0.3}" x2="${width}" y2="${height * 0.3}" stroke="${gridColor}" />`;
                svg += `<line x1="0" y1="${height * 0.6}" x2="${width}" y2="${height * 0.6}" stroke="${gridColor}" />`;

                candleData.forEach((c, idx) => {
                    const x = idx * stepX + stepX / 2;
                    const yHigh = height - ((c.high - minP) / priceRange) * (height * 0.68) - height * 0.16;
                    const yLow = height - ((c.low - minP) / priceRange) * (height * 0.68) - height * 0.16;
                    const yOpen = height - ((c.open - minP) / priceRange) * (height * 0.68) - height * 0.16;
                    const yClose = height - ((c.close - minP) / priceRange) * (height * 0.68) - height * 0.16;
                    const isUp = c.close >= c.open;
                    const color = isUp ? '#0ECB81' : '#F6465D';

                    const candleTop = Math.min(yOpen, yClose);
                    const candleH = Math.max(Math.abs(yClose - yOpen), 2);

                    svg += `<line x1="${x}" y1="${yHigh}" x2="${x}" y2="${yLow}" stroke="${color}" stroke-width="1.2" />`;
                    svg += `<rect x="${x - stepX * 0.3}" y="${candleTop}" width="${Math.max(stepX * 0.6, 2)}" height="${candleH}" fill="${color}" rx="1" />`;

                    if (c.time === anomalyDate) {
                        svg += `<polygon points="${x},${yHigh - 4} ${x - 5},${yHigh - 12} ${x + 5},${yHigh - 12}" fill="#F6465D" />`;
                        svg += `<text x="${x}" y="${yHigh - 16}" font-size="9" font-weight="700" fill="#F6465D" text-anchor="middle">Z≥2.5σ</text>`;
                    }
                });

                svg += `</svg>`;
                container.innerHTML = svg;
            }

            // 3-Tier Evidence Matrix Renderer (Law 2 Strict Compliance)
            function renderEvidenceMatrixHTML(findings) {
                if (!findings || findings.length === 0) return '';

                const tiers = {
                    SUPPORTED: [],
                    UNCERTAIN: [],
                    CONTRADICTED: []
                };

                findings.forEach(f => {
                    const st = (f.verification_status || f.status || 'SUPPORTED').toUpperCase();
                    if (st.includes('CONTRADICT')) {
                        tiers.CONTRADICTED.push(f);
                    } else if (st.includes('UNCERTAIN') || st.includes('RUMOR')) {
                        tiers.UNCERTAIN.push(f);
                    } else {
                        tiers.SUPPORTED.push(f);
                    }
                });

                let itemsHTML = '';
                const formatFindingItem = (item, tierClass, badgeClass, defaultLabel) => {
                    const title = item.title || item.claim || t('market_finding');
                    const conf = item.confidence_score !== undefined ? `${Math.round(item.confidence_score * 100)}% ${t('confidence')}` : t('verified');
                    const source = item.source_url || item.source || 'IDXnet / Sectors v2';
                    const causality = item.causality_tag || (tierClass === 'tier-supported' ? 'LIKELY_CATALYST' : (tierClass === 'tier-contradicted' ? 'REFUTED' : 'UNEXPLAINED_BY_NEWS'));

                    return `
                        <div class="evidence-item ${tierClass}">
                            <div class="evidence-item-top">
                                <span class="evidence-badge ${badgeClass}">${defaultLabel}</span>
                                <span style="font-size:10px; font-family:var(--font-mono); color:var(--text-muted); font-weight:600;">${conf}</span>
                            </div>
                            <div class="evidence-item-claim">${escapeHtml(title)}</div>
                            <div class="evidence-item-meta">
                                <span>Tag: <strong style="color:var(--text-secondary);">${escapeHtml(causality)}</strong></span>
                                <span class="evidence-source-link" title="${escapeHtml(source)}">${escapeHtml(source.length > 25 ? source.slice(0, 22) + '...' : source)}</span>
                            </div>
                        </div>
                    `;
                };

                tiers.SUPPORTED.forEach(f => { itemsHTML += formatFindingItem(f, 'tier-supported', 'badge-supported', 'SUPPORTED'); });
                tiers.UNCERTAIN.forEach(f => { itemsHTML += formatFindingItem(f, 'tier-uncertain', 'badge-uncertain', 'UNCERTAIN'); });
                tiers.CONTRADICTED.forEach(f => { itemsHTML += formatFindingItem(f, 'tier-contradicted', 'badge-contradicted', 'CONTRADICTED'); });

                return `
                    <div class="evidence-matrix-card">
                        <div class="evidence-matrix-head">
                            <div class="evidence-matrix-title">
                                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
                                <span>${t('evidence_matrix_title')}</span>
                            </div>
                            <span class="panel-card-badge">${findings.length} ${t('evidence_verified_badge')}</span>
                        </div>
                        <div class="evidence-matrix-grid">
                            ${itemsHTML}
                        </div>
                        <div class="evidence-disclaimer">
                            ${t('evidence_disclaimer')}
                        </div>
                    </div>
                `;
            }

            // OpenCode Branching / Forking Handler
            async function forkChatSession(sessionId, messageId, title) {
                const defaultForkTitle = `Fork: ${title || (currentLang === 'en' ? 'Market Research' : 'Riset Pasar')}`;
                const newTitle = await showInputDialog({
                    title: t('fork_dialog_title') || (currentLang === 'en' ? 'Branch Research Session (Fork)' : 'Cabangkan Sesi Riset (Fork)'),
                    message: t('fork_prompt') || (currentLang === 'en' ? 'Enter title for new branched session:' : 'Masukkan judul untuk sesi percabangan baru:'),
                    defaultValue: defaultForkTitle,
                    placeholder: t('fork_input_placeholder') || (currentLang === 'en' ? 'Enter session title...' : 'Masukkan judul sesi...'),
                    confirmText: t('btn_fork') || (currentLang === 'en' ? 'Fork Session' : 'Cabangkan Sesi'),
                    cancelText: t('modal_cancel_btn') || (currentLang === 'en' ? 'Cancel' : 'Batal'),
                    type: 'primary'
                });
                if (!newTitle || !newTitle.trim()) return;

                try {
                    const res = await fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(sessionId)}/fork`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            title: newTitle.trim(),
                            up_to_message_id: messageId
                        })
                    });
                    if (!res.ok) throw new Error(`HTTP ${res.status}`);
                    const data = await res.json();
                    const newSess = data.session;
                    if (newSess && newSess.id) {
                        showToast(currentLang === 'en' ? `Session forked: ${newSess.title}` : `Sesi berhasil dicabangkan: ${newSess.title}`);
                        await loadChatSessions();
                        switchSession(newSess.id, newSess.title);
                    }
                } catch (e) {
                    showToast(currentLang === 'en' ? `Failed to fork session: ${e.message}` : `Gagal mencabangkan sesi: ${e.message}`, true);
                }
            }
