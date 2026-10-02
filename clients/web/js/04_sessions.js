// Pin / Unpin Session Handler
            async function togglePinSession(sessionId, currentPinned) {
                try {
                    const res = await fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(sessionId)}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ is_pinned: !currentPinned })
                    });
                    if (res.ok) {
                        showToast(!currentPinned ? (currentLang === 'en' ? 'Session pinned to top' : 'Sesi disematkan ke atas (Pinned)') : (currentLang === 'en' ? 'Session unpinned' : 'Sematkan sesi dibatalkan'));
                        loadChatSessions();
                    }
                } catch (e) {
                    showToast(currentLang === 'en' ? 'Failed to update pin status' : 'Gagal mengubah status pin', true);
                }
            }

            // Rename Session Handler
            async function renameSession(sessionId, oldTitle) {
                const newTitle = prompt(t('rename_prompt'), oldTitle);
                if (!newTitle || newTitle.trim() === oldTitle.trim()) return;

                try {
                    const res = await fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(sessionId)}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ title: newTitle.trim() })
                    });
                    if (res.ok) {
                        showToast(currentLang === 'en' ? 'Conversation title updated' : 'Judul percakapan berhasil diperbarui');
                        if (currentSessionId === sessionId) {
                            document.getElementById('currentSessionLabel').textContent = newTitle.trim();
                        }
                        loadChatSessions();
                    }
                } catch (e) {
                    showToast(currentLang === 'en' ? 'Failed to rename title' : 'Gagal mengganti judul', true);
                }
            }

            // Cached sessions for search restore
            let cachedChatSessions = [];

            // Dynamic Real Chat History & Grouping with Pin & Rename Support
            async function loadChatSessions(restoreActive = true) {
                try {
                    let res = await fetch(`${API_BASE}/api/chat/sessions`);
                    if (!res.ok) {
                        res = await fetch(`${API_BASE}/api/sessions`);
                    }
                    if (!res.ok) return;
                    const data = await res.json();
                    cachedChatSessions = data.sessions || data.chat_sessions || data.data || [];
                    renderChatHistoryGroups(cachedChatSessions);

                    if (restoreActive) {
                        const savedActiveId = localStorage.getItem('niskava_active_session');
                        if (savedActiveId) {
                            const found = cachedChatSessions.find(s => (s.id || s.session_id) === savedActiveId);
                            if (found) {
                                switchSession(savedActiveId, formatSessionTitle(found));
                            }
                        }
                    }
                } catch(e) {
                    console.error('Failed to load chat sessions', e);
                }
            }

            function formatSessionTitle(s) {
                let title = (s.title || '').trim();
                // Treat generic or timestamp-based Telegram titles as placeholders.
                // Telegram session titles follow "Telegram (@user) · 30 Sep 15:04:05".
                // Once the first message arrives, UpdateChatSession sets a topic-based title,
                // but last_message_preview is a shorter fallback to display in the sidebar.
                const isGenericTitle = !title ||
                    title === 'Sesi Riset Pasar' ||
                    title.startsWith('CHAT-') ||
                    title.startsWith('WEB-') ||
                    title.startsWith('TELE-') ||
                    /^Telegram\s*(\(@[^)]+\))?\s*·/.test(title);
                if (isGenericTitle) {
                    title = (s.last_message_preview || s.first_message || '').trim();
                }
                if (!title) {
                    title = s.id || s.session_id || 'Percakapan Riset';
                }
                title = title.replace(/^#+\s*/, '').replace(/^[•\-\*]\s*/, '').replace(/[\r\n]+/g, ' ').trim();
                return title;
            }

            let activeHistoryDropdown = null;

            function closeHistoryDropdown() {
                if (activeHistoryDropdown) {
                    if (activeHistoryDropdown.btn) {
                        activeHistoryDropdown.btn.classList.remove('active');
                    }
                    if (activeHistoryDropdown.menu && activeHistoryDropdown.menu.parentNode) {
                        activeHistoryDropdown.menu.parentNode.removeChild(activeHistoryDropdown.menu);
                    }
                    activeHistoryDropdown = null;
                }
            }

            function toggleHistoryDropdown(e, sId, fullTitle, isPinned, moreBtn) {
                if (activeHistoryDropdown && activeHistoryDropdown.sessionId === sId) {
                    closeHistoryDropdown();
                    return;
                }
                closeHistoryDropdown();

                moreBtn.classList.add('active');
                const menu = document.createElement('div');
                menu.className = 'history-dropdown-menu';
                menu.innerHTML = `
                    <button class="history-dropdown-item" data-action="pin">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="${isPinned ? 'currentColor' : 'none'}" stroke="currentColor" stroke-width="2"><line x1="12" y1="17" x2="12" y2="22"></line><path d="M5 17h14v-2l-2-3V6a1 1 0 0 0-1-1H8a1 1 0 0 0-1 1v6l-2 3v2z"></path></svg>
                        <span>${isPinned ? t('menu_unpin') : t('menu_pin')}</span>
                    </button>
                    <button class="history-dropdown-item" data-action="rename">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
                        <span>${t('menu_rename')}</span>
                    </button>
                    <button class="history-dropdown-item danger" data-action="delete">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                        <span>${t('menu_delete')}</span>
                    </button>
                `;

                document.body.appendChild(menu);

                // Smart positioning based on viewport
                const rect = moreBtn.getBoundingClientRect();
                const menuWidth = 165;
                const menuHeight = 118;
                let top = rect.bottom + 4;
                let left = rect.right - menuWidth;

                if (top + menuHeight > window.innerHeight) {
                    top = rect.top - menuHeight - 4;
                }
                if (left < 10) left = 10;

                menu.style.top = `${top}px`;
                menu.style.left = `${left}px`;

                activeHistoryDropdown = { sessionId: sId, menu, btn: moreBtn };

                menu.querySelector('[data-action="pin"]').addEventListener('click', async (evt) => {
                    evt.stopPropagation();
                    closeHistoryDropdown();
                    await togglePinSession(sId, isPinned);
                });

                menu.querySelector('[data-action="rename"]').addEventListener('click', async (evt) => {
                    evt.stopPropagation();
                    closeHistoryDropdown();
                    await renameSession(sId, fullTitle);
                });

                menu.querySelector('[data-action="delete"]').addEventListener('click', async (evt) => {
                    evt.stopPropagation();
                    closeHistoryDropdown();
                    await deleteSession(sId);
                });
            }

            // Global listeners to auto-close dropdown on outside click, resize, or list scroll
            document.addEventListener('click', (e) => {
                if (activeHistoryDropdown && !e.target.closest('.history-dropdown-menu') && !e.target.closest('.history-more-btn')) {
                    closeHistoryDropdown();
                }
            });
            window.addEventListener('resize', closeHistoryDropdown);
            const historyScrollContainer = document.getElementById('historyList');
            if (historyScrollContainer) {
                historyScrollContainer.addEventListener('scroll', closeHistoryDropdown, { passive: true });
            }

            /**
             * Returns a small inline badge HTML string indicating the session's platform origin.
             * Telegram sessions carry a TELE- ID prefix per the DB convention.
             * Returns empty string for Web sessions.
             * @param {string} sessionId
             * @returns {string}
             */
            function getSessionPlatformBadge(sessionId) {
                if (sessionId && sessionId.startsWith('TELE-')) {
                    return `<span title="Sesi dari Telegram Bot" style="font-size:10px; background:rgba(39,174,245,0.15); color:#27aef5; border-radius:4px; padding:1px 5px; margin-right:4px; vertical-align:middle; flex-shrink:0; line-height:1.6; display:inline-block; font-weight:600;">TG</span>`;
                }
                return '';
            }

            function renderChatHistoryGroups(sessions) {
                closeHistoryDropdown();
                historyList.innerHTML = '';
                const historyCountEl = document.getElementById('historyCount');
                if (historyCountEl) historyCountEl.textContent = sessions.length;

                if (!sessions || sessions.length === 0) {
                    historyList.innerHTML = `<div class="history-empty-notice">${t('history_empty')}</div>`;
                    return;
                }

                // Sort: pinned sessions at top, then recently updated
                const sorted = [...sessions].sort((a, b) => {
                    const pinA = !!a.is_pinned;
                    const pinB = !!b.is_pinned;
                    if (pinA && !pinB) return -1;
                    if (!pinA && pinB) return 1;
                    return 0;
                });

                sorted.forEach(s => {
                    const sId = s.id || s.session_id;
                    const fullTitle = formatSessionTitle(s);
                    const isPinned = !!s.is_pinned;
                    let displayTitle = fullTitle;
                    if (displayTitle.length > 28) {
                        displayTitle = displayTitle.slice(0, 26) + '...';
                    }

                    const itemEl = document.createElement('div');
                    itemEl.className = 'history-item' + (sId === currentSessionId ? ' active' : '') + (isPinned ? ' pinned' : '');
                    itemEl.setAttribute('data-session-id', sId);

                    itemEl.innerHTML = `
                        ${isPinned ? `<span class="history-item-pin-badge" title="${t('menu_pin')}"><svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><line x1="12" y1="17" x2="12" y2="22"></line><path d="M5 17h14v-2l-2-3V6a1 1 0 0 0-1-1H8a1 1 0 0 0-1 1v6l-2 3v2z"></path></svg></span>` : ''}
                        ${getSessionPlatformBadge(sId)}
                        <span class="history-item-label" title="${escapeHtml(fullTitle)}">${escapeHtml(displayTitle)}</span>
                        <button class="history-more-btn" title="${t('options_tooltip')}" aria-label="Opsi">
                            <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor">
                                <circle cx="12" cy="5" r="2"></circle>
                                <circle cx="12" cy="12" r="2"></circle>
                                <circle cx="12" cy="19" r="2"></circle>
                            </svg>
                        </button>
                    `;

                    itemEl.addEventListener('click', (e) => {
                        if (e.target.closest('.history-more-btn') || e.target.closest('.history-dropdown-menu')) return;
                        closeHistoryDropdown();
                        switchSession(sId, fullTitle);
                    });

                    const moreBtn = itemEl.querySelector('.history-more-btn');
                    moreBtn.addEventListener('click', (e) => {
                        e.stopPropagation();
                        toggleHistoryDropdown(e, sId, fullTitle, isPinned, moreBtn);
                    });

                    historyList.appendChild(itemEl);
                });
            }

            async function switchSession(sessionId, titleText) {
                if (typeof closeMobileSidebar === 'function') {
                    closeMobileSidebar();
                }
                currentSessionId = sessionId;
                localStorage.setItem('niskava_active_session', sessionId);
                document.querySelectorAll('.history-item').forEach(el => {
                    el.classList.toggle('active', el.getAttribute('data-session-id') === sessionId);
                });
                if (titleText) {
                    document.getElementById('currentSessionLabel').textContent = titleText.length > 25 ? titleText.slice(0, 22) + '...' : titleText;
                }
                loadLiveGraph(sessionId);

                // Ensure workspace switches back to active Chat View from other tabs
                if (typeof switchMainView === 'function') {
                    switchMainView('chat');
                } else {
                    const invPageView = document.getElementById('investigationsPageView');
                    const graphPageView = document.getElementById('graphPageView');
                    const composerContainer = document.getElementById('composerContainer') || document.querySelector('.composer-container');
                    const globalComplianceBox = document.getElementById('globalComplianceBox');
                    const workspaceContent = document.getElementById('workspaceContent');
                    if (workspaceContent) workspaceContent.classList.remove('graph-mode');
                    if (invPageView) invPageView.style.display = 'none';
                    if (graphPageView) graphPageView.style.display = 'none';
                    if (globalComplianceBox) globalComplianceBox.style.display = 'block';
                    if (composerContainer) composerContainer.style.display = 'flex';
                    document.querySelectorAll('.nav-link-item').forEach(item => {
                        item.classList.toggle('active', item.getAttribute('data-nav') === 'chat');
                    });
                }

                if (chatView) {
                    chatView.innerHTML = '';
                }

                try {
                    const res = await fetch(`${API_BASE}/api/chat/history?session_id=${encodeURIComponent(sessionId)}`);
                    if (!res.ok) return;
                    const data = await res.json();
                    const messages = data.messages || [];

                    if (messages.length > 0) {
                        showChatView();
                        chatView.innerHTML = '';

                        messages.forEach(m => {
                            const role = (m.role || '').toLowerCase();
                            const content = m.content || '';
                            if (role === 'user' || role === 'human') {
                                appendUserMessage(content);
                            } else {
                                const asst = createAssistantMessageElement(sessionId, m.id || m.message_id);
                                asst.contentEl.innerHTML = renderMarkdown(content);
                                asst.setLatticeStatus('done');
                                if (m.thought) {
                                    updateReactSteps(asst, typeof m.thought === 'string' ? [m.thought] : m.thought);
                                } else {
                                    const stepsBox = asst.element.querySelector('.react-steps-box');
                                    if (stepsBox) stepsBox.style.display = 'none';
                                }
                                if (m.findings && m.findings.length > 0 && asst.evidenceContainer) {
                                    asst.evidenceContainer.style.display = 'block';
                                    asst.evidenceContainer.innerHTML = renderEvidenceMatrixHTML(m.findings);
                                }
                                chatView.appendChild(asst.element);
                            }
                        });

                        // Check if session is currently active/busy in background (e.g. after browser refresh)
                        try {
                            const sessRes = await fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(sessionId)}`);
                            if (sessRes.ok) {
                                const sessData = await sessRes.json();
                                if (sessData.session && sessData.session.status === 'BUSY') {
                                    isGenerating = true;
                                    updateSendButtonState();
                                    const activeAsst = createAssistantMessageElement(sessionId);
                                    activeAsst.contentEl.innerHTML = `<em>${currentLang === 'en' ? 'Continuing market intelligence investigation in background...' : 'Melanjutkan analisis intelijen pasar di latar belakang...'}</em>`;
                                    activeAsst.setLatticeStatus('thinking');
                                    chatView.appendChild(activeAsst.element);

                                    pollBusySession(sessionId, activeAsst);
                                }
                            }
                        } catch (err) {
                            console.error('Failed to verify session status', err);
                        }

                        scrollToBottom();
                    } else {
                        chatView.innerHTML = '';
                        showHeroView();
                    }
                } catch (e) {
                    console.error('Failed to load session history', e);
                }
            }

            let busyPollInterval = null;

            function pollBusySession(sessionId, asstElement) {
                if (busyPollInterval) {
                    clearInterval(busyPollInterval);
                    busyPollInterval = null;
                }

                busyPollInterval = setInterval(async () => {
                    if (currentSessionId !== sessionId) {
                        clearInterval(busyPollInterval);
                        busyPollInterval = null;
                        return;
                    }

                    try {
                        const checkRes = await fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(sessionId)}`);
                        if (!checkRes.ok) return;
                        const data = await checkRes.json();
                        if (data.session && data.session.status !== 'BUSY') {
                            clearInterval(busyPollInterval);
                            busyPollInterval = null;
                            isGenerating = false;
                            updateSendButtonState();

                            // Reload complete messages now that background agent finished
                            if (currentSessionId === sessionId) {
                                switchSession(sessionId, document.getElementById('currentSessionLabel').textContent);
                            }
                        }
                    } catch (e) {
                        console.error('Error during busy session poll', e);
                    }
                }, 1500);
            }

            async function deleteSession(sessionId) {
                if (!confirm(t('delete_confirm'))) return;
                try {
                    let res = await fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(sessionId)}`, { method: 'DELETE' });
                    if (!res.ok) {
                        res = await fetch(`${API_BASE}/api/chat/reset?session_id=${encodeURIComponent(sessionId)}`, { method: 'POST' });
                    }
                    if (res.ok) {
                        showToast(currentLang === 'en' ? 'Chat session deleted' : 'Sesi percakapan dihapus');
                        if (currentSessionId === sessionId) {
                            btnNewResearch.click();
                        } else {
                            loadChatSessions();
                        }
                    }
                } catch(e) {
                    console.error('Error deleting session', e);
                }
            }

            // Reset All History Handler
            const btnResetHistory = document.getElementById('btnResetHistory');
            if (btnResetHistory) {
                btnResetHistory.addEventListener('click', async () => {
                    if (confirm(t('delete_confirm'))) {
                        try {
                            const res = await fetch(`${API_BASE}/api/chat/reset`, { method: 'POST' });
                            if (res.ok) {
                                showToast(currentLang === 'en' ? 'All history reset successfully' : 'Seluruh riwayat berhasil direset');
                                currentSessionId = 'WEB-' + new Date().toISOString().slice(0,10).replace(/-/g,'') + '-' + Math.floor(1000 + Math.random() * 9000);
                                chatView.innerHTML = '';
                                chatView.classList.remove('active');
                                heroView.classList.remove('hidden');
                                document.getElementById('currentSessionLabel').textContent = t('header_session_default');
                                loadChatSessions();
                            }
                        } catch(e) {
                            console.error('Error resetting history', e);
                        }
                    }
                });
            }

            // Global Search Filter (SQLite full-text /api/chat/search with 300ms debounce)
            let searchDebounceTimer = null;
            if (sidebarSearch) {
                sidebarSearch.addEventListener('input', (e) => {
                    const query = e.target.value.trim();
                    clearTimeout(searchDebounceTimer);
                    const resultsBox = document.getElementById('searchResultsList');

                    if (!query) {
                        if (resultsBox) resultsBox.style.display = 'none';
                        renderChatHistoryGroups(cachedChatSessions);
                        return;
                    }

                    // Local DOM title filtering first for instant response
                    const items = historyList.querySelectorAll('.history-item');
                    items.forEach(item => {
                        const text = item.textContent.toLowerCase();
                        item.style.display = text.includes(query.toLowerCase()) ? 'flex' : 'none';
                    });

                    // Server-side SQLite full-text search with debounce
                    searchDebounceTimer = setTimeout(async () => {
                        try {
                            const res = await fetch(`${API_BASE}/api/chat/search?q=${encodeURIComponent(query)}&limit=20`);
                            if (res.ok) {
                                const data = await res.json();
                                const results = data.results || [];
                                renderGlobalSearchResults(query, results);
                            }
                        } catch(err) {
                            console.warn('Gagal mencari pesan percakapan:', err);
                        }
                    }, 300);
                });
            }

            function renderGlobalSearchResults(query, results) {
                const resultsBox = document.getElementById('searchResultsList');
                if (!resultsBox) return;

                if (!results || results.length === 0) {
                    resultsBox.style.display = 'block';
                    resultsBox.innerHTML = `
                        <div class="search-results-header">
                            <span>${t('search_results_title')}</span>
                            <span style="cursor:pointer;" onclick="document.getElementById('searchResultsList').style.display='none'">✕</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-muted); padding:10px; text-align:center;">
                            ${t('no_messages_match')} "<strong>${escapeHtml(query)}</strong>"
                        </div>
                    `;
                    return;
                }

                resultsBox.style.display = 'block';
                let html = `
                    <div class="search-results-header">
                        <span>${t('messages_found')} (${results.length})</span>
                        <span style="cursor:pointer;" onclick="document.getElementById('searchResultsList').style.display='none'">${t('close_btn')}</span>
                    </div>
                `;

                results.forEach(item => {
                    const title = item.session_title || item.session_id || t('research_chat');
                    let contentSnippet = (item.content || '').replace(/\s+/g, ' ');
                    const qRegex = new RegExp(`(${escapeRegex(query)})`, 'gi');
                    const highlighted = escapeHtml(contentSnippet.slice(0, 140)).replace(qRegex, '<mark>$1</mark>');

                    html += `
                        <div class="search-result-item" data-session-id="${escapeHtml(item.session_id)}">
                            <div class="search-result-title">${escapeHtml(title)}</div>
                            <div class="search-result-snippet">${highlighted}</div>
                        </div>
                    `;
                });

                resultsBox.innerHTML = html;

                resultsBox.querySelectorAll('.search-result-item').forEach(el => {
                    el.addEventListener('click', () => {
                        const sid = el.getAttribute('data-session-id');
                        resultsBox.style.display = 'none';
                        switchSession(sid, currentLang === 'en' ? 'Search Results' : 'Hasil Pencarian');
                    });
                });
            }

            function escapeRegex(string) {
                return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            }

            // 4. Auto-resizing textarea & send button state
            function updateSendButtonState() {
                const btnOpenGraphFull = document.getElementById('btnOpenGraphFull');
                if (btnOpenGraphFull) {
                    btnOpenGraphFull.href = `${API_BASE}/graph`;
                }

                if (isGenerating) {
                    btnSendMessage.disabled = false;
                    btnSendMessage.classList.add('btn-aborting');
                    btnSendMessage.title = t('btn_stop_tooltip');
                    btnSendMessage.setAttribute('data-tooltip', t('btn_stop_tooltip'));
                    btnSendMessage.innerHTML = `<svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor"><rect x="4" y="4" width="16" height="16" rx="2"></rect></svg>`;
                } else {
                    btnSendMessage.classList.remove('btn-aborting');
                    btnSendMessage.title = t('btn_send_tooltip');
                    btnSendMessage.setAttribute('data-tooltip', t('btn_send_tooltip'));
                    btnSendMessage.innerHTML = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="19" x2="12" y2="5"></line><polyline points="5 12 12 5 19 12"></polyline></svg>`;
                    const hasText = chatInput.value.trim().length > 0;
                    btnSendMessage.disabled = !hasText;
                }
            }

            chatInput.addEventListener('input', () => {
                chatInput.style.height = 'auto';
                chatInput.style.height = Math.min(chatInput.scrollHeight, 180) + 'px';
                updateSendButtonState();
            });

            // Enter key to send (Shift+Enter for newline)
            chatInput.addEventListener('keydown', (e) => {
                if ((e.key === 'Enter' || e.keyCode === 13) && !e.shiftKey) {
                    e.preventDefault();
                    if (!btnSendMessage.disabled && !isGenerating) {
                        handleSendMessage();
                    }
                }
            });

            btnSendMessage.addEventListener('click', () => {
                if (isGenerating) {
                    if (currentAbortController) {
                        currentAbortController.abort();
                    }
                    fetch(`${API_BASE}/api/chat/sessions/${encodeURIComponent(currentSessionId)}/abort`, { method: 'POST' }).catch(() => {});
                    showToast(t('toast_stopped_user'));
                    return;
                }
                if (!btnSendMessage.disabled) {
                    handleSendMessage();
                }
            });

            // Export Session Report Button Handler
            const btnExportSession = document.getElementById('btnExportSession');
            if (btnExportSession) {
                btnExportSession.addEventListener('click', () => {
                    window.open(`${API_BASE}/api/chat/sessions/${encodeURIComponent(currentSessionId)}/export?format=markdown`, '_blank');
                    showToast(currentLang === 'en' ? 'Downloading Markdown investigation report...' : 'Mengunduh laporan investigasi Markdown...');
                });
            }

            // Note: Cache maintenance (Clean Expired & Flush All) is centralized in Settings Modal (07_settings.js)
            // in strict compliance with Law 5 (Credit Budget Discipline).


            // 5. New Research button
            btnNewResearch.addEventListener('click', () => {
                if (isGenerating && currentAbortController) {
                    currentAbortController.abort();
                    isGenerating = false;
                }
                if (typeof closeMobileSidebar === 'function') {
                    closeMobileSidebar();
                }
                currentSessionId = 'WEB-' + new Date().toISOString().slice(0,10).replace(/-/g,'') + '-' + Math.floor(1000 + Math.random() * 9000);
                localStorage.removeItem('niskava_active_session');
                if (chatView) {
                    chatView.innerHTML = '';
                }
                if (typeof switchMainView === 'function') {
                    switchMainView('chat');
                }
                showHeroView();
                document.getElementById('currentSessionLabel').textContent = t('header_session_default');
                chatInput.value = '';
                chatInput.style.height = 'auto';
                updateSendButtonState();
                document.querySelectorAll('.history-item').forEach(el => el.classList.remove('active'));
                loadLiveGraph(currentSessionId);
                loadChatSessions(false);
                showToast(t('toast_new_session'));
            });

            // 6. Prompt Card and Action Chip Click Handlers
            document.querySelectorAll('.prompt-card').forEach(card => {
                card.addEventListener('click', () => {
                    if (isGenerating) {
                        showToast(t('toast_wait_stream') || 'Harap tunggu investigasi yang sedang berjalan...');
                        return;
                    }
                    card.classList.add('clicked');
                    setTimeout(() => card.classList.remove('clicked'), 250);
                    const prompt = card.getAttribute('data-prompt');
                    if (prompt) {
                        showChatView();
                        chatInput.value = prompt;
                        updateSendButtonState();
                        handleSendMessage();
                    }
                });
            });

            document.querySelectorAll('.action-chip').forEach(chip => {
                chip.addEventListener('click', () => {
                    if (isGenerating) {
                        showToast(t('toast_wait_stream') || 'Harap tunggu investigasi yang sedang berjalan...');
                        return;
                    }
                    chip.classList.add('clicked');
                    setTimeout(() => chip.classList.remove('clicked'), 250);
                    const query = chip.getAttribute('data-query');
                    if (query) {
                        showChatView();
                        chatInput.value = query;
                        updateSendButtonState();
                        handleSendMessage();
                    }
                });
            });

            // 7. Riwayat Items Handled Dynamically

            // 8. Header Session Title Dropdown Menu (#sessionTitleDropdown)
            const sessionTitleDropdown = document.getElementById('sessionTitleDropdown');
            let headerDropdownEl = null;

            function closeHeaderDropdown() {
                if (headerDropdownEl) {
                    headerDropdownEl.remove();
                    headerDropdownEl = null;
                }
                if (sessionTitleDropdown) {
                    sessionTitleDropdown.classList.remove('open');
                    sessionTitleDropdown.setAttribute('aria-expanded', 'false');
                }
            }

            function toggleHeaderDropdown(e) {
                if (e) e.stopPropagation();
                if (headerDropdownEl) {
                    closeHeaderDropdown();
                    return;
                }
                if (!sessionTitleDropdown) return;

                if (typeof closeHistoryDropdown === 'function') {
                    closeHistoryDropdown();
                }
                sessionTitleDropdown.classList.add('open');
                sessionTitleDropdown.setAttribute('aria-expanded', 'true');
                headerDropdownEl = document.createElement('div');
                headerDropdownEl.className = 'header-sessions-dropdown';

                const isEn = (currentLang === 'en');
                const recentSessions = (cachedChatSessions || []).slice(0, 8);

                let listHtml = '';
                if (recentSessions.length === 0) {
                    listHtml = `<div class="header-session-empty">${isEn ? 'No recent sessions' : 'Belum ada riwayat percakapan'}</div>`;
                } else {
                    listHtml = recentSessions.map(s => {
                        const sId = s.id || s.session_id;
                        const fullTitle = formatSessionTitle(s);
                        const isActive = (sId === currentSessionId);
                        return `
                            <button type="button" class="header-session-item${isActive ? ' active' : ''}" data-session-id="${escapeHtml(sId)}" title="${escapeHtml(fullTitle)}">
                                <div style="display:flex; align-items:center; gap:8px; min-width:0; overflow:hidden;">
                                    ${getSessionPlatformBadge(sId)}
                                    <span class="header-session-title">${escapeHtml(fullTitle)}</span>
                                </div>
                                ${isActive ? '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="color:var(--primary); flex-shrink:0;"><polyline points="20 6 9 17 4 12"></polyline></svg>' : ''}
                            </button>
                        `;
                    }).join('');
                }

                headerDropdownEl.innerHTML = `
                    <div class="header-sessions-header">
                        <span class="header-sessions-title">${isEn ? 'Recent Sessions' : 'Sesi Terkini'}</span>
                        <button type="button" class="btn-header-new-session" id="btnHeaderNewSession" title="${isEn ? 'New Research' : 'Mulai Riset Baru'}">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                            <span>${isEn ? 'New' : 'Riset Baru'}</span>
                        </button>
                    </div>
                    <div class="header-sessions-list">
                        ${listHtml}
                    </div>
                `;

                headerDropdownEl.addEventListener('click', (evt) => {
                    evt.stopPropagation();
                });

                const btnNew = headerDropdownEl.querySelector('#btnHeaderNewSession');
                if (btnNew) {
                    btnNew.addEventListener('click', () => {
                        closeHeaderDropdown();
                        if (btnNewResearch) btnNewResearch.click();
                    });
                }

                headerDropdownEl.querySelectorAll('.header-session-item').forEach(item => {
                    item.addEventListener('click', () => {
                        const sid = item.getAttribute('data-session-id');
                        const target = (cachedChatSessions || []).find(s => (s.id || s.session_id) === sid);
                        const fullTitle = target ? formatSessionTitle(target) : (isEn ? 'Chat Session' : 'Sesi Percakapan');
                        closeHeaderDropdown();
                        switchSession(sid, fullTitle);
                    });
                });

                sessionTitleDropdown.appendChild(headerDropdownEl);
            }

            if (sessionTitleDropdown) {
                sessionTitleDropdown.setAttribute('aria-expanded', 'false');
                sessionTitleDropdown.addEventListener('click', (e) => {
                    toggleHeaderDropdown(e);
                });
                sessionTitleDropdown.addEventListener('keydown', (e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        toggleHeaderDropdown(e);
                    } else if (e.key === 'Escape') {
                        closeHeaderDropdown();
                    }
                });
            }

            document.addEventListener('click', (e) => {
                if (headerDropdownEl && !e.target.closest('#sessionTitleDropdown')) {
                    closeHeaderDropdown();
                }
            });
            window.addEventListener('resize', closeHeaderDropdown);
