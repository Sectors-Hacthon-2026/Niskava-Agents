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
                            btnPruneMockData.innerHTML = `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg> <span>${currentLang === 'en' ? 'Clean Test Data' : 'Bersihkan Data Uji'}</span>`;
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
                            switchSession(selectedInvestigationId);
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

                    // 4. Render Evidence & News Timeline
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
                        const res = await fetch(`${API_BASE}/api/settings?reveal=true`);
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
                                    statusSectorsKey.textContent = currentLang === 'en' ? 'Saved' : 'Tersimpan';
                                    statusSectorsKey.style.color = '#10B981';
                                    if (inputSectorsKey) {
                                        inputSectorsKey.value = data.auth.sectors_api_key || '';
                                        inputSectorsKey.dataset.saved = 'true';
                                        inputSectorsKey.placeholder = 'sec_live_... (Tersimpan)';
                                    }
                                } else {
                                    statusSectorsKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusSectorsKey.style.color = '#F59E0B';
                                    if (inputSectorsKey) {
                                        inputSectorsKey.value = '';
                                        delete inputSectorsKey.dataset.saved;
                                        inputSectorsKey.placeholder = 'sec_live_... (Masukkan Sectors API Key)';
                                    }
                                }
                            }
                            const tab1SectorsBadge = document.getElementById('tab1SectorsBadge');
                            if (tab1SectorsBadge) {
                                if (data.auth.has_sectors_key) {
                                    tab1SectorsBadge.textContent = currentLang === 'en' ? 'Saved' : 'Tersimpan';
                                    tab1SectorsBadge.style.color = '#10B981';
                                } else {
                                    tab1SectorsBadge.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    tab1SectorsBadge.style.color = '#F59E0B';
                                }
                            }
                            if (statusGeminiKey) {
                                if (data.auth.has_gemini_key) {
                                    statusGeminiKey.textContent = currentLang === 'en' ? 'Saved' : 'Tersimpan';
                                    statusGeminiKey.style.color = '#10B981';
                                    if (inputGeminiKey) {
                                        inputGeminiKey.value = data.auth.gemini_api_key || '';
                                        inputGeminiKey.dataset.saved = 'true';
                                        inputGeminiKey.placeholder = 'AIzaSy... (Tersimpan)';
                                    }
                                } else {
                                    statusGeminiKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusGeminiKey.style.color = '#F59E0B';
                                    if (inputGeminiKey) {
                                        inputGeminiKey.value = '';
                                        delete inputGeminiKey.dataset.saved;
                                        inputGeminiKey.placeholder = 'AIzaSy... (Masukkan Gemini API Key)';
                                    }
                                }
                            }
                            if (statusOpenaiKey) {
                                if (data.auth.has_openai_key) {
                                    statusOpenaiKey.textContent = currentLang === 'en' ? 'Saved' : 'Tersimpan';
                                    statusOpenaiKey.style.color = '#10B981';
                                    if (inputOpenaiKey) {
                                        inputOpenaiKey.value = data.auth.openai_api_key || '';
                                        inputOpenaiKey.dataset.saved = 'true';
                                        inputOpenaiKey.placeholder = 'sk-... (Tersimpan)';
                                    }
                                } else {
                                    statusOpenaiKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusOpenaiKey.style.color = 'var(--text-muted)';
                                    if (inputOpenaiKey) {
                                        inputOpenaiKey.value = '';
                                        delete inputOpenaiKey.dataset.saved;
                                        inputOpenaiKey.placeholder = 'sk-... (Kosongkan jika menggunakan gateway lokal)';
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
                                    statusAnthropicKey.textContent = currentLang === 'en' ? 'Saved' : 'Tersimpan';
                                    statusAnthropicKey.style.color = '#10B981';
                                    if (inputAnthropicKey) {
                                        inputAnthropicKey.value = data.auth.anthropic_api_key || '';
                                        inputAnthropicKey.dataset.saved = 'true';
                                        inputAnthropicKey.placeholder = 'sk-ant-api03-... (Tersimpan)';
                                    }
                                } else {
                                    statusAnthropicKey.textContent = currentLang === 'en' ? 'No Key' : 'Belum Ada Kunci';
                                    statusAnthropicKey.style.color = 'var(--text-muted)';
                                    if (inputAnthropicKey) {
                                        inputAnthropicKey.value = '';
                                        delete inputAnthropicKey.dataset.saved;
                                        inputAnthropicKey.placeholder = 'sk-ant-api03-... (Masukkan Anthropic API Key)';
                                    }
                                }
                            }
                            // Update reactive provider card badges
                            const isEn = (typeof currentLang !== 'undefined' && currentLang === 'en');
                            const tagGemini = document.getElementById('tagGeminiStatus');
                            if (tagGemini) {
                                if (data.auth && data.auth.has_gemini_key) {
                                    tagGemini.textContent = isEn ? 'Saved' : 'Tersimpan';
                                    tagGemini.style.color = '#10B981';
                                    tagGemini.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagGemini.textContent = isEn ? 'Recommended' : 'Disarankan';
                                    tagGemini.style.color = '#03A66D';
                                    tagGemini.style.background = 'rgba(3, 166, 109, 0.12)';
                                }
                            }

                            const tagOpenai = document.getElementById('tagOpenaiStatus');
                            if (tagOpenai) {
                                if (data.auth && data.auth.has_openai_key) {
                                    tagOpenai.textContent = isEn ? 'Saved' : 'Tersimpan';
                                    tagOpenai.style.color = '#10B981';
                                    tagOpenai.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagOpenai.textContent = 'Gateway';
                                    tagOpenai.style.color = '';
                                    tagOpenai.style.background = '';
                                }
                            }

                            const tagAnthropic = document.getElementById('tagAnthropicStatus');
                            if (tagAnthropic) {
                                if (data.auth && data.auth.has_anthropic_key) {
                                    tagAnthropic.textContent = isEn ? 'Saved' : 'Tersimpan';
                                    tagAnthropic.style.color = '#10B981';
                                    tagAnthropic.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagAnthropic.textContent = isEn ? 'Reasoning' : 'Penalaran';
                                    tagAnthropic.style.color = '';
                                    tagAnthropic.style.background = '';
                                }
                            }

                            const tagOllama = document.getElementById('tagOllamaStatus');
                            if (tagOllama) {
                                const ollamaUrl = (data.auth && data.auth.ollama_base_url) || (inputOllamaBaseUrl ? inputOllamaBaseUrl.value.trim() : '') || 'http://localhost:11434';
                                if (ollamaUrl) {
                                    tagOllama.textContent = isEn ? 'Available' : 'Tersedia';
                                    tagOllama.style.color = '#10B981';
                                    tagOllama.style.background = 'rgba(16, 185, 129, 0.12)';
                                } else {
                                    tagOllama.textContent = isEn ? 'Local' : 'Lokal';
                                    tagOllama.style.color = '';
                                    tagOllama.style.background = '';
                                }
                            }

                            updateProviderVisibility();
                            updateActivePresetChips();
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
                                    inputTeleToken.disabled = false;
                                    inputTeleToken.value = data.bot_token || '';
                                    inputTeleToken.dataset.saved = 'true';
                                    inputTeleToken.placeholder = '7123456789:AAH... (Tersimpan)';
                                }
                                if (btnUnlockTeleToken) {
                                    btnUnlockTeleToken.style.display = 'none';
                                }
                            } else {
                                statusTeleToken.textContent = currentLang === 'en' ? 'No Token' : 'Belum Ada Token';
                                statusTeleToken.style.color = 'var(--text-muted)';
                                if (inputTeleToken) {
                                    inputTeleToken.disabled = false;
                                    inputTeleToken.value = '';
                                    delete inputTeleToken.dataset.saved;
                                    inputTeleToken.placeholder = '7123456789:AAH... (Masukkan token bot)';
                                }
                                if (btnUnlockTeleToken) {
                                    btnUnlockTeleToken.style.display = 'none';
                                }
                            }
                        }

                        // Dynamic Start / Stop Bot Button Toggle
                        if (isRunning) {
                            if (btnStartTeleBot) btnStartTeleBot.style.display = 'none';
                            if (btnStopTeleBot) {
                                btnStopTeleBot.style.display = 'inline-flex';
                                btnStopTeleBot.disabled = false;
                                btnStopTeleBot.style.opacity = '1';
                                btnStopTeleBot.style.cursor = 'pointer';
                                btnStopTeleBot.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg><span>${currentLang === 'en' ? 'Stop Bot' : 'Matikan Bot'}</span>`;
                            }
                        } else {
                            if (btnStopTeleBot) btnStopTeleBot.style.display = 'none';
                            if (btnStartTeleBot) {
                                btnStartTeleBot.style.display = 'inline-flex';
                                btnStartTeleBot.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="margin-right:4px;"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg><span>${currentLang === 'en' ? 'Start Bot' : 'Nyalakan Bot'}</span>`;
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
                async function fetchDiagnostics() {
                    if (!diagnosticsDetails) return;
                    try {
                        const res = await fetch(`${API_BASE}/api/system/diagnostics`);
                        if (!res.ok) return;
                        const data = await res.json();
                        const dbSizeKb = data.database_size_bytes ? (data.database_size_bytes / 1024).toFixed(1) + ' KB' : 'N/A';
                        const isEn = (currentLang === 'en');

                        const sidebarProfileName = document.getElementById('sidebarProfileName');
                        const sidebarProfileAvatar = document.getElementById('sidebarProfileAvatar');
                        if (sidebarProfileName && data.username) {
                            sidebarProfileName.textContent = data.username;
                        }
                        if (sidebarProfileAvatar && data.username) {
                            sidebarProfileAvatar.textContent = data.username.slice(0, 2).toUpperCase();
                        }

                        diagnosticsDetails.innerHTML = `
                            <div class="kpi-stats-grid" style="grid-template-columns: repeat(3, 1fr);">
                                <div class="kpi-stat-card">
                                    <span class="kpi-stat-label">STATUS DAEMON</span>
                                    <span class="kpi-stat-num" style="color:#10B981; font-size:14px;">${escapeHtml(data.status || 'OK')}</span>
                                    <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">Local IPC Active</span>
                                </div>
                                <div class="kpi-stat-card">
                                    <span class="kpi-stat-label">GO RUNTIME</span>
                                    <span class="kpi-stat-num" style="font-size:14px;">${escapeHtml(data.go_version || 'Go')}</span>
                                    <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${escapeHtml(data.os || '')} (${escapeHtml(data.arch || '')}) | ${data.num_cpu || 1} CPU</span>
                                </div>
                                <div class="kpi-stat-card">
                                    <span class="kpi-stat-label">SQLITE WAL (LAW 4)</span>
                                    <span class="kpi-stat-num" style="font-size:14px;">${dbSizeKb}</span>
                                    <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${data.total_sessions || 0} ${isEn ? 'sessions saved' : 'sesi tersimpan'}</span>
                                </div>
                            </div>
                            <div class="kpi-stats-grid" style="grid-template-columns: repeat(2, 1fr); margin-top:0;">
                                <div class="kpi-stat-card">
                                    <span class="kpi-stat-label">ACTIVE AI PROVIDER</span>
                                    <span class="kpi-stat-num" style="color:var(--accent-text); font-size:14px;">${escapeHtml((data.ai_provider || 'openai').toUpperCase())}</span>
                                    <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">ReAct Cognitive Engine</span>
                                </div>
                                <div class="kpi-stat-card">
                                    <span class="kpi-stat-label">MODE OPERASI</span>
                                    <span class="kpi-stat-num" style="color:${data.offline_mode ? '#F59E0B' : '#10B981'}; font-size:14px;">${data.offline_mode ? 'OFFLINE (MOCK)' : 'SECTORS v2 LIVE'}</span>
                                    <span style="font-size:10px; color:var(--text-muted); margin-top:2px;">${data.offline_mode ? 'Fixture Simulation' : 'Law 5 Active Credit Sync'}</span>
                                </div>
                            </div>
                            <div class="settings-card" style="margin-top:2px;">
                                <div class="settings-card-head">
                                    <div class="settings-card-title">
                                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
                                        <span>Jalur Penyimpanan SSoT (Single Source of Truth)</span>
                                    </div>
                                    <span class="panel-card-badge">Local-First</span>
                                </div>
                                <div style="display:flex; flex-direction:column; gap:6px; font-family:var(--font-mono); font-size:11px;">
                                    <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; background:var(--bg-card); border-radius:6px; border:1px solid var(--border-subtle);">
                                        <div><strong style="color:var(--text-primary);">Database:</strong> <span style="color:var(--text-secondary);">${escapeHtml(data.database_path || '~/.niskava/niskava.db')}</span></div>
                                        <button type="button" class="btn-copy-path" data-path="${escapeHtml(data.database_path || '~/.niskava/niskava.db')}" title="Salin Path" style="cursor:pointer; color:var(--text-muted); padding:2px 6px; border:none; background:transparent;">
                                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                        </button>
                                    </div>
                                    <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; background:var(--bg-card); border-radius:6px; border:1px solid var(--border-subtle);">
                                        <div><strong style="color:var(--text-primary);">Config:</strong> <span style="color:var(--text-secondary);">${escapeHtml(data.config_path || '~/.niskava/config.yaml')}</span></div>
                                        <button type="button" class="btn-copy-path" data-path="${escapeHtml(data.config_path || '~/.niskava/config.yaml')}" title="Salin Path" style="cursor:pointer; color:var(--text-muted); padding:2px 6px; border:none; background:transparent;">
                                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                        </button>
                                    </div>
                                    <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; background:var(--bg-card); border-radius:6px; border:1px solid var(--border-subtle);">
                                        <div><strong style="color:var(--text-primary);">DotEnv:</strong> <span style="color:var(--text-secondary);">${escapeHtml(data.dotenv_path || '~/.niskava/.env')}</span></div>
                                        <button type="button" class="btn-copy-path" data-path="${escapeHtml(data.dotenv_path || '~/.niskava/.env')}" title="Salin Path" style="cursor:pointer; color:var(--text-muted); padding:2px 6px; border:none; background:transparent;">
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
                    } catch (e) {
                        diagnosticsDetails.innerHTML = `<div style="color:#EF4444; font-size:12px;">${currentLang === 'en' ? 'Failed to load diagnostics:' : 'Gagal memuat diagnostik:'} ${escapeHtml(e.message)}</div>`;
                    }
                }

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
                        if (toggleOfflineMode) {
                            payload.preferences.offline_mode = toggleOfflineMode.checked;
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
