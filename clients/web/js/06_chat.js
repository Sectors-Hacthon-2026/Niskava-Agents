// 10. Message Dispatcher & SSE Stream Handler
            async function handleSendMessage() {
                const prompt = chatInput.value.trim();
                if (!prompt || isGenerating) return;

                // Switch from hero to chat view
                if (typeof showChatView === 'function') {
                    showChatView();
                } else {
                    heroView.classList.add('hidden');
                    chatView.classList.add('active');
                }
                localStorage.setItem('niskava_active_session', currentSessionId);

                // Append user message
                appendUserMessage(prompt);
                chatInput.value = '';
                chatInput.style.height = 'auto';
                isGenerating = true;
                updateSendButtonState();

                // Append assistant placeholder
                const assistantMsgObj = createAssistantMessageElement(currentSessionId);
                chatView.appendChild(assistantMsgObj.element);
                scrollToBottom();

                let fullText = '';
                const reactSteps = [];
                const findingsEmitted = [];

                try {
                    currentAbortController = new AbortController();
                    const response = await fetch(`${API_BASE}/api/chat`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            prompt: prompt,
                            session_id: currentSessionId
                        }),
                        signal: currentAbortController.signal
                    });

                    if (!response.ok) {
                        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                    }

                    const reader = response.body.getReader();
                    const decoder = new TextDecoder();
                    let buffer = '';

                    while (true) {
                        const { done, value } = await reader.read();
                        if (done) break;

                        buffer += decoder.decode(value, { stream: true });
                        const blocks = buffer.split(/\r?\n\r?\n/);
                        buffer = blocks.pop() || '';

                        for (const block of blocks) {
                            if (!block.trim() || block.startsWith(':')) continue;
                            const eventMatch = block.match(/^event:\s*(.+)$/m);
                            const dataMatch = block.match(/^data:\s*(.+)$/m);

                            const eventName = eventMatch ? eventMatch[1].trim() : 'message';
                            const rawData = dataMatch ? dataMatch[1].trim() : '';

                            let payload = {};
                            try { payload = JSON.parse(rawData); } catch(e) {}

                            if (eventName === 'agent_thought' || payload.event === 'agent_thought') {
                                const thoughtText = payload.thought || payload.content || '';
                                if (thoughtText && !reactSteps.includes(thoughtText)) {
                                    reactSteps.push(thoughtText);
                                    updateReactSteps(assistantMsgObj, reactSteps);
                                }
                            } else if (eventName === 'agent_tool_call' || payload.event === 'agent_tool_call') {
                                const toolName = payload.tool || 'Eksekusi Alat';
                                let argsSummary = '';
                                if (payload.args && Object.keys(payload.args).length > 0) {
                                    try {
                                        argsSummary = ' (' + Object.entries(payload.args).map(([k, v]) => `${k}: ${v}`).join(', ') + ')';
                                    } catch(e) {}
                                }
                                const stepText = `Tool: ${toolName}${argsSummary}`;
                                if (!reactSteps.includes(stepText)) {
                                    reactSteps.push(stepText);
                                    updateReactSteps(assistantMsgObj, reactSteps);
                                }
                            } else if (eventName === 'agent_observation' || payload.event === 'agent_observation') {
                                const toolName = payload.tool || 'Hasil';
                                const summary = payload.summary || payload.content || '';
                                const stepText = `Observasi [${toolName}]: ${summary.slice(0, 120)}${summary.length > 120 ? '...' : ''}`;
                                if (!reactSteps.includes(stepText)) {
                                    reactSteps.push(stepText);
                                    updateReactSteps(assistantMsgObj, reactSteps);
                                }
                            } else if (eventName === 'progress_step' || payload.event === 'progress_step') {
                                const stepMsg = payload.message || payload.stage || '';
                                const stepIndex = payload.step_index ? `[${payload.step_index}/${payload.total_steps || '?'}] ` : '';
                                const stepText = `Progres: ${stepIndex}${stepMsg}`;
                                if (stepMsg && !reactSteps.includes(stepText)) {
                                    reactSteps.push(stepText);
                                    updateReactSteps(assistantMsgObj, reactSteps);
                                }
                            } else if (eventName === 'anomaly_detected' || payload.event === 'anomaly_detected') {
                                const ticker = payload.ticker ? `[${payload.ticker}] ` : '';
                                const metric = payload.metric_type || 'Volume';
                                const zScore = payload.z_score !== undefined ? `${payload.z_score >= 0 ? '+' : ''}${Number(payload.z_score).toFixed(2)}σ` : '';
                                const stepText = `Anomali Terdeteksi: ${ticker}${metric} Deviasi ${zScore}`;
                                if (!reactSteps.includes(stepText)) {
                                    reactSteps.push(stepText);
                                    updateReactSteps(assistantMsgObj, reactSteps);
                                }
                                if (payload.ticker) {
                                    updateRightPanelAnomaly(payload);
                                }
                            } else if (eventName === 'finding_emitted' || payload.event === 'finding_emitted') {
                                const status = payload.verification_status || 'VERIFIED';
                                const title = payload.title || 'Temuan Intelijen';
                                const conf = payload.confidence_score ? ` (${(payload.confidence_score * 100).toFixed(0)}%)` : '';
                                const stepText = `Bukti Kausalitas [${status}]: ${title}${conf}`;
                                if (!reactSteps.includes(stepText)) {
                                    reactSteps.push(stepText);
                                    updateReactSteps(assistantMsgObj, reactSteps);
                                }
                                findingsEmitted.push(payload);
                                if (assistantMsgObj.evidenceContainer) {
                                    assistantMsgObj.evidenceContainer.style.display = 'block';
                                    assistantMsgObj.evidenceContainer.innerHTML = renderEvidenceMatrixHTML(findingsEmitted);
                                }
                            } else if (eventName === 'agent_message_chunk' || payload.event === 'agent_message_chunk') {
                                const chunk = payload.chunk || '';
                                fullText += chunk;
                                assistantMsgObj.contentEl.innerHTML = renderMarkdown(fullText);
                                scrollToBottom();
                            } else if (eventName === 'session_error' || payload.event === 'session_error') {
                                assistantMsgObj.setLatticeStatus('error');
                                const errMsg = payload.error || 'Terjadi kesalahan sistem saat mengeksekusi analisis.';
                                assistantMsgObj.contentEl.innerHTML += `<div style="margin-top:10px; color:var(--badge-red-text); background:var(--badge-red-bg); padding:12px; border-radius:8px; border:1px solid var(--badge-red-border);">
                                    <strong>Kendala Investigasi:</strong> ${escapeHtml(errMsg)}
                                </div>`;
                                scrollToBottom();
                            } else if (eventName === 'session_complete' || eventName === 'agent_message_complete') {
                                assistantMsgObj.setLatticeStatus('done');
                                loadLiveGraph(currentSessionId);
                            } else if (eventName === 'done' || payload.event === 'done') {
                                if (payload.session_id) {
                                    currentSessionId = payload.session_id;
                                    localStorage.setItem('niskava_active_session', currentSessionId);
                                }
                                if (payload.message_id && assistantMsgObj.setMsgId) {
                                    assistantMsgObj.setMsgId(payload.message_id);
                                }
                                assistantMsgObj.setLatticeStatus('done');
                                loadChatSessions(false);
                                loadLiveGraph(currentSessionId);
                            }
                        }
                    }

                    if (!fullText) {
                        assistantMsgObj.contentEl.innerHTML = renderMarkdown(reactSteps.length > 0 ?
                            `_Investigasi intelijen pasar selesai dieksekusi dengan ${reactSteps.length} langkah observasi._` :
                            `_Analisis selesai tanpa output teks tambahan._`);
                    }

                } catch (err) {
                    if (err.name !== 'AbortError') {
                        assistantMsgObj.setLatticeStatus('error');
                        const isNetworkErr = err.message && (err.message.includes('Failed to fetch') || err.message.includes('NetworkError'));
                        assistantMsgObj.contentEl.innerHTML = `<div style="color:var(--badge-red-text); background:var(--badge-red-bg); padding:12px; border-radius:8px; border:1px solid var(--badge-red-border); line-height:1.5;">
                            <strong>Kendala Koneksi Investigasi:</strong> ${escapeHtml(err.message)}.<br>
                            ${isNetworkErr ? '<small style="opacity:0.9;">Pastikan backend daemon aktif di terminal: <code>.\\niskava.exe serve --port 8080</code></small>' : ''}
                        </div>`;
                    }
                } finally {
                    isGenerating = false;
                    updateSendButtonState();
                    if (assistantMsgObj.getLatticeStatus && assistantMsgObj.getLatticeStatus() !== 'error') {
                        assistantMsgObj.setLatticeStatus('done');
                    }
                    loadChatSessions(false);
                    loadLiveGraph(currentSessionId);
                    scrollToBottom();
                }
            }

            function appendUserMessage(text) {
                const msgDiv = document.createElement('div');
                msgDiv.className = 'chat-message';
                msgDiv.innerHTML = `
                    <div class="message-avatar user">G</div>
                    <div class="message-body">
                        <div class="message-header-row">
                            <span class="message-author">Guest</span>
                            <span class="message-time">${t('just_now')}</span>
                        </div>
                        <div class="markdown-rendered"><p>${text.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</p></div>
                    </div>
                `;
                chatView.appendChild(msgDiv);
            }

            function createAssistantMessageElement(sessionId, msgId) {
                let activeSessionId = sessionId || currentSessionId;
                let activeMsgId = msgId;
                const msgDiv = document.createElement('div');
                msgDiv.className = 'chat-message';
                msgDiv.innerHTML = `
                    <div class="message-avatar ai"><img src="data:image/png;base64,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" alt="NISKAVA Agent" class="ai-avatar-img"></div>
                    <div class="message-body">
                        <div class="message-header-row">
                            <span class="message-author">NISKAVA Agent</span>
                            <span class="message-time">${t('just_now')}</span>
                        </div>
                        
                        <!-- ReAct Accordion with LatticeLoader (React Bits) -->
                        <div class="react-steps-box">
                            <div class="react-steps-toggle" role="button" tabindex="0" title="${t('investigation_details_tooltip')}">
                                <!-- LatticeLoader Component -->
                                <div class="ll-root" data-status="working" data-shape="round">
                                    <span class="ll-grid-wrap" aria-hidden="true">
                                        <!-- 3x3 Orbit Pattern Wave -->
                                        <span class="ll-run">
                                            <span class="ll-cell animating" style="animation-delay: 0ms;"></span>
                                            <span class="ll-cell animating" style="animation-delay: 108ms;"></span>
                                            <span class="ll-cell animating" style="animation-delay: 216ms;"></span>
                                            <span class="ll-cell animating" style="animation-delay: 756ms;"></span>
                                            <span class="ll-cell hole"></span>
                                            <span class="ll-cell animating" style="animation-delay: 324ms;"></span>
                                            <span class="ll-cell animating" style="animation-delay: 648ms;"></span>
                                            <span class="ll-cell animating" style="animation-delay: 540ms;"></span>
                                            <span class="ll-cell animating" style="animation-delay: 432ms;"></span>
                                        </span>
                                        <!-- 3x3 Mark (Check: [2, 3, 5, 7], Error: [0, 2, 4, 6, 8]) -->
                                        <span class="ll-mark">
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                            <span class="ll-cell"></span>
                                        </span>
                                    </span>
                                    <span class="ll-text-wrap" aria-hidden="true">
                                        <span class="ll-text ll-text-working">${t('investigating')}</span>
                                        <span class="ll-text ll-text-done" style="display:none; opacity:0;">${t('investigation_done')}</span>
                                        <span class="ll-text ll-text-error" style="display:none; opacity:0;">${t('investigation_failed')}</span>
                                    </span>
                                    <span class="ll-timer font-mono">0.0s</span>
                                </div>

                                <div style="display:flex; align-items:center; gap:8px;">
                                    <span class="react-count-badge" style="color:var(--accent-text); font-family:var(--font-mono); font-size:11.5px; font-weight:600;">0 ${t('steps')}</span>
                                    <svg class="steps-chevron" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color:var(--text-muted);"><polyline points="6 9 12 15 18 9"></polyline></svg>
                                </div>
                            </div>
                            <div class="react-steps-content" style="display:none;"></div>
                        </div>

                        <!-- Markdown Content -->
                        <div class="markdown-rendered"></div>

                        <!-- Dedicated 3-Tier Evidence Matrix (Law 2 Non-Advisory) -->
                        <div class="evidence-matrix-container" style="display:none; margin-top:14px;"></div>

                        <!-- Action row -->
                        <div class="message-action-row">
                            <button class="btn-msg-action btn-copy-report" title="${t('copy_tooltip')}">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                <span>${t('copy')}</span>
                            </button>
                            <button class="btn-msg-action btn-export-md" title="${t('export_tooltip')}">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                                <span>${t('export')}</span>
                            </button>
                            <button class="btn-msg-action btn-fork-chat" title="${t('fork_tooltip')}">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="18" r="3"></circle><circle cx="6" cy="6" r="3"></circle><path d="M6 9a9 9 0 0 0 9 9"></path></svg>
                                <span>${t('fork')}</span>
                            </button>
                        </div>
                    </div>
                `;

                const stepsBox = msgDiv.querySelector('.react-steps-box');
                const stepsToggle = msgDiv.querySelector('.react-steps-toggle');
                const stepsContent = msgDiv.querySelector('.react-steps-content');
                const countBadge = msgDiv.querySelector('.react-count-badge');
                const stepsChevron = msgDiv.querySelector('.steps-chevron');
                const contentEl = msgDiv.querySelector('.markdown-rendered');
                const evidenceContainer = msgDiv.querySelector('.evidence-matrix-container');
                
                // LatticeLoader Elements
                const llRoot = msgDiv.querySelector('.ll-root');
                const llMark = msgDiv.querySelector('.ll-mark');
                const textWorking = msgDiv.querySelector('.ll-text-working');
                const textDone = msgDiv.querySelector('.ll-text-done');
                const textError = msgDiv.querySelector('.ll-text-error');
                const llTimer = msgDiv.querySelector('.ll-timer');

                // Live stopwatch timer
                const startTime = performance.now();
                let timerInterval = setInterval(() => {
                    const elapsed = ((performance.now() - startTime) / 1000).toFixed(1);
                    llTimer.textContent = `${elapsed}s`;
                }, 100);

                let currentStatus = 'working';
                function setLatticeStatus(status) {
                    currentStatus = status;
                    clearInterval(timerInterval);
                    const finalElapsed = ((performance.now() - startTime) / 1000).toFixed(1);
                    llTimer.textContent = `${finalElapsed}s`;
                    llRoot.setAttribute('data-status', status);

                    const markCells = llMark.querySelectorAll('.ll-cell');

                    if (status === 'done') {
                        textWorking.style.display = 'none';
                        textError.style.display = 'none';
                        textDone.style.display = 'inline';
                        setTimeout(() => { textDone.style.opacity = '1'; }, 20);

                        // Turn on checkmark cells: 2, 3, 5, 7
                        [2, 3, 5, 7].forEach(idx => {
                            if (markCells[idx]) markCells[idx].setAttribute('data-on', '');
                        });
                    } else if (status === 'error') {
                        textWorking.style.display = 'none';
                        textDone.style.display = 'none';
                        textError.style.display = 'inline';
                        setTimeout(() => { textError.style.opacity = '1'; }, 20);

                        // Turn on error cross cells: 0, 2, 4, 6, 8
                        [0, 2, 4, 6, 8].forEach(idx => {
                            if (markCells[idx]) markCells[idx].setAttribute('data-on', '');
                        });
                    }
                }

                stepsToggle.addEventListener('click', () => {
                    const isShown = stepsContent.style.display !== 'none';
                    stepsContent.style.display = isShown ? 'none' : 'flex';
                    if (stepsChevron) {
                        stepsChevron.classList.toggle('open', !isShown);
                    }
                });

                const btnCopy = msgDiv.querySelector('.btn-copy-report');
                btnCopy.addEventListener('click', () => {
                    navigator.clipboard.writeText(contentEl.innerText).then(() => {
                        showToast(t('toast_copied'));
                    });
                });

                const btnExport = msgDiv.querySelector('.btn-export-md');
                btnExport.addEventListener('click', () => {
                    const blob = new Blob([contentEl.innerText], { type: 'text/markdown' });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `NISKAVA_Report_${activeSessionId}.md`;
                    a.click();
                    showToast(t('toast_exported'));
                });

                const btnFork = msgDiv.querySelector('.btn-fork-chat');
                if (btnFork) {
                    btnFork.addEventListener('click', () => {
                        forkChatSession(activeSessionId || currentSessionId, activeMsgId);
                    });
                }

                return {
                    element: msgDiv,
                    stepsContent,
                    countBadge,
                    contentEl,
                    evidenceContainer,
                    setLatticeStatus,
                    getLatticeStatus: () => currentStatus,
                    setMsgId: (id) => { activeMsgId = id; }
                };
            }

            function updateReactSteps(msgObj, steps) {
                const unit = currentLang === 'en' ? (steps.length === 1 ? t('step') : t('steps')) : t('steps');
                msgObj.countBadge.textContent = `${steps.length} ${unit}`;
                msgObj.stepsContent.innerHTML = steps.map(s => `
                    <div class="react-step-row">
                        <span class="react-step-bullet">●</span>
                        <span>${s.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</span>
                    </div>
                `).join('');
            }

            function scrollToBottom() {
                const workspaceContent = document.getElementById('workspaceContent');
                workspaceContent.scrollTo({
                    top: workspaceContent.scrollHeight,
                    behavior: 'smooth'
                });
            }
