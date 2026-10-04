// 10. Navigation, Dedicated Memory Graph & Settings/Toolkit Integration
            function initNavigationAndSettings() {
                const composerContainer = document.getElementById('composerContainer') || document.querySelector('.composer-container');
                const globalComplianceBox = document.getElementById('globalComplianceBox');
                const graphPageView = document.getElementById('graphPageView');
                const workspaceContent = document.getElementById('workspaceContent');
                const settingsModal = document.getElementById('settingsModal');

                // Graph Page Elements
                const statTotalNodes = document.getElementById('statTotalNodes');
                const statTotalEdges = document.getElementById('statTotalEdges');
                const statTopHub = document.getElementById('statTopHub');
                const statActiveSession = document.getElementById('statActiveSession');
                const graphFilterTicker = document.getElementById('graphFilterTicker');
                const graphFilterDepth = document.getElementById('graphFilterDepth');
                const graphFilterNodeType = document.getElementById('graphFilterNodeType');
                const btnApplyGraphFilter = document.getElementById('btnApplyGraphFilter');
                const btnResetGraphFilter = document.getElementById('btnResetGraphFilter');
                const graphFrame = document.getElementById('graphFrame');
                const btnOpenGraphExternal = document.getElementById('btnOpenGraphExternal');
                const btnReloadGraph = document.getElementById('btnReloadGraph');
                const btnBackToChat = document.getElementById('btnBackToChat');

                // Settings Modal Elements
                const btnCloseSettingsModal = document.getElementById('btnCloseSettingsModal');
                const btnCancelSettings = document.getElementById('btnCancelSettings');
                const btnSaveSettings = document.getElementById('btnSaveSettings');
                const modalTabs = document.querySelectorAll('.modal-tab-btn');
                const btnSettings = document.getElementById('btnSettings');

                // Settings Inputs
                const inputSectorsKey = document.getElementById('inputSectorsKey');
                const statusSectorsKey = document.getElementById('statusSectorsKey');
                const btnTestSectors = document.getElementById('btnTestSectors');
                const selectAiProvider = document.getElementById('selectAiProvider');
                const inputGeminiKey = document.getElementById('inputGeminiKey');
                const statusGeminiKey = document.getElementById('statusGeminiKey');
                const btnTestGemini = document.getElementById('btnTestGemini');
                const inputGeminiModel = document.getElementById('inputGeminiModel');
                const inputOpenaiKey = document.getElementById('inputOpenaiKey');
                const statusOpenaiKey = document.getElementById('statusOpenaiKey');
                const btnTestOpenai = document.getElementById('btnTestOpenai');
                const inputOpenaiBaseUrl = document.getElementById('inputOpenaiBaseUrl');
                const inputOpenaiModel = document.getElementById('inputOpenaiModel');
                const inputOllamaBaseUrl = document.getElementById('inputOllamaBaseUrl');
                const inputOllamaModel = document.getElementById('inputOllamaModel');
                const statusOllamaKey = document.getElementById('statusOllamaKey');
                const btnTestOllama = document.getElementById('btnTestOllama');
                const inputAnthropicKey = document.getElementById('inputAnthropicKey');
                const inputAnthropicModel = document.getElementById('inputAnthropicModel');
                const statusAnthropicKey = document.getElementById('statusAnthropicKey');
                const btnTestAnthropic = document.getElementById('btnTestAnthropic');
                const btnHeaderSettings = document.getElementById('btnHeaderSettings');

                function updateProviderVisibility() {
                    const prov = selectAiProvider ? selectAiProvider.value : 'gemini';
                    document.querySelectorAll('.provider-fields-group').forEach(group => {
                        const target = group.getAttribute('data-provider');
                        if (target === prov) {
                            group.style.display = 'block';
                        } else {
                            group.style.display = 'none';
                        }
                    });

                    // Sync visual provider cards
                    document.querySelectorAll('.provider-card-tile').forEach(card => {
                        const target = card.getAttribute('data-provider');
                        card.classList.toggle('selected', target === prov);
                    });

                    // Preset auto-fill helper only if field is blank (preserves custom model)
                    if (prov === 'gemini') {
                        if (inputGeminiModel && !inputGeminiModel.value.trim()) {
                            inputGeminiModel.value = 'gemini-2.0-flash';
                        }
                    } else if (prov === 'openai') {
                        if (inputOpenaiBaseUrl && !inputOpenaiBaseUrl.value.trim()) {
                            inputOpenaiBaseUrl.value = 'http://localhost:20128/v1';
                        }
                        if (inputOpenaiModel && !inputOpenaiModel.value.trim()) {
                            inputOpenaiModel.value = 'gpt-4o-mini';
                        }
                    } else if (prov === 'ollama') {
                        if (inputOllamaBaseUrl && !inputOllamaBaseUrl.value.trim()) {
                            inputOllamaBaseUrl.value = 'http://localhost:11434';
                        }
                        if (inputOllamaModel && !inputOllamaModel.value.trim()) {
                            inputOllamaModel.value = 'deepseek-r1:8b';
                        }
                    } else if (prov === 'anthropic') {
                        if (inputAnthropicModel && !inputAnthropicModel.value.trim()) {
                            inputAnthropicModel.value = 'claude-3-5-sonnet-20241022';
                        }
                    }
                    if (typeof updateActivePresetChips === 'function') {
                        updateActivePresetChips();
                    }
                }
                if (selectAiProvider) {
                    selectAiProvider.addEventListener('change', updateProviderVisibility);
                }

                // Interactive Provider Cards Click
                document.querySelectorAll('.provider-card-tile').forEach(card => {
                    card.addEventListener('click', (e) => {
                        e.preventDefault();
                        const p = card.getAttribute('data-provider');
                        if (selectAiProvider && p) {
                            selectAiProvider.value = p;
                            updateProviderVisibility();
                        }
                    });
                });

                // Password Mask Toggle (.btn-toggle-mask)
                document.querySelectorAll('.btn-toggle-mask').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        e.preventDefault();
                        const targetId = btn.getAttribute('data-target');
                        const input = document.getElementById(targetId);
                        if (!input) return;
                        const isPwd = (input.type === 'password');
                        input.type = isPwd ? 'text' : 'password';
                        btn.innerHTML = isPwd ?
                            '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line></svg>' :
                            '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>';
                        btn.title = isPwd ? 'Sembunyikan Kunci' : 'Tampilkan Kunci';
                    });
                });

                // Quick Preset Chips for Models & Base URLs
                function updateActivePresetChips() {
                    if (inputGeminiModel) {
                        const val = inputGeminiModel.value.trim();
                        document.querySelectorAll('.preset-chip[data-fill="inputGeminiModel"]').forEach(chip => {
                            chip.classList.toggle('active', chip.getAttribute('data-val') === val);
                        });
                    }
                    if (inputOpenaiModel) {
                        const val = inputOpenaiModel.value.trim();
                        document.querySelectorAll('.preset-chip[data-fill="inputOpenaiModel"]').forEach(chip => {
                            chip.classList.toggle('active', chip.getAttribute('data-val') === val);
                        });
                    }
                    if (inputOllamaModel) {
                        const val = inputOllamaModel.value.trim();
                        document.querySelectorAll('.preset-chip[data-fill="inputOllamaModel"]').forEach(chip => {
                            chip.classList.toggle('active', chip.getAttribute('data-val') === val);
                        });
                    }
                    if (inputAnthropicModel) {
                        const val = inputAnthropicModel.value.trim();
                        document.querySelectorAll('.preset-chip[data-fill="inputAnthropicModel"]').forEach(chip => {
                            chip.classList.toggle('active', chip.getAttribute('data-val') === val);
                        });
                    }
                    const slider = document.getElementById('timeout-slider');
                    if (slider) {
                        const tVal = String(Math.round(parseFloat(slider.value) || 60));
                        document.querySelectorAll('.preset-chip[data-timeout]').forEach(chip => {
                            chip.classList.toggle('active', chip.getAttribute('data-timeout') === tVal);
                        });
                    }
                }
                window.updateActivePresetChips = updateActivePresetChips;

                document.querySelectorAll('.preset-chip[data-fill]').forEach(chip => {
                    chip.addEventListener('click', (e) => {
                        e.preventDefault();
                        const targetId = chip.getAttribute('data-fill');
                        const val = chip.getAttribute('data-val');
                        const input = document.getElementById(targetId);
                        if (input && val) {
                            input.value = val;
                            input.focus();
                            updateActivePresetChips();
                            showToast(currentLang === 'en' ? `Preset applied: ${val}` : `Preset diterapkan: ${val}`);
                        }
                    });
                });

                // Quick Preset Chips for Timeout
                document.querySelectorAll('.preset-chip[data-timeout]').forEach(chip => {
                    chip.addEventListener('click', (e) => {
                        e.preventDefault();
                        const val = chip.getAttribute('data-timeout');
                        if (val) {
                            populateTimeoutSlider(val);
                            updateActivePresetChips();
                        }
                    });
                });

                // Focus & input behavior for saved API keys
                ['inputSectorsKey', 'inputGeminiKey', 'inputOpenaiKey', 'inputAnthropicKey', 'inputTeleToken'].forEach(id => {
                    const el = document.getElementById(id);
                    if (!el) return;
                    el.addEventListener('focus', () => {
                        if (el.dataset.saved === 'true' && el.value) {
                            el.select();
                        }
                    });
                });

                ['inputGeminiModel', 'inputOpenaiModel', 'inputOllamaModel', 'inputAnthropicModel'].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) {
                        el.addEventListener('input', updateActivePresetChips);
                    }
                });

                // Preferences Elements
                const timeoutSlider = document.getElementById('timeout-slider');

                // Telegram Elements
                const teleBotStatusBadge = document.getElementById('teleBotStatusBadge');
                const teleBotUsername = document.getElementById('teleBotUsername');
                const btnStartTeleBot = document.getElementById('btnStartTeleBot');
                const btnStopTeleBot = document.getElementById('btnStopTeleBot');
                const inputTeleToken = document.getElementById('inputTeleToken');
                const statusTeleToken = document.getElementById('statusTeleToken');
                const btnUnlockTeleToken = document.getElementById('btnUnlockTeleToken');
                const inputTeleUsers = document.getElementById('inputTeleUsers');
                const inputTeleTestChatId = document.getElementById('inputTeleTestChatId');
                const btnSendTeleTest = document.getElementById('btnSendTeleTest');

                if (btnUnlockTeleToken) {
                    btnUnlockTeleToken.addEventListener('click', () => {
                        if (inputTeleToken) {
                            inputTeleToken.disabled = false;
                            inputTeleToken.value = '';
                            inputTeleToken.placeholder = '7123456789:AAH... (Masukkan token baru)';
                            inputTeleToken.focus();
                        }
                        btnUnlockTeleToken.style.display = 'none';
                    });
                }

                // Cache & Diagnostics Elements
                const statCacheTotal = document.getElementById('statCacheTotal');
                const statCreditSaved = document.getElementById('statCreditSaved');
                const statCachePermanent = document.getElementById('statCachePermanent');
                const statCacheExpired = document.getElementById('statCacheExpired');
                const btnCleanCache = document.getElementById('btnCleanCache');
                const diagnosticsDetails = document.getElementById('diagnosticsDetails');

                let lastSettingsData = null;
                let lastTelegramData = null;
                let lastDiagnosticsData = null;

                switchMainView = function(targetNav) {
                    if (typeof closeMobileSidebar === 'function') {
                        closeMobileSidebar();
                    }

                    document.querySelectorAll('.nav-link-item').forEach(item => {
                        item.classList.toggle('active', item.getAttribute('data-nav') === targetNav);
                    });

                    const invPageView = document.getElementById('investigationsPageView');

                    if (targetNav === 'investigations') {
                        if (workspaceContent) workspaceContent.classList.remove('graph-mode');
                        if (heroView) {
                            heroView.classList.add('hidden');
                            heroView.style.display = 'none';
                        }
                        if (chatView) {
                            chatView.classList.remove('active');
                            chatView.style.display = 'none';
                        }
                        if (globalComplianceBox) globalComplianceBox.style.display = 'none';
                        if (composerContainer) composerContainer.style.display = 'none';
                        if (graphPageView) graphPageView.style.display = 'none';
                        if (invPageView) invPageView.style.display = 'flex';
                        loadInvestigations();
                    } else if (targetNav === 'graphify') {
                        if (workspaceContent) workspaceContent.classList.add('graph-mode');
                        if (invPageView) invPageView.style.display = 'none';
                        if (heroView) {
                            heroView.classList.add('hidden');
                            heroView.style.display = 'none';
                        }
                        if (chatView) {
                            chatView.classList.remove('active');
                            chatView.style.display = 'none';
                        }
                        if (globalComplianceBox) globalComplianceBox.style.display = 'none';
                        if (composerContainer) composerContainer.style.display = 'none';
                        if (graphPageView) graphPageView.style.display = 'flex';
                        loadGraphPageData();
                    } else if (targetNav === 'toolkit') {
                        if (workspaceContent) workspaceContent.classList.remove('graph-mode');
                        openSettingsModal();
                    } else {
                        // Default: chat view
                        if (workspaceContent) workspaceContent.classList.remove('graph-mode');
                        if (invPageView) invPageView.style.display = 'none';
                        if (graphPageView) graphPageView.style.display = 'none';
                        if (globalComplianceBox) globalComplianceBox.style.display = 'block';
                        if (composerContainer) composerContainer.style.display = 'flex';
                        const hasMessages = chatView && chatView.children.length > 0;
                        if (hasMessages) {
                            if (typeof showChatView === 'function') {
                                showChatView();
                            } else {
                                if (heroView) { heroView.classList.add('hidden'); heroView.style.display = 'none'; }
                                if (chatView) { chatView.classList.add('active'); chatView.style.display = 'flex'; }
                            }
                        } else {
                            if (typeof showHeroView === 'function') {
                                showHeroView();
                            } else {
                                if (heroView) { heroView.classList.remove('hidden'); heroView.style.display = 'flex'; }
                                if (chatView) { chatView.classList.remove('active'); chatView.style.display = 'none'; }
                            }
                        }
                    }
                }

                // Dedicated Memory Graph Page loader
                function loadGraphPageData() {
                    const externalUrl = `${API_BASE}/graph`;
                    if (btnOpenGraphExternal) btnOpenGraphExternal.href = externalUrl;

                    // Load the full authentic memory graph directly
                    if (graphFrame) {
                        const currentSrc = graphFrame.getAttribute('src');
                        if (!currentSrc || currentSrc === 'about:blank') {
                            graphFrame.src = externalUrl;
                        }
                    }
                }

                if (btnReloadGraph) {
                    btnReloadGraph.addEventListener('click', () => {
                        if (graphFrame) {
                            const sep = API_BASE.includes('?') ? '&' : '?';
                            graphFrame.src = `${API_BASE}/graph${sep}_t=${Date.now()}`;
                        }
                        showToast(currentLang === 'en' ? 'Memory Graph refreshed' : 'Graf Memori disegarkan');
                    });
                }

                if (btnApplyGraphFilter) {
                    btnApplyGraphFilter.addEventListener('click', () => {
                        loadGraphPageData();
                        showToast(currentLang === 'en' ? 'Memory Graph filter applied' : 'Filter Memory Graph diterapkan');
                    });
                }

                if (btnResetGraphFilter) {
                    btnResetGraphFilter.addEventListener('click', () => {
                        if (graphFilterTicker) graphFilterTicker.value = '';
                        if (graphFilterDepth) graphFilterDepth.value = '1';
                        if (graphFilterNodeType) graphFilterNodeType.value = '';
                        loadGraphPageData();
                        showToast(currentLang === 'en' ? 'Memory Graph filter reset' : 'Filter Memory Graph direset');
                    });
                }

                // Interactive Top Hub Node Click -> Filter
                if (statTopHub) {
                    statTopHub.style.cursor = 'pointer';
                    statTopHub.title = currentLang === 'en' ? 'Click to filter graph to this ticker' : 'Klik untuk memfilter graf ke emiten ini';
                    statTopHub.addEventListener('click', () => {
                        const rawText = statTopHub.textContent.trim();
                        const hubTicker = rawText.split(' ')[0].replace(/[^a-zA-Z0-9]/g, '');
                        if (hubTicker && hubTicker !== '-') {
                            if (graphFilterTicker) graphFilterTicker.value = hubTicker;
                            loadGraphPageData();
                            showToast(currentLang === 'en' ? `Memory Graph filter directed to ${hubTicker}` : `Filter Memory Graph diarahkan ke ${hubTicker}`);
                        }
                    });
                }

                if (btnBackToChat) {
                    btnBackToChat.addEventListener('click', () => {
                        switchMainView('chat');
                    });
                }

                // Market Radar & Forensic Dossier Handlers
                let selectedInvestigationId = null;
                const btnInvBackToChat = document.getElementById('btnInvBackToChat');
                if (btnInvBackToChat) {
                    btnInvBackToChat.addEventListener('click', () => switchMainView('chat'));
                }

                const btnInvOpenChat = document.getElementById('btnInvOpenChat');
                if (btnInvOpenChat) {
                    btnInvOpenChat.addEventListener('click', () => {
                        if (selectedInvestigationId) {
                            switchSession(selectedInvestigationId);
                        }
                        switchMainView('chat');
                    });
                }

                const btnInvExportMd = document.getElementById('btnInvExportMd');
                if (btnInvExportMd) {
                    btnInvExportMd.addEventListener('click', () => {
                        if (!selectedInvestigationId) {
                            showToast(currentLang === 'en' ? 'No active investigation selected' : 'Belum ada sesi investigasi yang dipilih', true);
                            return;
                        }
                        const titleEl = document.getElementById('invDossierTitle');
                        const ticker = titleEl ? titleEl.textContent.replace('DOSSIER:', '').replace('HASIL AUDIT:', '').trim() : 'IDX';
                        const zScore = document.getElementById('invDossierZScore')?.textContent || '-';
                        const actualVol = document.getElementById('invDossierActualVol')?.textContent || '-';
                        const baselineVol = document.getElementById('invDossierBaselineVol')?.textContent || '-';
                        const returnDev = document.getElementById('invDossierReturnDev')?.textContent || '-';
                        const dateStr = new Date().toISOString().slice(0, 10);

                        let md = `# LAPORAN AUDIT OTONOM INTELIJEN PASAR: ${ticker}\n\n`;
                        md += `**ID Sesi:** \`${selectedInvestigationId}\` | **Tanggal Ekspor:** ${dateStr}\n\n`;
                        md += `## 1. Indikator Kuantitatif Deterministik (Law 1)\n\n`;
                        md += `- **Volume Z-Score ($V_z$):** ${zScore}\n`;
                        md += `- **Volume Perdagangan Aktual:** ${actualVol}\n`;
                        md += `- **Baseline Historis MA20:** ${baselineVol}\n`;
                        md += `- **Deviasi Return ($R_t$):** ${returnDev}\n\n`;

                        md += `## 2. Matriks Verifikasi Bukti 3-Tier (Law 2)\n\n`;
                        const findingCards = document.querySelectorAll('#invEvidenceContainer .evidence-item');
                        if (findingCards.length > 0) {
                            findingCards.forEach(c => {
                                const claim = c.querySelector('.evidence-claim')?.textContent.trim() || '';
                                const badge = c.querySelector('.evidence-status-badge')?.textContent.trim() || 'UNCERTAIN';
                                const source = c.querySelector('.evidence-source')?.textContent.trim() || '-';
                                md += `### [${badge}] ${claim}\n`;
                                md += `- **Sumber Rujukan:** ${source}\n\n`;
                            });
                        } else {
                            md += `*Data temuan matriks bukti diverifikasi langsung melalui dokumen keterbukaan informasi IDXnet.*\n\n`;
                        }

                        md += `## 3. Timeline Rekonsiliasi Kronologis Kausalitas\n\n`;
                        const signals = document.querySelectorAll('#invTimelineContainer .signal-item');
                        if (signals.length > 0) {
                            signals.forEach(sig => {
                                const time = sig.querySelector('.signal-time')?.textContent.trim() || '-';
                                const title = sig.querySelector('.signal-title')?.textContent.trim() || '';
                                const desc = sig.querySelector('.signal-desc')?.textContent.trim() || '';
                                md += `- **${time}** - **${title}**: ${desc}\n`;
                            });
                            md += `\n`;
                        }

                        md += `> [!IMPORTANT]\n`;
                        md += `> **Kepatuhan Regulasi Pasar Modal (Law 2 & Law 3):**\n`;
                        md += `> Niskava Agent adalah platform riset intelijen pasar modal berbasis bukti untuk Bursa Efek Indonesia (IDX), BUKAN penasihat investasi berizin dan BUKAN broker perdagangan saham. Seluruh dossier, skor anomali statistik, dan matriks bukti disajikan secara deskriptif untuk tujuan verifikasi fakta dan transparansi pasar, serta BUKAN merupakan rekomendasi beli/jual saham atau saran finansial terpersonalisasi.\n`;

                        const blob = new Blob([md], { type: 'text/markdown;charset=utf-8' });
                        const url = window.URL.createObjectURL(blob);
                        const a = document.createElement('a');
                        a.href = url;
                        a.download = `niskava-audit-${ticker}-${selectedInvestigationId.slice(0, 8)}.md`;
                        document.body.appendChild(a);
                        a.click();
                        document.body.removeChild(a);
                        window.URL.revokeObjectURL(url);
                        showToast(currentLang === 'en' ? 'Audit report markdown downloaded' : 'Laporan audit markdown berhasil diekspor');
                    });
                }

                const btnRadarRefresh = document.getElementById('btnRadarRefresh');
                if (btnRadarRefresh) {
                    btnRadarRefresh.addEventListener('click', () => {
                        loadRadarAnomalies();
                        showToast(currentLang === 'en' ? 'Refreshing anomaly signals...' : 'Menyegarkan data anomali...');
                    });
                }

                function formatRadarMetricName(type) {
                    if (!type) return 'Volume Spike';
                    if (type === 'volume_zscore') return 'Volume Anomaly (Vz)';
                    if (type === 'price_zscore') return 'Price Volatility';
                    if (type === 'foreign_flow_zscore') return 'Foreign Flow Divergence';
                    return type.replace(/_/g, ' ').toUpperCase();
                }

                function formatRadarVolumeShort(num) {
                    const val = Number(num);
                    if (isNaN(val)) return '-';
                    if (Math.abs(val) >= 1e9) return (val / 1e9).toFixed(1) + 'B';
                    if (Math.abs(val) >= 1e6) return (val / 1e6).toFixed(1) + 'M';
                    if (Math.abs(val) >= 1e3) return (val / 1e3).toFixed(1) + 'k';
                    return val.toLocaleString('id-ID');
                }

                function renderRadarEmptyState(container) {
                    if (!container) return;
                    const msg = (typeof t === 'function' && t('radar_empty')) || (currentLang === 'en'
                        ? 'No anomalies above 2.0σ threshold yet. Enter an IDX ticker below to inspect.'
                        : 'Belum ada anomali terdeteksi melebihi ambang batas 2.0σ. Masukkan kode emiten di bawah untuk memeriksa.');
                    container.innerHTML = `
                        <div class="radar-empty-state">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="opacity:0.6;"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                            <span>${msg}</span>
                        </div>
                    `;
                }

                async function loadRadarAnomalies() {
                    const listEl = document.getElementById('radarAnomaliesList');
                    if (!listEl) return;

                    try {
                        const res = await fetch(`${API_BASE}/api/radar/anomalies?limit=12&min_z=2.0`);
                        if (!res.ok) {
                            renderRadarEmptyState(listEl);
                            return;
                        }
                        const data = await res.json();
                        const rawAnomalies = (data && Array.isArray(data.anomalies)) ? data.anomalies : [];

                        // Deduplicate tickers so each ticker appears at most once in the screener
                        const seenTickers = new Set();
                        const anomalies = [];
                        for (const item of rawAnomalies) {
                            const t = (item.ticker || '').trim().toUpperCase();
                            if (t && !seenTickers.has(t)) {
                                seenTickers.add(t);
                                anomalies.push(item);
                            }
                        }

                        if (anomalies.length === 0) {
                            renderRadarEmptyState(listEl);
                            return;
                        }

                        listEl.innerHTML = '';
                        anomalies.forEach((item) => {
                            const z = Number(item.z_score || 0);
                            const ticker = escapeHtml(item.ticker || 'IDX');
                            const invId = item.investigation_id || `INV-${ticker}`;
                            const dateStr = item.anomaly_date || (item.created_at ? item.created_at.slice(0, 10) : '');

                            let zBadgeClass = 'radar-zscore-normal';
                            if (Math.abs(z) >= 3.0) {
                                zBadgeClass = 'radar-zscore-critical';
                            } else if (Math.abs(z) >= 2.5) {
                                zBadgeClass = 'radar-zscore-warning';
                            }

                            const metricLabel = item.metric_type ? formatRadarMetricName(item.metric_type) : 'Volume Spike';
                            let metricDetail = '';
                            if (item.metric_value && item.baseline_value) {
                                metricDetail = `Aktual ${formatRadarVolumeShort(item.metric_value)} / MA20 ${formatRadarVolumeShort(item.baseline_value)}`;
                            } else if (item.description) {
                                metricDetail = escapeHtml(item.description);
                            }

                            const card = document.createElement('div');
                            card.className = 'radar-card';
                            card.setAttribute('role', 'button');
                            card.setAttribute('tabindex', '0');
                            card.title = `${ticker}: Z-Score ${z >= 0 ? '+' : ''}${z.toFixed(2)}σ (${metricLabel})`;
                            const inspectLabel = (typeof t === 'function' && t('radar_btn_inspect')) || (currentLang === 'en' ? 'Audit Stock' : 'Audit Saham Ini');

                            card.innerHTML = `
                                <div class="radar-card-top">
                                    <span class="radar-card-ticker">${ticker}</span>
                                    <span class="radar-card-zscore ${zBadgeClass}">${z >= 0 ? '+' : ''}${z.toFixed(2)}σ</span>
                                </div>
                                <div class="radar-card-metric">
                                    <span style="color:var(--text-muted); font-size:10px;">${escapeHtml(metricLabel)}</span>
                                    <span class="radar-card-metric-val">${metricDetail || '-'}</span>
                                </div>
                                <div class="radar-card-date">${dateStr ? escapeHtml(dateStr) : '-'}</div>
                                <button class="radar-card-btn" type="button">
                                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                                    <span>${inspectLabel}</span>
                                </button>
                            `;

                            const onInspect = (e) => {
                                if (e) e.stopPropagation();
                                inspectRadarCase(item);
                            };

                            card.addEventListener('click', onInspect);
                            card.addEventListener('keydown', (e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                    e.preventDefault();
                                    onInspect(e);
                                }
                            });

                            const inspectBtn = card.querySelector('.radar-card-btn');
                            if (inspectBtn) {
                                inspectBtn.addEventListener('click', onInspect);
                            }

                            listEl.appendChild(card);
                        });

                        // Auto-select first anomaly if none selected yet
                        if (!selectedInvestigationId && anomalies.length > 0) {
                            const first = anomalies[0];
                            selectInvestigation(first.investigation_id || `INV-${first.ticker}`, first);
                        }
                    } catch (e) {
                        console.warn('Gagal memuat anomali pasar:', e);
                        renderRadarEmptyState(listEl);
                    }
                }

                function inspectRadarCase(arg1, arg2) {
                    let ticker = '';
                    let invId = '';
                    let meta = {};
                    if (typeof arg1 === 'object' && arg1 !== null) {
                        meta = arg1;
                        ticker = (arg1.ticker || '').trim().toUpperCase();
                        invId = arg1.investigation_id || ('INV-' + ticker);
                    } else {
                        ticker = (arg1 || '').trim().toUpperCase();
                        invId = arg2 || ('INV-' + ticker);
                        meta = { ticker: ticker };
                    }
                    if (!ticker) return;

                    const inputInvTicker = document.getElementById('inputInvTicker');
                    if (inputInvTicker) {
                        inputInvTicker.value = ticker;
                    }
                    selectInvestigation(invId, meta);
                    const dossierCard = document.getElementById('invDossierCard');
                    if (dossierCard) {
                        dossierCard.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
                    }
                    showToast(currentLang === 'en' ? `Inspecting audit for ${ticker}` : `Membuka audit saham ${ticker}`);
                }

                const btnRunFormalInvestigation = document.getElementById('btnRunFormalInvestigation');
                const inputInvTicker = document.getElementById('inputInvTicker');
                const selectInvDays = document.getElementById('selectInvDays');

                if (btnRunFormalInvestigation) {
                    btnRunFormalInvestigation.addEventListener('click', async () => {
                        const ticker = (inputInvTicker ? inputInvTicker.value : '').trim().toUpperCase();
                        if (!ticker) {
                            showToast(currentLang === 'en' ? 'Please enter IDX ticker symbol (e.g., ANTM)' : 'Masukkan kode ticker emiten BEI (misal: ANTM)', true);
                            if (inputInvTicker) inputInvTicker.focus();
                            return;
                        }
                        const days = selectInvDays ? selectInvDays.value : '30';

                        // Switch to chat view and initiate formal audit stream
                        switchMainView('chat');
                        if (chatInput) {
                            chatInput.value = `Jalankan investigasi formal 7-tahap otonom pada emiten ${ticker} untuk observasi ${days} hari terakhir.`;
                            handleSendMessage();
                        }
                        showToast(currentLang === 'en' ? `Starting audit for ${ticker}...` : `Memulai audit emiten ${ticker}...`);
                    });
                }

                async function loadInvestigations() {
                    // Load anomaly screener strip and auto-populate active case
                    loadRadarAnomalies();
                }

                function formatVolumeCompact(num) {
                    const n = Number(num) || 0;
                    if (n >= 1e9) return (n / 1e9).toFixed(1) + 'B';
                    if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M';
                    if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K';
                    return n.toLocaleString('id-ID');
                }

                function renderDossierEvidenceEmpty(ticker) {
                    const isEn = currentLang === 'en';
                    return `
                        <div class="dossier-empty-state">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="opacity:0.6;"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                            <strong>${isEn ? `No Verified Evidence Dossier for ${escapeHtml(ticker)}` : `Belum Ada Berkas Bukti Terverifikasi untuk ${escapeHtml(ticker)}`}</strong>
                            <span>${isEn ? 'This stock has not completed a 7-stage causality audit. Start an investigation to correlate IDXnet disclosures and official news.' : 'Emiten ini belum melewati audit kausalitas 7-tahap. Mulai investigasi untuk menyaring keterbukaan informasi IDXnet dan berita bursa.'}</span>
                            <button class="btn-primary" style="margin-top:6px;" onclick="initiateAuditFromDossier('${escapeHtml(ticker)}')">
                                ${isEn ? `Start Causality Audit for ${escapeHtml(ticker)}` : `Mulai Audit Kausalitas ${escapeHtml(ticker)}`}
                            </button>
                        </div>
                    `;
                }

                window.initiateAuditFromDossier = function(ticker) {
                    switchMainView('chat');
                    const chatInput = document.getElementById('chatInput');
                    if (chatInput) {
                        chatInput.value = `Jalankan investigasi formal 7-tahap otonom pada emiten ${ticker} untuk observasi 30 hari terakhir.`;
                        if (typeof handleSendMessage === 'function') {
                            handleSendMessage();
                        }
                    }
                    showToast(currentLang === 'en' ? `Starting audit for ${ticker}...` : `Memulai audit emiten ${ticker}...`);
                };

                window.openDossierFromChat = function(ticker) {
                    if (!ticker) return;
                    const cleanTicker = ticker.trim().toUpperCase();

                    // 1. Switch view to investigations dossier
                    if (typeof switchMainView === 'function') {
                        switchMainView('investigations');
                    }

                    // 2. Select ticker in dossier and load real candles & evidence
                    if (typeof inspectRadarCase === 'function') {
                        inspectRadarCase({
                            ticker: cleanTicker,
                            investigation_id: 'INV-' + cleanTicker
                        });
                    }

                    // 3. Smoothly scroll dossier into view
                    const dossierCard = document.getElementById('invDossierCard');
                    if (dossierCard) {
                        dossierCard.scrollIntoView({ behavior: 'smooth', block: 'start' });
                    }
                    showToast(typeof currentLang !== 'undefined' && currentLang === 'en' ? `Opened ${cleanTicker} dossier` : `Membuka berkas dossier ${cleanTicker}`);
                };

                async function selectInvestigation(invId, invMeta) {
                    selectedInvestigationId = invId;
                    const emptyEl = document.getElementById('invDossierEmpty');
                    const contentEl = document.getElementById('invDossierContent');
                    const titleEl = document.getElementById('invDossierTitle');
                    const statusEl = document.getElementById('invDossierStatus');
                    const metaEl = document.getElementById('invDossierMeta');

                    if (emptyEl) emptyEl.style.display = 'none';
                    if (contentEl) contentEl.style.display = 'flex';

                    const ticker = (invMeta && invMeta.ticker) || (invMeta && invMeta.title) || (invId ? invId.replace(/^INV-/, '') : 'IDX');
                    if (titleEl) titleEl.textContent = `HASIL AUDIT: ${ticker}`;
                    if (statusEl) {
                        const st = (invMeta && invMeta.status) || 'COMPLETED';
                        statusEl.textContent = st;
                        statusEl.className = `inv-status-pill inv-status-${st.toLowerCase() === 'completed' ? 'completed' : 'running'}`;
                    }
                    if (metaEl) {
                        metaEl.textContent = `${currentLang === 'en' ? 'Observation Horizon: 30 Trading Days | Sources: Sectors Financial API v2 & IDXnet' : 'Horizon: 30 Hari Perdagangan | Sumber: Sectors Financial API v2 & IDXnet'}`;
                    }

                    // 1. Fetch Anomalies (Law 1)
                    let anom = null;
                    try {
                        const anomRes = await fetch(`${API_BASE}/api/investigations/${invId}/anomalies`);
                        if (anomRes.ok) {
                            const anomData = await anomRes.json();
                            const anomalies = Array.isArray(anomData) ? anomData : (anomData.anomalies || []);
                            if (anomalies.length > 0) anom = anomalies[0];
                        }
                    } catch (e) {
                        console.warn('Gagal memuat anomali investigasi:', e);
                    }

                    const zVal = (anom && anom.z_score != null) ? Number(anom.z_score) : (invMeta && invMeta.z_score != null ? Number(invMeta.z_score) : null);
                    const actualVal = (anom && anom.actual_value != null) ? anom.actual_value : (invMeta && invMeta.actual_value != null ? invMeta.actual_value : null);
                    const baseVal = (anom && anom.baseline_mean != null) ? anom.baseline_mean : (invMeta && invMeta.baseline_value != null ? invMeta.baseline_value : null);
                    const devVal = (anom && anom.metric_type && anom.metric_type !== 'volume_z_score') ? anom.metric_type : ((invMeta && invMeta.metric_type) || '-');

                    const zScoreEl = document.getElementById('invDossierZScore');
                    const actualVolEl = document.getElementById('invDossierActualVol');
                    const baselineVolEl = document.getElementById('invDossierBaselineVol');
                    const returnDevEl = document.getElementById('invDossierReturnDev');

                    if (zScoreEl) {
                        if (zVal != null && !isNaN(zVal)) {
                            zScoreEl.textContent = `${zVal >= 0 ? '+' : ''}${zVal.toFixed(2)}σ`;
                            zScoreEl.style.color = Math.abs(zVal) >= 3.0 ? '#AD1C1C' : (Math.abs(zVal) >= 2.5 ? '#8C5600' : '#0F6735');
                        } else {
                            zScoreEl.textContent = '-';
                            zScoreEl.style.color = 'var(--text-muted)';
                        }
                    }
                    if (actualVolEl) actualVolEl.textContent = actualVal != null ? formatVolumeCompact(actualVal) : '-';
                    if (baselineVolEl) baselineVolEl.textContent = baseVal != null ? formatVolumeCompact(baseVal) : '-';
                    if (returnDevEl) returnDevEl.textContent = devVal;

                    // 2. Render Candlestick Chart (TradingView with real market candles & provenance badge)
                    try {
                        const chartContainer = document.getElementById('invChartContainer');
                        if (chartContainer) {
                            const candleRes = await fetch(`${API_BASE}/api/market/candles?ticker=${encodeURIComponent(ticker)}&days=30`);
                            let candles = [];
                            const badgeEl = document.getElementById('invChartSourceBadge');
                            if (candleRes.ok) {
                                const candleJson = await candleRes.json();
                                if (badgeEl) {
                                    if (candleJson.source === 'sectors_api') {
                                        badgeEl.className = 'source-badge source-badge-live';
                                        badgeEl.textContent = '● LIVE: Sectors Financial API v2';
                                    } else if (candleJson.source === 'cache') {
                                        badgeEl.className = 'source-badge source-badge-live';
                                        badgeEl.textContent = '● CACHE: Sectors API v2 (Permanent)';
                                    } else {
                                        badgeEl.className = 'source-badge source-badge-synthetic';
                                        badgeEl.textContent = '▲ SIMULASI: Offline Synthetic Generator';
                                    }
                                }
                                if (candleJson && Array.isArray(candleJson.data) && candleJson.data.length > 0) {
                                    candles = candleJson.data.map(c => ({
                                        time: c.date,
                                        open: Number(c.open),
                                        high: Number(c.high),
                                        low: Number(c.low),
                                        close: Number(c.close),
                                        volume: Number(c.volume)
                                    }));
                                }
                            } else if (badgeEl) {
                                badgeEl.className = 'source-badge source-badge-error';
                                badgeEl.textContent = '✕ Error Data Pasar';
                            }
                            const anomalyDate = (anom && anom.anomaly_date) || (invMeta && invMeta.anomaly_date) || (candles.length > 5 ? candles[candles.length - 4].time : undefined);
                            renderTradingViewCandlestick('invChartContainer', candles, anomalyDate);
                        }
                    } catch (e) {
                        console.warn('Gagal render chart candlestick investigasi:', e);
                    }

                    // 3. Fetch Findings (Law 2 3-Tier Evidence Matrix)
                    try {
                        const findRes = await fetch(`${API_BASE}/api/investigations/${invId}/findings`);
                        const evidenceContainer = document.getElementById('invEvidenceContainer');
                        if (evidenceContainer) {
                            let findings = [];
                            if (findRes.ok) {
                                const findJson = await findRes.json();
                                findings = Array.isArray(findJson) ? findJson : (findJson.findings || []);
                            }
                            if (findings.length > 0) {
                                evidenceContainer.innerHTML = renderEvidenceMatrixHTML(findings);
                            } else {
                                evidenceContainer.innerHTML = renderDossierEvidenceEmpty(ticker);
                            }
                        }
                    } catch (e) {
                        console.warn('Gagal memuat findings:', e);
                    }

                    // 4. Render Evidence & News Timeline (Dynamic from API)
                    const timelineContainer = document.getElementById('invTimelineContainer');
                    if (timelineContainer) {
                        try {
                            const timeRes = await fetch(`${API_BASE}/api/investigations/${invId}/timeline`);
                            if (timeRes.ok) {
                                const timeData = await timeRes.json();
                                const events = Array.isArray(timeData.events) ? timeData.events : [];
                                if (events.length > 0) {
                                    timelineContainer.innerHTML = events.map(ev => `
                                        <div class="signal-item">
                                            <span class="signal-time">${escapeHtml(ev.event_timestamp || '')}</span>
                                            <div class="signal-body">
                                                <strong class="signal-title">${escapeHtml(ev.headline || ev.event_type || '')}</strong>
                                                <p class="signal-desc">${escapeHtml(ev.details || '')}</p>
                                            </div>
                                        </div>
                                    `).join('');
                                    return;
                                }
                            }
                        } catch (e) {
                            console.warn('Gagal memuat timeline:', e);
                        }
                        const isEn = currentLang === 'en';
                        timelineContainer.innerHTML = `
                            <div class="dossier-empty-state">
                                <span>${isEn ? 'No timeline events recorded yet. Chronological events will appear here once the investigation is executed.' : 'Belum ada kronologi peristiwa. Peristiwa kronologis akan tercatat otomatis saat investigasi dijalankan.'}</span>
                            </div>
                        `;
                    }
                }

                // Sidebar Navigation Links
                document.querySelectorAll('.nav-link-item').forEach(link => {
                    link.addEventListener('click', (e) => {
                        e.preventDefault();
                        const nav = link.getAttribute('data-nav');
                        if (nav) switchMainView(nav);
                    });
                });

                // Settings Modal Open/Close & Tabs
                function openSettingsModal() {
                    if (!settingsModal) return;
                    settingsModal.style.display = 'flex';
                    fetchSettingsData();
                    fetchTelegramStatus();
                    fetchSectorsUsage();
                    fetchDiagnostics();
                }

                function closeSettingsModal() {
                    if (settingsModal) settingsModal.style.display = 'none';
                }

                if (btnSettings) {
                    btnSettings.addEventListener('click', (e) => {
                        e.preventDefault();
                        openSettingsModal();
                    });
                }

                if (btnCloseSettingsModal) btnCloseSettingsModal.addEventListener('click', closeSettingsModal);
                if (btnCancelSettings) btnCancelSettings.addEventListener('click', closeSettingsModal);

                settingsModal.addEventListener('click', (e) => {
                    if (e.target === settingsModal) closeSettingsModal();
                });

                // Tabs Switcher
                modalTabs.forEach(tab => {
                    tab.addEventListener('click', () => {
                        modalTabs.forEach(t => t.classList.remove('active'));
                        tab.classList.add('active');
                        const targetTab = tab.getAttribute('data-tab');
                        document.querySelectorAll('.tab-pane').forEach(pane => {
                            pane.style.display = (pane.id === `pane-${targetTab}`) ? 'block' : 'none';
                        });
                    });
                });

                // Inference Timeout Slider Handlers
                function onTimeoutSliderInput(val) {
                    const num = parseFloat(val) || 60;
                    const valueEl = document.getElementById('timeout-value-badge');
                    const badgeEl = document.getElementById('timeout-profile-badge');
                    if (valueEl) {
                        valueEl.textContent = `${Math.round(num)}s`;
                    }
                    let profileKey = 'custom';
                    if (num <= 25) {
                        profileKey = 'fast';
                    } else if (num <= 65) {
                        profileKey = 'balanced';
                    } else if (num <= 125) {
                        profileKey = 'deep';
                    } else if (num <= 185) {
                        profileKey = 'local';
                    } else {
                        profileKey = 'custom';
                    }
                    if (badgeEl) {
                        badgeEl.textContent = `[${t('profile_' + profileKey)}]`;
                    }
                }
                window.onTimeoutSliderInput = onTimeoutSliderInput;

                function populateTimeoutSlider(timeoutSecs) {
                    const slider = document.getElementById('timeout-slider');
                    const val = (timeoutSecs !== undefined && timeoutSecs !== null && !isNaN(timeoutSecs)) ? parseFloat(timeoutSecs) : 60;
                    if (slider) {
                        slider.value = val;
                    }
                    onTimeoutSliderInput(val);
                }
                window.populateTimeoutSlider = populateTimeoutSlider;

                // Fetch Settings & Populate Form
                async function fetchSettingsData() {
                    try {
                        const res = await fetch(`${API_BASE}/api/settings?reveal=true`);
                        if (!res.ok) return;
                        const data = await res.json();
                        lastSettingsData = data;

                        if (data.auth) {
                            if (selectAiProvider && data.auth.ai_provider) {
                                selectAiProvider.value = data.auth.ai_provider;
                            }
                            if (inputGeminiModel && data.auth.gemini_model) {
                                inputGeminiModel.value = data.auth.gemini_model;
                            }
                            if (statusSectorsKey) {
                                if (data.auth.has_sectors_key) {
                                    statusSectorsKey.textContent = t('tag_saved');
                                    statusSectorsKey.style.color = '#10B981';
                                    if (inputSectorsKey) {
                                        inputSectorsKey.value = data.auth.sectors_api_key || '';
                                        inputSectorsKey.dataset.saved = 'true';
                                        inputSectorsKey.placeholder = `sec_live_... (${t('tag_saved')})`;
                                    }
                                } else {
                                    statusSectorsKey.textContent = t('tag_no_key');
                                    statusSectorsKey.style.color = '#F59E0B';
                                    if (inputSectorsKey) {
                                        inputSectorsKey.value = '';
                                        delete inputSectorsKey.dataset.saved;
                                        inputSectorsKey.placeholder = 'sec_live_... (Kosongkan jika tidak diubah)';
                                    }
                                }
                            }
                            const tab1SectorsBadge = document.getElementById('tab1SectorsBadge');
                            if (tab1SectorsBadge) {
                                if (data.auth.has_sectors_key) {
                                    tab1SectorsBadge.textContent = t('tag_saved');
                                    tab1SectorsBadge.style.color = '#10B981';
                                } else {
                                    tab1SectorsBadge.textContent = t('tag_no_key');
                                    tab1SectorsBadge.style.color = '#F59E0B';
                                }
                            }
                            if (statusGeminiKey) {
                                if (data.auth.has_gemini_key) {
                                    statusGeminiKey.textContent = t('tag_saved');
                                    statusGeminiKey.style.color = '#10B981';
                                    if (inputGeminiKey) {
                                        inputGeminiKey.value = data.auth.gemini_api_key || '';
                                        inputGeminiKey.dataset.saved = 'true';
                                        inputGeminiKey.placeholder = `AIzaSy... (${t('tag_saved')})`;
                                    }
                                } else {
                                    statusGeminiKey.textContent = t('tag_no_key');
                                    statusGeminiKey.style.color = '#F59E0B';
                                    if (inputGeminiKey) {
                                        inputGeminiKey.value = '';
                                        delete inputGeminiKey.dataset.saved;
                                        inputGeminiKey.placeholder = 'AIzaSy... (Kosongkan jika tidak diubah)';
                                    }
                                }
                            }
                            if (statusOpenaiKey) {
                                if (data.auth.has_openai_key) {
                                    statusOpenaiKey.textContent = t('tag_saved');
                                    statusOpenaiKey.style.color = '#10B981';
                                    if (inputOpenaiKey) {
                                        inputOpenaiKey.value = data.auth.openai_api_key || '';
                                        inputOpenaiKey.dataset.saved = 'true';
                                        inputOpenaiKey.placeholder = `sk-... (${t('tag_saved')})`;
                                    }
                                } else {
                                    statusOpenaiKey.textContent = t('tag_no_key');
                                    statusOpenaiKey.style.color = 'var(--text-muted)';
                                    if (inputOpenaiKey) {
                                        inputOpenaiKey.value = '';
                                        delete inputOpenaiKey.dataset.saved;
                                        inputOpenaiKey.placeholder = 'sk-... (Kosongkan jika tidak diubah)';
                                    }
                                }
                            }
                            if (inputOpenaiBaseUrl && data.auth.openai_base_url) {
                                inputOpenaiBaseUrl.value = data.auth.openai_base_url;
                            }
                            if (inputOpenaiModel && data.auth.openai_model) {
                                inputOpenaiModel.value = data.auth.openai_model;
                            }
                            if (inputOllamaBaseUrl && data.auth.ollama_base_url) {
                                inputOllamaBaseUrl.value = data.auth.ollama_base_url;
                            }
                            if (inputOllamaModel && data.auth.ollama_model) {
                                inputOllamaModel.value = data.auth.ollama_model;
                            }
                            if (statusOllamaKey) {
                                statusOllamaKey.textContent = data.auth.ollama_base_url || 'http://localhost:11434';
                                statusOllamaKey.style.color = '#10B981';
                            }
                            if (inputAnthropicModel && data.auth.anthropic_model) {
                                inputAnthropicModel.value = data.auth.anthropic_model;
                            }
                            if (statusAnthropicKey) {
                                if (data.auth.has_anthropic_key) {
                                    statusAnthropicKey.textContent = t('tag_saved');
                                    statusAnthropicKey.style.color = '#10B981';
                                    if (inputAnthropicKey) {
                                        inputAnthropicKey.value = data.auth.anthropic_api_key || '';
                                        inputAnthropicKey.dataset.saved = 'true';
                                        inputAnthropicKey.placeholder = `sk-ant-... (${t('tag_saved')})`;
                                    }
                                } else {
                                    statusAnthropicKey.textContent = t('tag_no_key');
                                    statusAnthropicKey.style.color = 'var(--text-muted)';
                                    if (inputAnthropicKey) {
                                        inputAnthropicKey.value = '';
                                        delete inputAnthropicKey.dataset.saved;
                                        inputAnthropicKey.placeholder = 'sk-ant-... (Kosongkan jika tidak diubah)';
                                    }
                                }
                            }
                            // Update reactive provider card badges
                            const tagGemini = document.getElementById('tagGeminiStatus');
                            if (tagGemini) {
                                if (data.auth && data.auth.has_gemini_key) {
                                    tagGemini.textContent = t('tag_saved');
                                    tagGemini.style.color = '#10B981';
                                    tagGemini.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagGemini.textContent = t('tag_recommended');
                                    tagGemini.style.color = '#03A66D';
                                    tagGemini.style.background = 'rgba(3, 166, 109, 0.12)';
                                }
                            }

                            const tagOpenai = document.getElementById('tagOpenaiStatus');
                            if (tagOpenai) {
                                if (data.auth && data.auth.has_openai_key) {
                                    tagOpenai.textContent = t('tag_saved');
                                    tagOpenai.style.color = '#10B981';
                                    tagOpenai.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagOpenai.textContent = t('tag_gateway');
                                    tagOpenai.style.color = '';
                                    tagOpenai.style.background = '';
                                }
                            }

                            const tagAnthropic = document.getElementById('tagAnthropicStatus');
                            if (tagAnthropic) {
                                if (data.auth && data.auth.has_anthropic_key) {
                                    tagAnthropic.textContent = t('tag_saved');
                                    tagAnthropic.style.color = '#10B981';
                                    tagAnthropic.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagAnthropic.textContent = t('tag_reasoning');
                                    tagAnthropic.style.color = '';
                                    tagAnthropic.style.background = '';
                                }
                            }

                            const tagOllama = document.getElementById('tagOllamaStatus');
                            if (tagOllama) {
                                const ollamaUrl = (data.auth && data.auth.ollama_base_url) || (inputOllamaBaseUrl ? inputOllamaBaseUrl.value.trim() : '') || 'http://localhost:11434';
                                if (ollamaUrl) {
                                    tagOllama.textContent = t('tag_available');
                                    tagOllama.style.color = '#10B981';
                                    tagOllama.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagOllama.textContent = t('tag_local');
                                    tagOllama.style.color = '';
                                    tagOllama.style.background = '';
                                }
                            }

                            updateProviderVisibility();
                            updateActivePresetChips();
                        }
                        if (data.preferences) {
                            if (data.preferences.llm_timeout_secs) {
                                populateTimeoutSlider(data.preferences.llm_timeout_secs);
                            }
                        }

                        // Update Data Freshness Badge in Header
                        const badge = document.getElementById('dataFreshnessBadge');
                        const label = document.getElementById('labelFreshnessStatus');
                        if (badge && label) {
                            if (data.auth && data.auth.has_sectors_key) {
                                badge.className = 'data-freshness-badge';
                                label.textContent = 'IDX Live (EOD)';
                                badge.title = currentLang === 'en' ? 'Sectors Financial API v2 Active (Click to Flush Cache)' : 'Sectors Financial API v2 Aktif (Klik untuk Flush Cache)';
                            } else {
                                badge.className = 'data-freshness-badge cached';
                                label.textContent = currentLang === 'en' ? 'No Sectors Key' : 'Belum Ada Kunci IDX';
                                badge.title = currentLang === 'en' ? 'Configure SECTORS_API_KEY in Settings' : 'Konfigurasi SECTORS_API_KEY di Pengaturan';
                            }
                        }
                    } catch (e) {
                        console.warn('Gagal memuat konfigurasi settings:', e);
                    }

                }

                // Provider Live Connection Test
                async function testConnection(target, inputEl, statusBadge, baseUrlEl) {
                    if (statusBadge) {
                        statusBadge.textContent = currentLang === 'en' ? 'Testing connection...' : 'Menguji koneksi...';
                        statusBadge.style.color = 'var(--text-muted)';
                    }
                    try {
                        const payload = { target };
                        if (inputEl && inputEl.value.trim()) {
                            const val = inputEl.value.trim();
                            if (!val.includes('****')) {
                                payload.api_key = val;
                            }
                        }
                        if (baseUrlEl && baseUrlEl.value.trim()) {
                            payload.base_url = baseUrlEl.value.trim();
                        }
                        const res = await fetch(`${API_BASE}/api/settings/test-connection`, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify(payload)
                        });
                        const result = await res.json();
                        if (result.success) {
                            if (statusBadge) {
                                statusBadge.textContent = currentLang === 'en' ? `Connected (${result.latency_ms}ms)` : `Terhubung (${result.latency_ms}ms)`;
                                statusBadge.style.color = '#10B981';
                            }
                            if (target === 'sectors') {
                                const tab1Badge = document.getElementById('tab1SectorsBadge');
                                if (tab1Badge) {
                                    tab1Badge.textContent = currentLang === 'en' ? `Connected (${result.latency_ms}ms)` : `Terhubung (${result.latency_ms}ms)`;
                                    tab1Badge.style.color = '#10B981';
                                }
                            } else if (target === 'gemini') {
                                const tag = document.getElementById('tagGeminiStatus');
                                if (tag) {
                                    tag.textContent = currentLang === 'en' ? 'Ready' : 'Siap';
                                    tag.style.color = '#10B981';
                                    tag.style.background = 'rgba(16, 185, 129, 0.12)';
                                }
                            } else if (target === 'openai') {
                                const tag = document.getElementById('tagOpenaiStatus');
                                if (tag) {
                                    tag.textContent = currentLang === 'en' ? 'Ready' : 'Siap';
                                    tag.style.color = '#10B981';
                                    tag.style.background = 'rgba(16, 185, 129, 0.12)';
                                }
                            } else if (target === 'anthropic') {
                                const tag = document.getElementById('tagAnthropicStatus');
                                if (tag) {
                                    tag.textContent = currentLang === 'en' ? 'Ready' : 'Siap';
                                    tag.style.color = '#10B981';
                                    tag.style.background = 'rgba(16, 185, 129, 0.12)';
                                }
                            } else if (target === 'ollama') {
                                const tag = document.getElementById('tagOllamaStatus');
                                if (tag) {
                                    tag.textContent = currentLang === 'en' ? 'Available' : 'Tersedia';
                                    tag.style.color = '#10B981';
                                    tag.style.background = 'rgba(16, 185, 129, 0.12)';
                                }
                            }
                            showToast(`${target.toUpperCase()}: ${result.message} (${result.latency_ms}ms)`);
                        } else {
                            if (statusBadge) {
                                statusBadge.textContent = currentLang === 'en' ? 'Failed' : 'Gagal';
                                statusBadge.style.color = '#EF4444';
                            }
                            showToast(currentLang === 'en' ? `${target.toUpperCase()} Failed: ${result.message || 'Connection refused'}` : `${target.toUpperCase()} Gagal: ${result.message || 'Koneksi ditolak'}`, true);
                        }
                    } catch (e) {
                        if (statusBadge) {
                            statusBadge.textContent = 'Error';
                            statusBadge.style.color = '#EF4444';
                        }
                        showToast(currentLang === 'en' ? `Network error: ${e.message}` : `Kesalahan jaringan: ${e.message}`, true);
                    }
                }

                const btnGoToSectorsTab = document.getElementById('btnGoToSectorsTab');
                if (btnGoToSectorsTab) {
                    btnGoToSectorsTab.addEventListener('click', (e) => {
                        e.preventDefault();
                        const tabBtnSectors = document.getElementById('tabBtnSectors');
                        if (tabBtnSectors) {
                            tabBtnSectors.click();
                            if (inputSectorsKey) inputSectorsKey.focus();
                        }
                    });
                }

                if (btnTestSectors) {
                    btnTestSectors.addEventListener('click', (e) => {
                        e.preventDefault();
                        testConnection('sectors', inputSectorsKey, statusSectorsKey);
                    });
                }

                if (btnTestGemini) {
                    btnTestGemini.addEventListener('click', (e) => {
                        e.preventDefault();
                        testConnection('gemini', inputGeminiKey, statusGeminiKey);
                    });
                }

                if (btnTestOpenai) {
                    btnTestOpenai.addEventListener('click', (e) => {
                        e.preventDefault();
                        testConnection('openai', inputOpenaiKey, statusOpenaiKey, inputOpenaiBaseUrl);
                    });
                }

                if (btnTestOllama) {
                    btnTestOllama.addEventListener('click', (e) => {
                        e.preventDefault();
                        testConnection('ollama', null, statusOllamaKey, inputOllamaBaseUrl);
                    });
                }

                if (btnTestAnthropic) {
                    btnTestAnthropic.addEventListener('click', (e) => {
                        e.preventDefault();
                        testConnection('anthropic', inputAnthropicKey, statusAnthropicKey);
                    });
                }

                // Telegram Whitelist Chips & Bot Controls
                let currentTeleUsers = [];
                let isSavingTeleUsers = false;
                const teleAllowedUsersChips = document.getElementById('teleAllowedUsersChips');
                const inputAddTeleUser = document.getElementById('inputAddTeleUser');
                const btnAddTeleUser = document.getElementById('btnAddTeleUser');
                const telePollerHeartbeat = document.getElementById('telePollerHeartbeat');

                async function persistTelegramAllowedUsers(users) {
                    try {
                        isSavingTeleUsers = true;
                        const res = await fetch(`${API_BASE}/api/settings/telegram`, {
                            method: 'PATCH',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ allowed_users: users || [] })
                        });
                        if (res.ok) {
                            showToast(currentLang === 'en' ? 'Telegram whitelist autosaved' : 'Whitelist Telegram tersimpan otomatis');
                        } else {
                            const err = await res.json().catch(() => ({}));
                            showToast(err.error || (currentLang === 'en' ? 'Failed to save whitelist' : 'Gagal menyimpan whitelist'), true);
                        }
                    } catch (e) {
                        console.warn('Failed to persist telegram allowed users:', e);
                        showToast(currentLang === 'en' ? 'Error saving whitelist' : 'Kesalahan menyimpan whitelist', true);
                    } finally {
                        setTimeout(() => {
                            isSavingTeleUsers = false;
                        }, 500);
                    }
                }

                function renderTeleUserChips() {
                    if (!teleAllowedUsersChips) return;
                    teleAllowedUsersChips.innerHTML = '';
                    if (currentTeleUsers.length === 0) {
                        teleAllowedUsersChips.innerHTML = `<span style="font-size:11px; color:var(--text-muted); font-style:italic;">${t('tele_no_whitelist')}</span>`;
                    } else {
                        currentTeleUsers.forEach((u, idx) => {
                            const chip = document.createElement('div');
                            chip.className = 'tele-chip';
                            chip.innerHTML = `
                                <span>${escapeHtml(u)}</span>
                                <button class="tele-chip-remove" type="button" data-idx="${idx}" title="Hapus User">&times;</button>
                            `;
                            chip.querySelector('.tele-chip-remove').addEventListener('click', () => {
                                currentTeleUsers.splice(idx, 1);
                                renderTeleUserChips();
                                persistTelegramAllowedUsers(currentTeleUsers);
                            });
                            teleAllowedUsersChips.appendChild(chip);
                        });
                    }
                    if (inputTeleUsers) {
                        inputTeleUsers.value = currentTeleUsers.join(', ');
                    }
                }

                function addTeleUserFromInput() {
                    if (!inputAddTeleUser) return;
                    const val = inputAddTeleUser.value.trim().replace(/^@/, '');
                    if (!val) return;
                    if (!currentTeleUsers.includes(val)) {
                        currentTeleUsers.push(val);
                        renderTeleUserChips();
                        persistTelegramAllowedUsers(currentTeleUsers);
                    }
                    inputAddTeleUser.value = '';
                }

                if (btnAddTeleUser) {
                    btnAddTeleUser.addEventListener('click', (e) => {
                        e.preventDefault();
                        addTeleUserFromInput();
                    });
                }

                if (inputAddTeleUser) {
                    inputAddTeleUser.addEventListener('keydown', (e) => {
                        if (e.key === 'Enter') {
                            e.preventDefault();
                            addTeleUserFromInput();
                        }
                    });
                }

                async function fetchTelegramStatus() {
                    try {
                        const res = await fetch(`${API_BASE}/api/settings/telegram`);
                        if (!res.ok) return;
                        const data = await res.json();
                        lastTelegramData = data;
                        const isRunning = (data.status === 'RUNNING');

                        if (teleBotStatusBadge) {
                            teleBotStatusBadge.textContent = data.status || 'STANDBY';
                            if (isRunning) {
                                teleBotStatusBadge.style.background = '#ECFDF5';
                                teleBotStatusBadge.style.color = '#047857';
                            } else {
                                teleBotStatusBadge.style.background = 'var(--pill-active-bg)';
                                teleBotStatusBadge.style.color = 'var(--text-secondary)';
                            }
                        }
                        if (teleBotUsername) {
                            teleBotUsername.textContent = data.bot_username ? `@${data.bot_username}` : '@NiskavaAgentBot';
                        }
                        if (telePollerHeartbeat) {
                            telePollerHeartbeat.classList.toggle('active', isRunning);
                        }
                        if (statusTeleToken) {
                            if (data.has_token) {
                                statusTeleToken.textContent = t('tele_token_saved');
                                statusTeleToken.style.color = '#10B981';
                                if (inputTeleToken) {
                                    inputTeleToken.disabled = false;
                                    inputTeleToken.value = data.bot_token || '';
                                    inputTeleToken.dataset.saved = 'true';
                                    inputTeleToken.placeholder = `7123456789:AAH... (${t('tag_saved')})`;
                                }
                                if (btnUnlockTeleToken) {
                                    btnUnlockTeleToken.style.display = 'none';
                                }
                            } else {
                                statusTeleToken.textContent = t('tele_token_empty');
                                statusTeleToken.style.color = 'var(--text-muted)';
                                if (inputTeleToken) {
                                    inputTeleToken.disabled = false;
                                    inputTeleToken.value = '';
                                    delete inputTeleToken.dataset.saved;
                                    inputTeleToken.placeholder = '7123456789:AAH... (Kosongkan jika tidak diubah)';
                                }
                                if (btnUnlockTeleToken) {
                                    btnUnlockTeleToken.style.display = 'none';
                                }
                            }
                        }
                        if (btnUnlockTeleToken) {
                            btnUnlockTeleToken.textContent = t('btn_unlock_tele_token');
                        }

                        // Dynamic Start / Stop Bot Button Toggle
                        if (isRunning) {
                            if (btnStartTeleBot) btnStartTeleBot.style.display = 'none';
                            if (btnStopTeleBot) {
                                btnStopTeleBot.style.display = 'inline-flex';
                                btnStopTeleBot.disabled = false;
                                btnStopTeleBot.style.opacity = '1';
                                btnStopTeleBot.style.cursor = 'pointer';
                                btnStopTeleBot.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg><span>${t('btn_stop_bot')}</span>`;
                            }
                        } else {
                            if (btnStopTeleBot) btnStopTeleBot.style.display = 'none';
                            if (btnStartTeleBot) {
                                btnStartTeleBot.style.display = 'inline-flex';
                                btnStartTeleBot.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg><span>${t('btn_start_bot')}</span>`;
                                if (!data.has_token) {
                                    btnStartTeleBot.disabled = true;
                                    btnStartTeleBot.style.opacity = '0.5';
                                    btnStartTeleBot.style.cursor = 'not-allowed';
                                    btnStartTeleBot.title = currentLang === 'en' ? 'Configure Telegram Bot Token first' : 'Konfigurasi Token Bot terlebih dahulu';
                                } else {
                                    btnStartTeleBot.disabled = false;
                                    btnStartTeleBot.style.opacity = '1';
                                    btnStartTeleBot.style.cursor = 'pointer';
                                    btnStartTeleBot.title = '';
                                }
                            }
                        }

                        // Anti-clobbering sync: do not overwrite local state if user is typing or save is in flight
                        const isInputFocused = (inputAddTeleUser && document.activeElement === inputAddTeleUser);
                        if (!isSavingTeleUsers && !isInputFocused) {
                            let serverUsers = [];
                            if (data.allowed_users && Array.isArray(data.allowed_users)) {
                                serverUsers = data.allowed_users.map(s => String(s).trim()).filter(Boolean);
                            } else if (inputTeleUsers && inputTeleUsers.value) {
                                serverUsers = inputTeleUsers.value.split(',').map(s => s.trim()).filter(Boolean);
                            }

                            // Avoid redundant DOM replacement if server array matches current local state
                            const isDifferent = (serverUsers.length !== currentTeleUsers.length) ||
                                serverUsers.some((u, idx) => u !== currentTeleUsers[idx]);

                            if (isDifferent) {
                                currentTeleUsers = serverUsers;
                                renderTeleUserChips();
                            }
                        }
                    } catch (e) {
                        console.warn('Gagal memuat status telegram:', e);
                    }
                }

                if (btnStartTeleBot) {
                    btnStartTeleBot.addEventListener('click', async (e) => {
                        e.preventDefault();
                        try {
                            btnStartTeleBot.disabled = true;
                            btnStartTeleBot.innerHTML = `<span>${currentLang === 'en' ? 'Starting...' : 'Menyalakan...'}</span>`;
                            const res = await fetch(`${API_BASE}/api/telegram/start`, { method: 'POST' });
                            const data = await res.json();
                            if (res.ok) {
                                showToast(data.message || (currentLang === 'en' ? 'Telegram Bot started' : 'Telegram Bot dimulai'));
                            } else {
                                showToast(data.error || (currentLang === 'en' ? 'Failed to start bot' : 'Gagal memulai bot'), true);
                            }
                            await fetchTelegramStatus();
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Failed to start bot: ${err.message}` : `Gagal memulai bot: ${err.message}`, true);
                            await fetchTelegramStatus();
                        }
                    });
                }

                if (btnStopTeleBot) {
                    btnStopTeleBot.addEventListener('click', async (e) => {
                        e.preventDefault();
                        try {
                            btnStopTeleBot.disabled = true;
                            btnStopTeleBot.innerHTML = `<span>${currentLang === 'en' ? 'Stopping...' : 'Menghentikan...'}</span>`;
                            const res = await fetch(`${API_BASE}/api/telegram/stop`, { method: 'POST' });
                            const data = await res.json();
                            if (res.ok) {
                                showToast(data.message || (currentLang === 'en' ? 'Telegram Bot stopped' : 'Telegram Bot dihentikan'));
                            } else {
                                showToast(data.error || (currentLang === 'en' ? 'Failed to stop bot' : 'Gagal menghentikan bot'), true);
                            }
                            await fetchTelegramStatus();
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Failed to stop bot: ${err.message}` : `Gagal menghentikan bot: ${err.message}`, true);
                            await fetchTelegramStatus();
                        }
                    });
                }

                if (btnSendTeleTest) {
                    btnSendTeleTest.addEventListener('click', async (e) => {
                        e.preventDefault();
                        const chatIdVal = inputTeleTestChatId ? inputTeleTestChatId.value.trim() : '';
                        if (!chatIdVal) {
                            showToast(currentLang === 'en' ? 'Please enter Telegram Chat ID' : 'Harap masukkan Chat ID Telegram', true);
                            return;
                        }
                        try {
                            const res = await fetch(`${API_BASE}/api/telegram/test`, {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({ chat_id: Number(chatIdVal) || chatIdVal })
                            });
                            const data = await res.json();
                            if (res.ok) {
                                showToast(currentLang === 'en' ? 'Test message sent to Telegram successfully!' : 'Pesan uji coba berhasil dikirim ke Telegram!');
                            } else {
                                showToast(currentLang === 'en' ? `Send failed: ${data.error || 'Check Chat ID and token'}` : `Gagal kirim: ${data.error || 'Periksa Chat ID dan token'}`, true);
                            }
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Send error: ${err.message}` : `Error kirim: ${err.message}`, true);
                        }
                    });
                }

                // Sectors Cache Management (Law 5)
                async function fetchSectorsUsage() {
                    try {
                        const res = await fetch(`${API_BASE}/api/system/sectors-usage`);
                        if (!res.ok) return;
                        const data = await res.json();
                        if (statCacheTotal) statCacheTotal.textContent = data.total_entries ?? 0;
                        if (statCreditSaved) statCreditSaved.textContent = data.estimated_credit_saved ?? 0;
                        if (statCachePermanent) statCachePermanent.textContent = data.permanent_entries ?? 0;
                        if (statCacheExpired) statCacheExpired.textContent = data.expired_entries ?? 0;

                        // Sectors v2 Credit Progress Bar (Law 5 Credit Preservation)
                        const total = data.total_entries ?? 0;
                        const creditFill = document.getElementById('sectorsCreditFill');
                        const creditText = document.getElementById('statCreditUsageText');
                        const hitRatio = document.getElementById('statCacheHitRatio');
                        const creditRemaining = document.getElementById('statCreditRemaining');

                        const pct = Math.min(100, Math.round((total / 1000) * 100));
                        if (creditFill) {
                            creditFill.style.width = `${pct}%`;
                            creditFill.style.background = pct > 80 ? '#EF4444' : 'var(--primary)';
                        }
                        if (creditText) creditText.textContent = `${total} / 1.000`;
                        if (hitRatio) {
                            const ratio = total > 0 ? (((data.permanent_entries || total) / total) * 100).toFixed(0) : 100;
                            hitRatio.textContent = `${ratio}%`;
                        }
                        if (creditRemaining) {
                            creditRemaining.textContent = currentLang === 'en' ? `Est. Remaining: ~${Math.max(0, 1000 - total)} calls` : `Sisa Estimasi: ~${Math.max(0, 1000 - total)} panggilan`;
                        }
                    } catch (e) {
                        console.warn('Gagal memuat status cache sectors:', e);
                    }
                }

                if (btnCleanCache) {
                    btnCleanCache.addEventListener('click', async (e) => {
                        e.preventDefault();
                        try {
                            const res = await fetch(`${API_BASE}/api/system/cache/clean`, { method: 'POST' });
                            const data = await res.json();
                            showToast(currentLang === 'en' ? `Cache cleaned: ${data.cleaned_entries ?? 0} expired entries removed` : `Cache dibersihkan: ${data.cleaned_entries ?? 0} entri kadaluarsa dihapus`);
                            fetchSectorsUsage();
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Failed to clean cache: ${err.message}` : `Gagal bersihkan cache: ${err.message}`, true);
                        }
                    });
                }

                const btnFlushAllCache = document.getElementById('btnFlushAllCache');
                if (btnFlushAllCache) {
                    btnFlushAllCache.addEventListener('click', async (e) => {
                        e.preventDefault();
                        const isEn = (currentLang === 'en');
                        const confirmMsg = isEn ?
                            'Flush ALL cache (including permanent historical OHLCV candles)? Subsequent queries will consume Sectors API credits.' :
                            'Flush SEMUA cache (termasuk candlestick OHLCV historis)? Kueri berikutnya akan membutuhkan kuota kredit Sectors API.';
                        if (!confirm(confirmMsg)) return;

                        try {
                            btnFlushAllCache.disabled = true;
                            btnFlushAllCache.textContent = isEn ? 'Flushing...' : 'Memproses...';
                            const res = await fetch(`${API_BASE}/api/system/cache/clean?all=1`, { method: 'POST' });
                            const data = await res.json();
                            showToast(isEn ? `All cache flushed (${data.cleaned_entries ?? 0} entries removed)` : `Seluruh cache di-flush (${data.cleaned_entries ?? 0} entri dihapus)`);
                            fetchSectorsUsage();
                        } catch (err) {
                            showToast(isEn ? `Failed to flush cache: ${err.message}` : `Gagal flush cache: ${err.message}`, true);
                        } finally {
                            btnFlushAllCache.disabled = false;
                            btnFlushAllCache.textContent = isEn ? 'Flush All Cache' : 'Flush Semua Cache';
                        }
                    });
                }

                // System Diagnostics
                function renderDiagnosticsHTML(data) {
                    if (!diagnosticsDetails || !data) return;
                    const dbSizeKb = data.database_size_bytes ? (data.database_size_bytes / 1024).toFixed(1) + ' KB' : 'N/A';
                    const isEn = (currentLang === 'en');

                    diagnosticsDetails.innerHTML = `
                        <div class="kpi-stats-grid" style="grid-template-columns: repeat(3, 1fr);">
                            <div class="kpi-stat-card">
                                <span class="kpi-stat-label">${t('diag_status_daemon')}</span>
                                <span class="kpi-stat-num" style="color:#10B981; font-size:14px;">${escapeHtml(data.status || 'OK')}</span>
                                <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${t('diag_local_ipc')}</span>
                            </div>
                            <div class="kpi-stat-card">
                                <span class="kpi-stat-label">${t('diag_go_runtime')}</span>
                                <span class="kpi-stat-num" style="font-size:14px;">${escapeHtml(data.go_version || 'Go')}</span>
                                <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${escapeHtml(data.os || '')} (${escapeHtml(data.arch || '')}) | ${data.num_cpu || 1} CPU</span>
                            </div>
                            <div class="kpi-stat-card">
                                <span class="kpi-stat-label">${t('diag_sqlite_wal')}</span>
                                <span class="kpi-stat-num" style="font-size:14px;">${dbSizeKb}</span>
                                <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${data.total_sessions || 0} ${t('diag_sessions_saved')}</span>
                            </div>
                        </div>
                        <div class="kpi-stats-grid" style="grid-template-columns: repeat(2, 1fr); margin-top:0;">
                            <div class="kpi-stat-card">
                                <span class="kpi-stat-label">${t('diag_active_provider')}</span>
                                <span class="kpi-stat-num" style="color:var(--accent-text); font-size:14px;">${escapeHtml((data.ai_provider || 'openai').toUpperCase())}</span>
                                <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">ReAct Cognitive Engine</span>
                            </div>
                            <div class="kpi-stat-card">
                                <span class="kpi-stat-label">${t('diag_op_mode')}</span>
                                <span class="kpi-stat-num" style="color:#10B981; font-size:14px;">SECTORS v2 LIVE</span>
                                <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${t('diag_credit_sync')}</span>
                            </div>
                        </div>
                        <div class="settings-card" style="margin-top:2px;">
                            <div class="settings-card-head">
                                <div class="settings-card-title">
                                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
                                    <span>${t('settings_ssot_paths_title')}</span>
                                </div>
                                <span class="panel-card-badge">${t('diag_local_first_badge')}</span>
                            </div>
                            <div style="display:flex; flex-direction:column; gap:6px; font-family:var(--font-mono); font-size:11px;">
                                <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; background:var(--bg-card); border-radius:6px; border:1px solid var(--border-subtle);">
                                    <div><strong style="color:var(--text-primary);">Database:</strong> <span style="color:var(--text-secondary);">${escapeHtml(data.database_path || '~/.niskava/niskava.db')}</span></div>
                                    <button type="button" class="btn-copy-path" data-path="${escapeHtml(data.database_path || '~/.niskava/niskava.db')}" title="${t('settings_copy_path')}" style="cursor:pointer; color:var(--text-muted); padding:2px 6px; border:none; background:transparent;">
                                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                    </button>
                                </div>
                                <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; background:var(--bg-card); border-radius:6px; border:1px solid var(--border-subtle);">
                                    <div><strong style="color:var(--text-primary);">Config:</strong> <span style="color:var(--text-secondary);">${escapeHtml(data.config_path || '~/.niskava/config.yaml')}</span></div>
                                    <button type="button" class="btn-copy-path" data-path="${escapeHtml(data.config_path || '~/.niskava/config.yaml')}" title="${t('settings_copy_path')}" style="cursor:pointer; color:var(--text-muted); padding:2px 6px; border:none; background:transparent;">
                                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                    </button>
                                </div>
                                <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; background:var(--bg-card); border-radius:6px; border:1px solid var(--border-subtle);">
                                    <div><strong style="color:var(--text-primary);">DotEnv:</strong> <span style="color:var(--text-secondary);">${escapeHtml(data.dotenv_path || '~/.niskava/.env')}</span></div>
                                    <button type="button" class="btn-copy-path" data-path="${escapeHtml(data.dotenv_path || '~/.niskava/.env')}" title="${t('settings_copy_path')}" style="cursor:pointer; color:var(--text-muted); padding:2px 6px; border:none; background:transparent;">
                                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                    </button>
                                </div>
                            </div>
                        </div>
                    `;

                    // Attach copy path events
                    diagnosticsDetails.querySelectorAll('.btn-copy-path').forEach(btn => {
                        btn.addEventListener('click', (e) => {
                            e.preventDefault();
                            const path = btn.getAttribute('data-path');
                            if (path) {
                                copyToClipboard(path).then((success) => {
                                    if (success) {
                                        showToast(isEn ? 'Path copied to clipboard' : 'Jalur disalin ke papan klip');
                                    } else {
                                        showToast(isEn ? 'Failed to copy path' : 'Gagal menyalin jalur', 'error');
                                    }
                                });
                            }
                        });
                    });
                }

                async function fetchDiagnostics() {
                    if (!diagnosticsDetails) return;
                    try {
                        const res = await fetch(`${API_BASE}/api/system/diagnostics`);
                        if (!res.ok) return;
                        const data = await res.json();
                        lastDiagnosticsData = data;

                        const sidebarProfileName = document.getElementById('sidebarProfileName');
                        const sidebarProfileAvatar = document.getElementById('sidebarProfileAvatar');
                        const uname = data.username || 'User';
                        if (sidebarProfileName) {
                            sidebarProfileName.textContent = (uname.toLowerCase() === 'user') ? t('profile_default_user') : uname;
                        }
                        if (sidebarProfileAvatar) {
                            sidebarProfileAvatar.textContent = (uname.toLowerCase() === 'user') ? 'US' : uname.slice(0, 2).toUpperCase();
                        }

                        renderDiagnosticsHTML(data);
                    } catch (e) {
                        diagnosticsDetails.innerHTML = `<div style="color:#EF4444; font-size:12px;">${currentLang === 'en' ? 'Failed to load diagnostics:' : 'Gagal memuat diagnostik:'} ${escapeHtml(e.message)}</div>`;
                    }
                }

                function updateSettingsLanguage() {
                    const slider = document.getElementById('timeout-slider');
                    if (slider) {
                        onTimeoutSliderInput(slider.value);
                    }
                    if (lastSettingsData) {
                        const data = lastSettingsData;
                        const tagGemini = document.getElementById('tagGeminiStatus');
                        if (tagGemini) {
                            if (data.auth && data.auth.has_gemini_key) {
                                tagGemini.textContent = t('tag_saved');
                            } else {
                                tagGemini.textContent = t('tag_recommended');
                            }
                        }
                        const tagOpenai = document.getElementById('tagOpenaiStatus');
                        if (tagOpenai) {
                            if (data.auth && data.auth.has_openai_key) {
                                tagOpenai.textContent = t('tag_saved');
                            } else {
                                tagOpenai.textContent = t('tag_gateway');
                            }
                        }
                        const tagAnthropic = document.getElementById('tagAnthropicStatus');
                        if (tagAnthropic) {
                            if (data.auth && data.auth.has_anthropic_key) {
                                tagAnthropic.textContent = t('tag_saved');
                            } else {
                                tagAnthropic.textContent = t('tag_reasoning');
                            }
                        }
                        const tagOllama = document.getElementById('tagOllamaStatus');
                        if (tagOllama) {
                            const ollamaUrl = (data.auth && data.auth.ollama_base_url) || (inputOllamaBaseUrl ? inputOllamaBaseUrl.value.trim() : '') || 'http://localhost:11434';
                            tagOllama.textContent = ollamaUrl ? t('tag_available') : t('tag_local');
                        }
                        if (statusSectorsKey) {
                            statusSectorsKey.textContent = (data.auth && data.auth.has_sectors_key) ? t('tag_saved') : t('tag_no_key');
                        }
                        const tab1SectorsBadge = document.getElementById('tab1SectorsBadge');
                        if (tab1SectorsBadge) {
                            tab1SectorsBadge.textContent = (data.auth && data.auth.has_sectors_key) ? t('tag_saved') : t('tag_no_key');
                        }
                        if (statusGeminiKey) {
                            statusGeminiKey.textContent = (data.auth && data.auth.has_gemini_key) ? t('tag_saved') : t('tag_no_key');
                        }
                        if (statusOpenaiKey) {
                            statusOpenaiKey.textContent = (data.auth && data.auth.has_openai_key) ? t('tag_saved') : t('tag_no_key');
                        }
                        if (statusAnthropicKey) {
                            statusAnthropicKey.textContent = (data.auth && data.auth.has_anthropic_key) ? t('tag_saved') : t('tag_no_key');
                        }
                        const badge = document.getElementById('dataFreshnessBadge');
                        const label = document.getElementById('labelFreshnessStatus');
                        if (badge && label) {
                            if (data.auth && data.auth.has_sectors_key) {
                                label.textContent = 'IDX Live (EOD)';
                                badge.title = currentLang === 'en' ? 'Sectors Financial API v2 Active (Click to Flush Cache)' : 'Sectors Financial API v2 Aktif (Klik untuk Flush Cache)';
                            } else {
                                label.textContent = currentLang === 'en' ? 'No Sectors Key' : 'Belum Ada Kunci IDX';
                                badge.title = currentLang === 'en' ? 'Configure SECTORS_API_KEY in Settings' : 'Konfigurasi SECTORS_API_KEY di Pengaturan';
                            }
                        }
                    }
                    if (lastTelegramData) {
                        const data = lastTelegramData;
                        if (statusTeleToken) {
                            statusTeleToken.textContent = data.has_token ? t('tele_token_saved') : t('tele_token_empty');
                        }
                        if (btnUnlockTeleToken) {
                            btnUnlockTeleToken.textContent = t('btn_unlock_tele_token');
                        }
                        const isRunning = (data.status === 'RUNNING');
                        if (btnStopTeleBot && isRunning) {
                            btnStopTeleBot.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg><span>${t('btn_stop_bot')}</span>`;
                        }
                        if (btnStartTeleBot && !isRunning) {
                            btnStartTeleBot.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg><span>${t('btn_start_bot')}</span>`;
                            if (!data.has_token) {
                                btnStartTeleBot.title = currentLang === 'en' ? 'Configure Telegram Bot Token first' : 'Konfigurasi Token Bot terlebih dahulu';
                            }
                        }
                        renderTeleUserChips();
                    }
                    const btnFlush = document.getElementById('btnFlushAllCache');
                    if (btnFlush) {
                        btnFlush.textContent = t('btn_flush_all_cache');
                    }
                    const sidebarProfileName = document.getElementById('sidebarProfileName');
                    if (sidebarProfileName && (!lastDiagnosticsData || !lastDiagnosticsData.username || lastDiagnosticsData.username.toLowerCase() === 'user')) {
                        sidebarProfileName.textContent = t('profile_default_user');
                    }
                    if (lastDiagnosticsData) {
                        renderDiagnosticsHTML(lastDiagnosticsData);
                    }
                }
                window.updateSettingsLanguage = updateSettingsLanguage;

                // Save Settings
                if (btnSaveSettings) {
                    btnSaveSettings.addEventListener('click', async (e) => {
                        e.preventDefault();
                        const payload = { auth: {}, preferences: {}, telegram: {} };

                        if (inputSectorsKey) {
                            const val = inputSectorsKey.value.trim();
                            if (!val.includes('****')) {
                                payload.auth.sectors_api_key = val;
                            }
                        }
                        if (selectAiProvider && selectAiProvider.value) {
                            payload.auth.ai_provider = selectAiProvider.value;
                        }
                        if (inputGeminiKey) {
                            const val = inputGeminiKey.value.trim();
                            if (!val.includes('****')) {
                                payload.auth.gemini_api_key = val;
                            }
                        }
                        if (inputGeminiModel && inputGeminiModel.value.trim()) {
                            payload.auth.gemini_model = inputGeminiModel.value.trim();
                        }
                        if (inputOpenaiKey) {
                            const val = inputOpenaiKey.value.trim();
                            if (!val.includes('****')) {
                                payload.auth.openai_api_key = val;
                            }
                        }
                        if (inputOpenaiBaseUrl && inputOpenaiBaseUrl.value.trim()) {
                            payload.auth.openai_base_url = inputOpenaiBaseUrl.value.trim();
                        }
                        if (inputOpenaiModel && inputOpenaiModel.value.trim()) {
                            payload.auth.openai_model = inputOpenaiModel.value.trim();
                        }
                        if (inputOllamaBaseUrl && inputOllamaBaseUrl.value.trim()) {
                            payload.auth.ollama_base_url = inputOllamaBaseUrl.value.trim();
                        }
                        if (inputOllamaModel && inputOllamaModel.value.trim()) {
                            payload.auth.ollama_model = inputOllamaModel.value.trim();
                        }
                        if (inputAnthropicKey) {
                            const val = inputAnthropicKey.value.trim();
                            if (!val.includes('****')) {
                                payload.auth.anthropic_api_key = val;
                            }
                        }
                        if (inputAnthropicModel && inputAnthropicModel.value.trim()) {
                            payload.auth.anthropic_model = inputAnthropicModel.value.trim();
                        }
                        if (inputTeleToken) {
                            const val = inputTeleToken.value.trim();
                            if (!val.includes('****')) {
                                payload.telegram.bot_token = val;
                            }
                        }
                        payload.telegram.allowed_users = currentTeleUsers;

                        if (document.getElementById('timeout-slider')) {
                            payload.preferences.llm_timeout_secs = parseFloat(document.getElementById('timeout-slider').value);
                        }

                        try {
                            btnSaveSettings.disabled = true;
                            btnSaveSettings.textContent = currentLang === 'en' ? 'Saving...' : 'Menyimpan...';
                            const res = await fetch(`${API_BASE}/api/settings?reveal=true`, {
                                method: 'PATCH',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify(payload)
                            });
                            if (res.ok) {
                                showToast(currentLang === 'en' ? 'Settings saved & hot-reloaded successfully!' : 'Pengaturan berhasil disimpan & hot-reloaded!');
                                closeSettingsModal();
                                fetchSettingsData();
                                fetchTelegramStatus();
                            } else {
                                const err = await res.json();
                                showToast(currentLang === 'en' ? `Save failed: ${err.error || 'An error occurred'}` : `Gagal simpan: ${err.error || 'Terjadi kesalahan'}`, true);
                            }
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Network error: ${err.message}` : `Kesalahan jaringan: ${err.message}`, true);
                        } finally {
                            btnSaveSettings.disabled = false;
                            btnSaveSettings.textContent = t('btn_save');
                        }
                    });
                }

                if (btnHeaderSettings) {
                    btnHeaderSettings.addEventListener('click', (e) => {
                        e.preventDefault();
                        openSettingsModal();
                    });
                }

                window.addEventListener('keydown', (e) => {
                    if ((e.metaKey || e.ctrlKey) && e.key === ',') {
                        e.preventDefault();
                        openSettingsModal();
                    } else if (e.key === 'Escape' && settingsModal && settingsModal.style.display !== 'none') {
                        e.preventDefault();
                        closeSettingsModal();
                    } else if ((e.metaKey || e.ctrlKey) && (e.key === 's' || e.key === 'S') && settingsModal && settingsModal.style.display !== 'none') {
                        e.preventDefault();
                        if (btnSaveSettings && !btnSaveSettings.disabled) {
                            btnSaveSettings.click();
                        }
                    }
                });

                // Preload diagnostics and settings on startup for profile and live badges
                fetchDiagnostics();
                fetchSettingsData();
            }
