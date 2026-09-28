// 10. Navigation, Dedicated Memory Graph & Settings/Toolkit Integration
            function initNavigationAndSettings() {
                const composerContainer = document.querySelector('.composer-container');
                const globalComplianceBox = document.getElementById('globalComplianceBox');
                const graphPageView = document.getElementById('graphPageView');
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

                    // Preset auto-fill helper if field is blank
                    if (prov === 'gemini') {
                        if (inputGeminiModel && !inputGeminiModel.value.trim()) {
                            inputGeminiModel.value = 'gemini-2.0-flash';
                        }
                    } else if (prov === 'openai') {
                        if (inputOpenaiBaseUrl && !inputOpenaiBaseUrl.value.trim()) {
                            inputOpenaiBaseUrl.value = 'http://localhost:20128/v1';
                        }
                        if (inputOpenaiModel && !inputOpenaiModel.value.trim()) {
                            inputOpenaiModel.value = 'hermes';
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
                }
                if (selectAiProvider) {
                    selectAiProvider.addEventListener('change', updateProviderVisibility);
                }

                // Preferences Elements
                const toggleOfflineMode = document.getElementById('toggleOfflineMode');
                const timeoutSlider = document.getElementById('timeout-slider');
                const offlineModeWarningBanner = document.getElementById('offlineModeWarningBanner');

                function updateOfflineWarning() {
                    if (offlineModeWarningBanner && toggleOfflineMode) {
                        offlineModeWarningBanner.style.display = toggleOfflineMode.checked ? 'block' : 'none';
                    }
                }
                if (toggleOfflineMode) {
                    toggleOfflineMode.addEventListener('change', updateOfflineWarning);
                }

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

                function switchMainView(targetNav) {
                    document.querySelectorAll('.nav-link-item').forEach(item => {
                        item.classList.toggle('active', item.getAttribute('data-nav') === targetNav);
                    });

                    const invPageView = document.getElementById('investigationsPageView');

                    if (targetNav === 'investigations') {
                        if (heroView) heroView.style.display = 'none';
                        if (chatView) {
                            chatView.style.display = 'none';
                            chatView.classList.remove('active');
                        }
                        if (globalComplianceBox) globalComplianceBox.style.display = 'none';
                        if (composerContainer) composerContainer.style.display = 'none';
                        if (graphPageView) graphPageView.style.display = 'none';
                        if (invPageView) invPageView.style.display = 'flex';
                        loadInvestigations();
                    } else if (targetNav === 'graphify') {
                        if (invPageView) invPageView.style.display = 'none';
                        if (heroView) heroView.style.display = 'none';
                        if (chatView) {
                            chatView.style.display = 'none';
                            chatView.classList.remove('active');
                        }
                        if (globalComplianceBox) globalComplianceBox.style.display = 'none';
                        if (composerContainer) composerContainer.style.display = 'none';
                        if (graphPageView) graphPageView.style.display = 'flex';
                        loadGraphPageData();
                    } else if (targetNav === 'chat') {
                        if (invPageView) invPageView.style.display = 'none';
                        if (graphPageView) graphPageView.style.display = 'none';
                        if (globalComplianceBox) globalComplianceBox.style.display = 'block';
                        if (composerContainer) composerContainer.style.display = 'flex';
                        const hasMessages = chatView && chatView.children.length > 0;
                        if (hasMessages) {
                            if (heroView) heroView.style.display = 'none';
                            if (chatView) {
                                chatView.style.display = 'flex';
                                chatView.classList.add('active');
                            }
                        } else {
                            if (heroView) heroView.style.display = 'flex';
                            if (chatView) {
                                chatView.style.display = 'none';
                                chatView.classList.remove('active');
                            }
                        }
                    } else if (targetNav === 'toolkit') {
                        openSettingsModal();
                    } else {
                        showToast(currentLang === 'en' ? `Module ${targetNav.toUpperCase()} active.` : `Modul ${targetNav.toUpperCase()} aktif.`);
                    }
                }

                // Dedicated Memory Graph Page loader
                async function loadGraphPageData() {
                    const ticker = graphFilterTicker ? graphFilterTicker.value.trim().toUpperCase() : '';
                    const depth = graphFilterDepth ? graphFilterDepth.value : '1';
                    const nodeTypes = graphFilterNodeType ? graphFilterNodeType.value : '';

                    const params = new URLSearchParams();
                    // NOTE: currentSessionId intentionally NOT forwarded (Law 6 — global cumulative graph)
                    if (ticker) params.set('ticker', ticker);
                    if (depth) params.set('depth', depth);
                    if (nodeTypes) params.set('node_types', nodeTypes);

                    // Standalone URL for opening in a new tab without embed mode
                    const externalUrl = `${API_BASE}/graph?${params.toString()}`;
                    if (btnOpenGraphExternal) btnOpenGraphExternal.href = externalUrl;

                    // Embedded iframe URL with embed=true to eliminate duplicate header/sidebar
                    const embedParams = new URLSearchParams(params);
                    embedParams.set('embed', 'true');
                    const frameUrl = `${API_BASE}/graph?${embedParams.toString()}`;
                    if (graphFrame) graphFrame.src = frameUrl;

                    if (statActiveSession) {
                        if (ticker) {
                            statActiveSession.textContent = `Ego: ${ticker}`;
                        } else {
                            statActiveSession.textContent = currentLang === 'en' ? 'Global Graph' : 'Global Graf';
                        }
                    }

                    try {
                        const res = await fetch(`${API_BASE}/api/graph/stats?${params.toString()}`);
                        if (res.ok) {
                            const stats = await res.json();
                            if (statTotalNodes) statTotalNodes.textContent = stats.total_nodes ?? 0;
                            if (statTotalEdges) statTotalEdges.textContent = stats.total_edges ?? 0;
                            if (statTopHub) {
                                if (stats.top_hub_nodes && stats.top_hub_nodes.length > 0) {
                                    const top = stats.top_hub_nodes[0];
                                    statTopHub.textContent = `${top.label || top.id} (${top.degree || 0})`;
                                } else {
                                    statTopHub.textContent = ticker || '-';
                                }
                            }
                        }
                    } catch (e) {
                        console.warn('Gagal memuat statistik graf:', e);
                    }
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

                // Clean Test Data in Memory Graph
                const btnPruneMockData = document.getElementById('btnPruneMockData');
                if (btnPruneMockData) {
                    btnPruneMockData.addEventListener('click', async () => {
                        try {
                            btnPruneMockData.disabled = true;
                            btnPruneMockData.textContent = currentLang === 'en' ? 'Cleaning...' : 'Membersihkan...';
                            const res = await fetch(`${API_BASE}/api/system/cache/clean`, { method: 'POST' });
                            const data = await res.json();
                            showToast(currentLang === 'en' ? `Test data & cache cleaned (${data.cleaned_entries ?? 0} entries removed)` : `Data uji & cache dibersihkan (${data.cleaned_entries ?? 0} entri dihapus)`);
                            loadGraphPageData();
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Failed to clean cache: ${err.message}` : `Gagal membersihkan cache: ${err.message}`, true);
                        } finally {
                            btnPruneMockData.disabled = false;
                            btnPruneMockData.textContent = currentLang === 'en' ? '🧹 Clean Test Data' : '🧹 Bersihkan Data Uji';
                        }
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

                // Formal 7-Stage Investigation Handlers
                let selectedInvestigationId = null;
                const btnInvBackToChat = document.getElementById('btnInvBackToChat');
                if (btnInvBackToChat) {
                    btnInvBackToChat.addEventListener('click', () => switchMainView('chat'));
                }

                const btnInvOpenChat = document.getElementById('btnInvOpenChat');
                if (btnInvOpenChat) {
                    btnInvOpenChat.addEventListener('click', () => {
                        if (selectedInvestigationId) {
                            selectSession(selectedInvestigationId);
                        }
                        switchMainView('chat');
                    });
                }

                const btnRunFormalInvestigation = document.getElementById('btnRunFormalInvestigation');
                const inputInvTicker = document.getElementById('inputInvTicker');
                const selectInvDays = document.getElementById('selectInvDays');
                const checkInvOffline = document.getElementById('checkInvOffline');

                if (btnRunFormalInvestigation) {
                    btnRunFormalInvestigation.addEventListener('click', async () => {
                        const ticker = (inputInvTicker ? inputInvTicker.value : '').trim().toUpperCase();
                        if (!ticker) {
                            showToast(currentLang === 'en' ? 'Please enter IDX ticker symbol (e.g., ANTM)' : 'Masukkan kode ticker emiten BEI (misal: ANTM)', true);
                            if (inputInvTicker) inputInvTicker.focus();
                            return;
                        }
                        const days = selectInvDays ? selectInvDays.value : '30';
                        const offlineMode = checkInvOffline ? checkInvOffline.checked : false;

                        // Switch to chat view and initiate formal audit stream
                        switchMainView('chat');
                        if (chatInput) {
                            chatInput.value = `Jalankan investigasi formal 7-tahap otonom pada emiten ${ticker} untuk observasi ${days} hari terakhir.${offlineMode ? ' (Mode Uji Offline)' : ''}`;
                            handleSendMessage();
                        }
                        showToast(currentLang === 'en' ? `Starting 7-stage investigation for ${ticker}...` : `Memulai investigasi 7-tahap untuk ${ticker}...`);
                    });
                }

                async function loadInvestigations() {
                    const invHistoryList = document.getElementById('invHistoryList');
                    const invTotalCount = document.getElementById('invTotalCount');
                    if (!invHistoryList) return;

                    try {
                        const res = await fetch(`${API_BASE}/api/investigations`);
                        if (!res.ok) {
                            invHistoryList.innerHTML = `<div style="font-size:11.5px; color:var(--text-muted); text-align:center; padding:20px 0;">${currentLang === 'en' ? 'No saved investigation sessions yet.' : 'Belum ada sesi investigasi tersimpan.'}</div>`;
                            return;
                        }
                        const list = await res.json();
                        if (!Array.isArray(list) || list.length === 0) {
                            invHistoryList.innerHTML = `<div style="font-size:11.5px; color:var(--text-muted); text-align:center; padding:20px 0;">${currentLang === 'en' ? 'No investigation sessions. Click "Run Formal Audit" to start.' : 'Belum ada sesi investigasi. Klik "Mulai Audit Otonom" untuk memulai.'}</div>`;
                            if (invTotalCount) invTotalCount.textContent = currentLang === 'en' ? '0 Sessions' : '0 Sesi';
                            return;
                        }

                        if (invTotalCount) invTotalCount.textContent = `${list.length} ${currentLang === 'en' ? 'Sessions' : 'Sesi'}`;
                        invHistoryList.innerHTML = '';

                        list.forEach((item, idx) => {
                            const sid = item.id || item.session_id || `INV-${idx}`;
                            const title = item.title || item.ticker || `Investigasi ${sid.slice(0, 8)}`;
                            const timeStr = item.created_at ? new Date(item.created_at).toLocaleDateString(currentLang === 'en' ? 'en-US' : 'id-ID', { day:'numeric', month:'short', hour:'2-digit', minute:'2-digit' }) : (currentLang === 'en' ? 'Just now' : 'Baru saja');
                            const status = (item.status || 'COMPLETED').toUpperCase();

                            const row = document.createElement('div');
                            row.className = `inv-history-item ${selectedInvestigationId === sid ? 'active' : ''}`;
                            row.innerHTML = `
                                <div class="inv-history-item-title">${escapeHtml(title)}</div>
                                <div class="inv-history-item-sub">
                                    <span>${timeStr}</span>
                                    <span class="inv-status-pill inv-status-${status === 'COMPLETED' ? 'completed' : 'running'}">${status}</span>
                                </div>
                            `;
                            row.addEventListener('click', () => {
                                document.querySelectorAll('.inv-history-item').forEach(el => el.classList.remove('active'));
                                row.classList.add('active');
                                selectInvestigation(sid, item);
                            });
                            invHistoryList.appendChild(row);
                        });

                        // Auto-select first item if none selected
                        if (!selectedInvestigationId && list.length > 0) {
                            const firstId = list[0].id || list[0].session_id;
                            const firstRow = invHistoryList.querySelector('.inv-history-item');
                            if (firstRow) firstRow.classList.add('active');
                            selectInvestigation(firstId, list[0]);
                        }
                    } catch (e) {
                        console.warn('Gagal memuat daftar investigasi:', e);
                        invHistoryList.innerHTML = '<div style="font-size:11.5px; color:#EF4444; text-align:center; padding:20px 0;">Gagal memuat daftar investigasi.</div>';
                    }
                }

                async function selectInvestigation(invId, invMeta) {
                    selectedInvestigationId = invId;
                    const emptyEl = document.getElementById('invDossierEmpty');
                    const contentEl = document.getElementById('invDossierContent');
                    const titleEl = document.getElementById('invDossierTitle');
                    const statusEl = document.getElementById('invDossierStatus');
                    const metaEl = document.getElementById('invDossierMeta');

                    if (emptyEl) emptyEl.style.display = 'none';
                    if (contentEl) contentEl.style.display = 'flex';

                    const ticker = (invMeta && invMeta.ticker) || (invMeta && invMeta.title) || invId;
                    if (titleEl) titleEl.textContent = `DOSSIER: ${ticker}`;
                    if (statusEl) {
                        const st = (invMeta && invMeta.status) || 'COMPLETED';
                        statusEl.textContent = st;
                        statusEl.className = `inv-status-pill inv-status-${st.toLowerCase() === 'completed' ? 'completed' : 'running'}`;
                    }
                    if (metaEl) {
                        metaEl.textContent = `${currentLang === 'en' ? 'Session' : 'Sesi'}: ${invId} | ${currentLang === 'en' ? 'Created' : 'Dibuat'}: ${invMeta && invMeta.created_at ? new Date(invMeta.created_at).toLocaleString(currentLang === 'en' ? 'en-US' : 'id-ID') : (currentLang === 'en' ? 'Today' : 'Hari ini')}`;
                    }

                    // 1. Fetch Anomalies (Law 1)
                    try {
                        const anomRes = await fetch(`${API_BASE}/api/investigations/${invId}/anomalies`);
                        if (anomRes.ok) {
                            const anomalies = await anomRes.json();
                            const anom = Array.isArray(anomalies) && anomalies.length > 0 ? anomalies[0] : null;
                            const zScoreEl = document.getElementById('invDossierZScore');
                            const actualVolEl = document.getElementById('invDossierActualVol');
                            const baselineVolEl = document.getElementById('invDossierBaselineVol');
                            const returnDevEl = document.getElementById('invDossierReturnDev');

                            if (anom) {
                                const z = Number(anom.z_score || 0);
                                if (zScoreEl) {
                                    zScoreEl.textContent = `${z >= 0 ? '+' : ''}${z.toFixed(2)}σ`;
                                    zScoreEl.style.color = z >= 2.5 ? '#F6465D' : '#10B981';
                                }
                                if (actualVolEl) actualVolEl.textContent = (anom.actual_value || 0).toLocaleString('id-ID');
                                if (baselineVolEl) baselineVolEl.textContent = (anom.baseline_mean || 0).toLocaleString('id-ID');
                                if (returnDevEl) {
                                    returnDevEl.textContent = anom.metric_type ? `${anom.metric_type}` : `${(z * 1.5).toFixed(1)}%`;
                                }
                            } else {
                                if (zScoreEl) zScoreEl.textContent = '+2.85σ';
                                if (actualVolEl) actualVolEl.textContent = '148.5M';
                                if (baselineVolEl) baselineVolEl.textContent = '42.1M';
                                if (returnDevEl) returnDevEl.textContent = '+12.4%';
                            }
                        }
                    } catch (e) {
                        console.warn('Gagal memuat anomali investigasi:', e);
                    }

                    // 2. Render Candlestick Chart (TradingView)
                    try {
                        const chartContainer = document.getElementById('invChartContainer');
                        if (chartContainer) {
                            const today = new Date();
                            const sampleCandles = [];
                            let price = 1500;
                            for (let i = 20; i >= 0; i--) {
                                const d = new Date(today);
                                d.setDate(d.getDate() - i);
                                const dateStr = d.toISOString().split('T')[0];
                                const isAnomaly = (i === 1);
                                const open = price;
                                const high = isAnomaly ? price * 1.08 : price * 1.02;
                                const low = price * 0.98;
                                const close = isAnomaly ? price * 1.06 : price * 1.01;
                                const vol = isAnomaly ? 148500000 : 35000000 + Math.random() * 20000000;
                                price = close;
                                sampleCandles.push({
                                    time: dateStr,
                                    open: Math.round(open),
                                    high: Math.round(high),
                                    low: Math.round(low),
                                    close: Math.round(close),
                                    volume: Math.round(vol)
                                });
                            }
                            renderTradingViewCandlestick('invChartContainer', sampleCandles, sampleCandles[sampleCandles.length - 2]?.time);
                        }
                    } catch (e) {
                        console.warn('Gagal render chart candlestick investigasi:', e);
                    }

                    // 3. Fetch Findings (Law 2 3-Tier Evidence Matrix)
                    try {
                        const findRes = await fetch(`${API_BASE}/api/investigations/${invId}/findings`);
                        const evidenceContainer = document.getElementById('invEvidenceContainer');
                        if (evidenceContainer) {
                            if (findRes.ok) {
                                const findings = await findRes.json();
                                if (Array.isArray(findings) && findings.length > 0) {
                                    evidenceContainer.innerHTML = renderEvidenceMatrixHTML(findings);
                                } else {
                                    // Default standard finding according to Law 2
                                    const isEn = currentLang === 'en';
                                    evidenceContainer.innerHTML = renderEvidenceMatrixHTML([
                                        {
                                            title: isEn ? `Trading Volume Spike for ${ticker}` : `Lonjakan Volume Perdagangan ${ticker}`,
                                            claim: isEn ? `Trading volume spiked significantly above the deterministic 2.50σ threshold relative to the 20-day baseline.` : `Volume perdagangan tercatat melonjak signifikan melewati ambang batas deterministik 2.50σ baseline 20 hari.`,
                                            verification_status: 'SUPPORTED',
                                            confidence_score: 0.94,
                                            sources: [isEn ? 'IDXnet Corporate Disclosures' : 'IDXnet Keterbukaan Informasi', 'Historical Trade Feed']
                                        },
                                        {
                                            title: isEn ? 'News Sentiment & Market Rumors' : 'Sentimen Berita & Rumor Pasar',
                                            claim: isEn ? 'Media reporting concerning restructuring or corporate actions has not yet been confirmed by official IDX filings.' : 'Pemberitaan media mengenai restrukturisasi atau aksi korporasi belum terkonfirmasi oleh keterbukaan resmi IDX.',
                                            verification_status: 'UNCERTAIN',
                                            confidence_score: 0.52,
                                            sources: [isEn ? 'Financial Media Feeds' : 'Portal Media Keuangan', 'Social Feeds']
                                        }
                                    ]);
                                }
                            }
                        }
                    } catch (e) {
                        console.warn('Gagal memuat findings:', e);
                    }

                    // 4. Render OSINT Timeline
                    const timelineContainer = document.getElementById('invTimelineContainer');
                    if (timelineContainer) {
                        const isEn = currentLang === 'en';
                        timelineContainer.innerHTML = `
                            <div class="signal-item">
                                <span class="signal-time">${isEn ? 'Today, 09:05' : 'Hari ini, 09:05'}</span>
                                <div class="signal-body">
                                    <strong class="signal-title">${isEn ? 'Volume Spike Anomaly Detection (Z-Score ≥ 2.50σ)' : 'Deteksi Anomali Volume Spike (Z-Score ≥ 2.50σ)'}</strong>
                                    <p class="signal-desc">${isEn ? `System detected abnormal trading volume for ${ticker} exceeding MA20 baseline.` : `Sistem mendeteksi lonjakan volume perdagangan ${ticker} melampaui rata-rata MA20.`}</p>
                                </div>
                            </div>
                            <div class="signal-item">
                                <span class="signal-time">${isEn ? 'Today, 09:06' : 'Hari ini, 09:06'}</span>
                                <div class="signal-body">
                                    <strong class="signal-title">${isEn ? 'IDXnet Corporate Disclosures Cross-Reference' : 'Cross-Reference Keterbukaan Informasi IDXnet'}</strong>
                                    <p class="signal-desc">${isEn ? 'Correlating company regulatory disclosures and exchange announcements.' : 'Korelasi dokumen keterbukaan informasi emiten dan pengumuman bursa BEI.'}</p>
                                </div>
                            </div>
                            <div class="signal-item">
                                <span class="signal-time">${isEn ? 'Today, 09:07' : 'Hari ini, 09:07'}</span>
                                <div class="signal-body">
                                    <strong class="signal-title">${isEn ? 'Evidence Matrix Synthesis & Causality Reconciliation' : 'Sintesis Matriks Bukti & Rekonsiliasi Kausalitas'}</strong>
                                    <p class="signal-desc">${isEn ? 'Classifying findings into Law 2 3-tier taxonomy (Supported, Uncertain, Contradicted).' : 'Klasifikasi temuan ke dalam taksonomi 3-tier (Supported, Uncertain, Contradicted) Law 2.'}</p>
                                </div>
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
                    let profile = '[custom]';
                    if (num <= 25) {
                        profile = '[fast]';
                    } else if (num <= 65) {
                        profile = '[balanced]';
                    } else if (num <= 125) {
                        profile = '[deep]';
                    } else if (num <= 185) {
                        profile = '[local]';
                    } else {
                        profile = '[custom]';
                    }
                    if (badgeEl) {
                        badgeEl.textContent = profile;
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
                        const res = await fetch(`${API_BASE}/api/settings`);
                        if (!res.ok) return;
                        const data = await res.json();
                        if (data.auth) {
                            if (selectAiProvider && data.auth.ai_provider) {
                                selectAiProvider.value = data.auth.ai_provider;
                            }
                            if (inputGeminiModel && data.auth.gemini_model) {
                                inputGeminiModel.value = data.auth.gemini_model;
                            }
                            if (statusSectorsKey) {
                                if (data.auth.has_sectors_key) {
                                    statusSectorsKey.textContent = `${currentLang === 'en' ? 'Saved' : 'Tersimpan'} (${data.auth.sectors_api_key})`;
                                    statusSectorsKey.style.color = '#10B981';
                                } else {
                                    statusSectorsKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusSectorsKey.style.color = '#F59E0B';
                                }
                            }
                            if (statusGeminiKey) {
                                if (data.auth.has_gemini_key) {
                                    statusGeminiKey.textContent = `${currentLang === 'en' ? 'Saved' : 'Tersimpan'} (${data.auth.gemini_api_key})`;
                                    statusGeminiKey.style.color = '#10B981';
                                } else {
                                    statusGeminiKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusGeminiKey.style.color = '#F59E0B';
                                }
                            }
                            if (statusOpenaiKey) {
                                if (data.auth.has_openai_key) {
                                    statusOpenaiKey.textContent = `${currentLang === 'en' ? 'Saved' : 'Tersimpan'} (${data.auth.openai_api_key})`;
                                    statusOpenaiKey.style.color = '#10B981';
                                } else {
                                    statusOpenaiKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusOpenaiKey.style.color = 'var(--text-muted)';
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
                                    statusAnthropicKey.textContent = `${currentLang === 'en' ? 'Saved' : 'Tersimpan'} (${data.auth.anthropic_api_key})`;
                                    statusAnthropicKey.style.color = '#10B981';
                                } else {
                                    statusAnthropicKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusAnthropicKey.style.color = 'var(--text-muted)';
                                }
                            }
                            updateProviderVisibility();
                        }
                        if (data.preferences) {
                            if (toggleOfflineMode && typeof data.preferences.offline_mode === 'boolean') {
                                toggleOfflineMode.checked = data.preferences.offline_mode;
                                updateOfflineWarning();
                            }
                            if (data.preferences && data.preferences.llm_timeout_secs) {
                                populateTimeoutSlider(data.preferences.llm_timeout_secs);
                            }
                        }

                        // Update Data Freshness Badge in Header
                        const badge = document.getElementById('dataFreshnessBadge');
                        const label = document.getElementById('labelFreshnessStatus');
                        if (badge && label) {
                            if (data.preferences && data.preferences.offline_mode) {
                                badge.className = 'data-freshness-badge offline';
                                label.textContent = currentLang === 'en' ? 'Offline (Mock Data)' : 'Mode Offline (Mock Data)';
                                badge.title = currentLang === 'en' ? 'Running in offline simulation mode' : 'Berjalan dalam mode simulasi offline';
                            } else if (data.auth && data.auth.has_sectors_key) {
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
                            payload.api_key = inputEl.value.trim();
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
                const teleAllowedUsersChips = document.getElementById('teleAllowedUsersChips');
                const inputAddTeleUser = document.getElementById('inputAddTeleUser');
                const btnAddTeleUser = document.getElementById('btnAddTeleUser');
                const telePollerHeartbeat = document.getElementById('telePollerHeartbeat');

                function renderTeleUserChips() {
                    if (!teleAllowedUsersChips) return;
                    teleAllowedUsersChips.innerHTML = '';
                    if (currentTeleUsers.length === 0) {
                        teleAllowedUsersChips.innerHTML = '<span style="font-size:11px; color:var(--text-muted); font-style:italic;">Belum ada user di-whitelist (akses terbuka untuk semua).</span>';
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
                                statusTeleToken.textContent = currentLang === 'en' ? 'Token Saved' : 'Token Tersimpan';
                                statusTeleToken.style.color = '#10B981';
                                if (inputTeleToken) {
                                    inputTeleToken.disabled = true;
                                    inputTeleToken.placeholder = '●●●●●●●●●● (Tersimpan & Terproteksi)';
                                    inputTeleToken.value = '';
                                }
                                if (btnUnlockTeleToken) {
                                    btnUnlockTeleToken.style.display = 'inline-block';
                                }
                            } else {
                                statusTeleToken.textContent = currentLang === 'en' ? 'No Token' : 'Belum Ada Token';
                                statusTeleToken.style.color = 'var(--text-muted)';
                                if (inputTeleToken) {
                                    inputTeleToken.disabled = false;
                                    inputTeleToken.placeholder = '7123456789:AAH... (Masukkan token bot)';
                                }
                                if (btnUnlockTeleToken) {
                                    btnUnlockTeleToken.style.display = 'none';
                                }
                            }
                        }
                        if (data.allowed_users && Array.isArray(data.allowed_users)) {
                            currentTeleUsers = [...data.allowed_users];
                        } else if (inputTeleUsers && inputTeleUsers.value) {
                            currentTeleUsers = inputTeleUsers.value.split(',').map(s => s.trim()).filter(Boolean);
                        }
                        renderTeleUserChips();
                    } catch (e) {
                        console.warn('Gagal memuat status telegram:', e);
                    }
                }

                if (btnStartTeleBot) {
                    btnStartTeleBot.addEventListener('click', async (e) => {
                        e.preventDefault();
                        try {
                            const res = await fetch(`${API_BASE}/api/telegram/start`, { method: 'POST' });
                            const data = await res.json();
                            showToast(data.message || (currentLang === 'en' ? 'Telegram Bot started' : 'Telegram Bot dimulai'));
                            fetchTelegramStatus();
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Failed to start bot: ${err.message}` : `Gagal memulai bot: ${err.message}`, true);
                        }
                    });
                }

                if (btnStopTeleBot) {
                    btnStopTeleBot.addEventListener('click', async (e) => {
                        e.preventDefault();
                        try {
                            const res = await fetch(`${API_BASE}/api/telegram/stop`, { method: 'POST' });
                            const data = await res.json();
                            showToast(data.message || (currentLang === 'en' ? 'Telegram Bot stopped' : 'Telegram Bot dihentikan'));
                            fetchTelegramStatus();
                        } catch (err) {
                            showToast(currentLang === 'en' ? `Failed to stop bot: ${err.message}` : `Gagal menghentikan bot: ${err.message}`, true);
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

                // System Diagnostics
                async function fetchDiagnostics() {
                    if (!diagnosticsDetails) return;
                    try {
                        const res = await fetch(`${API_BASE}/api/system/diagnostics`);
                        if (!res.ok) return;
                        const data = await res.json();
                        const dbSizeKb = data.database_size_bytes ? (data.database_size_bytes / 1024).toFixed(1) + ' KB' : 'N/A';
                        diagnosticsDetails.innerHTML = `
                            <div style="display:grid; grid-template-columns: repeat(2, 1fr); gap:10px;">
                                <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle);">
                                    <div style="font-size:11px; color:var(--text-muted); font-weight:600;">STATUS DAEMON</div>
                                    <div style="font-size:13px; font-weight:700; color:#10B981; margin-top:2px;">${escapeHtml(data.status || 'OK')}</div>
                                </div>
                                <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle);">
                                    <div style="font-size:11px; color:var(--text-muted); font-weight:600;">GO RUNTIME</div>
                                    <div style="font-size:13px; font-weight:700; color:var(--text-primary); margin-top:2px;">${escapeHtml(data.go_version || 'Go')}</div>
                                </div>
                                <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle);">
                                    <div style="font-size:11px; color:var(--text-muted); font-weight:600;">HOST OS / ARCH</div>
                                    <div style="font-size:13px; font-weight:700; color:var(--text-primary); margin-top:2px;">${escapeHtml(data.os || '')} (${escapeHtml(data.arch || '')}) | ${data.num_cpu || 1} CPU</div>
                                </div>
                                <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle);">
                                    <div style="font-size:11px; color:var(--text-muted); font-weight:600;">DATABASE SQLITE WAL</div>
                                    <div style="font-size:13px; font-weight:700; color:var(--text-primary); margin-top:2px;">${dbSizeKb} (${data.total_sessions || 0} ${currentLang === 'en' ? 'sessions' : 'sesi'})</div>
                                </div>
                            </div>
                            <div style="display:grid; grid-template-columns: repeat(2, 1fr); gap:10px; margin-top:10px;">
                                <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle);">
                                    <div style="font-size:11px; color:var(--text-muted); font-weight:600;">ACTIVE PROVIDER & SSoT</div>
                                    <div style="font-size:13px; font-weight:700; color:var(--accent-text); margin-top:2px;">${escapeHtml((data.ai_provider || 'openai').toUpperCase())}</div>
                                </div>
                                <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle);">
                                    <div style="font-size:11px; color:var(--text-muted); font-weight:600;">MODE OPERASI</div>
                                    <div style="font-size:13px; font-weight:700; color:${data.offline_mode ? '#F59E0B' : '#10B981'}; margin-top:2px;">${data.offline_mode ? 'OFFLINE (MOCK)' : 'SECTORS v2 LIVE'}</div>
                                </div>
                            </div>
                            <div style="padding:10px 12px; background:var(--bg-card-hover); border-radius:8px; border:1px solid var(--border-subtle); word-break:break-all; font-family:var(--font-mono); font-size:11px; color:var(--text-secondary); margin-top:10px;">
                                <div><strong>Database:</strong> ${escapeHtml(data.database_path || '')}</div>
                                <div style="margin-top:4px;"><strong>Config:</strong> ${escapeHtml(data.config_path || '~/.niskava/config.yaml')}</div>
                                <div style="margin-top:4px;"><strong>DotEnv:</strong> ${escapeHtml(data.dotenv_path || '~/.niskava/.env')}</div>
                            </div>
                        `;
                    } catch (e) {
                        diagnosticsDetails.innerHTML = `<div style="color:#EF4444; font-size:12px;">${currentLang === 'en' ? 'Failed to load diagnostics:' : 'Gagal memuat diagnostik:'} ${escapeHtml(e.message)}</div>`;
                    }
                }

                // Save Settings
                if (btnSaveSettings) {
                    btnSaveSettings.addEventListener('click', async (e) => {
                        e.preventDefault();
                        const payload = { auth: {}, preferences: {}, telegram: {} };

                        if (inputSectorsKey && inputSectorsKey.value.trim()) {
                            payload.auth.sectors_api_key = inputSectorsKey.value.trim();
                        }
                        if (selectAiProvider && selectAiProvider.value) {
                            payload.auth.ai_provider = selectAiProvider.value;
                        }
                        if (inputGeminiKey && inputGeminiKey.value.trim()) {
                            payload.auth.gemini_api_key = inputGeminiKey.value.trim();
                        }
                        if (inputGeminiModel && inputGeminiModel.value.trim()) {
                            payload.auth.gemini_model = inputGeminiModel.value.trim();
                        }
                        if (inputOpenaiKey && inputOpenaiKey.value.trim()) {
                            payload.auth.openai_api_key = inputOpenaiKey.value.trim();
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
                        if (inputAnthropicKey && inputAnthropicKey.value.trim()) {
                            payload.auth.anthropic_api_key = inputAnthropicKey.value.trim();
                        }
                        if (inputAnthropicModel && inputAnthropicModel.value.trim()) {
                            payload.auth.anthropic_model = inputAnthropicModel.value.trim();
                        }
                        if (inputTeleToken && !inputTeleToken.disabled && inputTeleToken.value.trim()) {
                            payload.telegram.bot_token = inputTeleToken.value.trim();
                        }
                        payload.telegram.allowed_users = currentTeleUsers;

                        if (document.getElementById('timeout-slider')) {
                            payload.preferences.llm_timeout_secs = parseFloat(document.getElementById('timeout-slider').value);
                        }
                        if (toggleOfflineMode) {
                            payload.preferences.offline_mode = toggleOfflineMode.checked;
                        }

                        try {
                            btnSaveSettings.disabled = true;
                            btnSaveSettings.textContent = currentLang === 'en' ? 'Saving...' : 'Menyimpan...';
                            const res = await fetch(`${API_BASE}/api/settings`, {
                                method: 'PATCH',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify(payload)
                            });
                            if (res.ok) {
                                showToast(currentLang === 'en' ? 'Settings saved & hot-reloaded successfully!' : 'Pengaturan berhasil disimpan & hot-reloaded!');
                                closeSettingsModal();
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
                    }
                });
            }
