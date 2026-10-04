// ==========================================
// 10. Message Dispatcher, Attachments & SSE Stream Handler
// ==========================================

if (typeof stagedAttachments === 'undefined') {
    window.stagedAttachments = [];
}

const fileAttachmentInput = document.getElementById('fileAttachmentInput');
const btnAttach = document.getElementById('btnAttach');
const attachmentPreviewBar = document.getElementById('attachmentPreviewBar');
const composerContainer = document.getElementById('composerContainer');
const composerBox = document.getElementById('composerBox') || (composerContainer ? composerContainer.querySelector('.composer-box') : null);

function formatFileSize(bytes) {
    if (typeof bytes !== 'number' || isNaN(bytes)) return '0 B';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

function getFileIconSVG(filename) {
    const ext = (filename || '').split('.').pop().toLowerCase();
    if (ext === 'pdf') {
        return `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>`;
    }
    if (['csv', 'xlsx', 'xls'].includes(ext)) {
        return `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="3" y1="9" x2="21" y2="9"></line><line x1="3" y1="15" x2="21" y2="15"></line><line x1="9" y1="3" x2="9" y2="21"></line><line x1="15" y1="3" x2="15" y2="21"></line></svg>`;
    }
    if (['png', 'jpg', 'jpeg'].includes(ext)) {
        return `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><circle cx="8.5" cy="8.5" r="1.5"></circle><polyline points="21 15 16 10 5 21"></polyline></svg>`;
    }
    return `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path><polyline points="13 2 13 9 20 9"></polyline></svg>`;
}

function renderAttachmentChips() {
    if (!attachmentPreviewBar) return;
    if (!stagedAttachments || stagedAttachments.length === 0) {
        attachmentPreviewBar.innerHTML = '';
        attachmentPreviewBar.style.display = 'none';
        attachmentPreviewBar.classList.remove('has-items');
        updateSendButtonState();
        return;
    }

    attachmentPreviewBar.style.display = 'flex';
    attachmentPreviewBar.classList.add('has-items');
    attachmentPreviewBar.innerHTML = '';

    stagedAttachments.forEach((file, index) => {
        const chip = document.createElement('div');
        chip.className = 'attachment-chip';
        chip.title = `${file.name} (${formatFileSize(file.size)})`;

        chip.innerHTML = `
            <span class="attachment-chip-icon">${getFileIconSVG(file.name)}</span>
            <span class="attachment-chip-name">${escapeHtml(file.name)}</span>
            <span class="attachment-chip-size">${formatFileSize(file.size)}</span>
            <button type="button" class="btn-remove-chip" aria-label="Hapus lampiran" data-index="${index}">&times;</button>
        `;

        const btnRemove = chip.querySelector('.btn-remove-chip');
        btnRemove.addEventListener('click', (e) => {
            e.stopPropagation();
            removeStagedAttachment(index);
        });

        attachmentPreviewBar.appendChild(chip);
    });

    updateSendButtonState();
}

function removeStagedAttachment(index) {
    if (index >= 0 && index < stagedAttachments.length) {
        stagedAttachments.splice(index, 1);
        renderAttachmentChips();
    }
}

function clearStagedAttachments() {
    stagedAttachments = [];
    if (fileAttachmentInput) fileAttachmentInput.value = '';
    renderAttachmentChips();
}
window.clearStagedAttachments = clearStagedAttachments;

function handleFilesAdded(files) {
    if (!files || files.length === 0) return;
    const maxSizeBytes = 25 * 1024 * 1024;
    const allowedExts = ['.pdf', '.txt', '.csv', '.xlsx', '.docx', '.png', '.jpg', '.jpeg', '.md', '.json'];
    let addedCount = 0;

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const ext = '.' + file.name.split('.').pop().toLowerCase();

        if (!allowedExts.includes(ext)) {
            showToast(typeof currentLang !== 'undefined' && currentLang === 'en'
                ? `Unsupported format: ${file.name}`
                : `Format tidak didukung: ${file.name}`, true);
            continue;
        }

        if (file.size > maxSizeBytes) {
            showToast(typeof currentLang !== 'undefined' && currentLang === 'en'
                ? `File exceeds 25MB limit: ${file.name}`
                : `File melebihi batas 25MB: ${file.name}`, true);
            continue;
        }

        // Avoid exact duplicates
        const exists = stagedAttachments.some(f => f.name === file.name && f.size === file.size);
        if (!exists) {
            stagedAttachments.push(file);
            addedCount++;
        }
    }

    if (addedCount > 0) {
        renderAttachmentChips();
    }
}

// Attach Button Listener
if (btnAttach && fileAttachmentInput) {
    btnAttach.addEventListener('click', (e) => {
        e.preventDefault();
        fileAttachmentInput.click();
    });

    fileAttachmentInput.addEventListener('change', () => {
        if (fileAttachmentInput.files && fileAttachmentInput.files.length > 0) {
            handleFilesAdded(fileAttachmentInput.files);
            fileAttachmentInput.value = '';
        }
    });
}

// Drag & Drop on composerContainer
if (composerContainer) {
    ['dragenter', 'dragover'].forEach(eventName => {
        composerContainer.addEventListener(eventName, (e) => {
            e.preventDefault();
            e.stopPropagation();
            composerContainer.classList.add('drag-active');
            if (composerBox) composerBox.classList.add('drag-active');
        }, false);
    });

    ['dragleave', 'dragend'].forEach(eventName => {
        composerContainer.addEventListener(eventName, (e) => {
            e.preventDefault();
            e.stopPropagation();
            composerContainer.classList.remove('drag-active');
            if (composerBox) composerBox.classList.remove('drag-active');
        }, false);
    });

    composerContainer.addEventListener('drop', (e) => {
        e.preventDefault();
        e.stopPropagation();
        composerContainer.classList.remove('drag-active');
        if (composerBox) composerBox.classList.remove('drag-active');

        if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            handleFilesAdded(e.dataTransfer.files);
        }
    }, false);
}

// Message Dispatcher & SSE Stream Handler
async function handleSendMessage() {
    const promptText = chatInput.value.trim();
    if ((!promptText && (!stagedAttachments || stagedAttachments.length === 0)) || isGenerating) return;

    // Check if user is sending from the hero screen (new chat)
    const wasHeroActive = heroView && (heroView.style.display !== 'none' && !heroView.classList.contains('hidden'));

    // Switch from hero to chat view
    if (typeof showChatView === 'function') {
        showChatView();
    } else {
        heroView.classList.add('hidden');
        chatView.classList.add('active');
    }

    // If starting from hero view, ensure chatView starts clean with zero stale elements
    if (wasHeroActive && chatView) {
        chatView.innerHTML = '';
    }

    localStorage.setItem('niskava_active_session', currentSessionId);

    // If attachments are staged, upload them first
    let uploadedAttachments = [];
    const filesToUpload = [...(stagedAttachments || [])];
    stagedAttachments = [];
    renderAttachmentChips();

    if (filesToUpload.length > 0) {
        const uploadPromises = filesToUpload.map(async (file) => {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('session_id', currentSessionId);
            try {
                const res = await fetch(`${API_BASE}/api/upload`, {
                    method: 'POST',
                    body: formData
                });
                if (res.ok) {
                    return await res.json();
                } else {
                    console.error('Upload failed with status:', res.status);
                    return null;
                }
            } catch (err) {
                console.error('Upload error:', err);
                return null;
            }
        });

        const uploadResults = await Promise.all(uploadPromises);
        uploadedAttachments = uploadResults.filter(Boolean);
    }

    const defaultPrompt = typeof currentLang !== 'undefined' && currentLang === 'en'
        ? 'Analyze and audit the attached document(s).'
        : 'Analisis dan audit dokumen terlampir.';
    const finalPrompt = promptText || defaultPrompt;

    // Append user message
    appendUserMessage(finalPrompt, uploadedAttachments);
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

    const reqPayload = {
        prompt: finalPrompt,
        session_id: currentSessionId
    };
    if (uploadedAttachments.length > 0) {
        reqPayload.attachment_ids = uploadedAttachments.map(a => a.id);
    }

    try {
        currentAbortController = new AbortController();
        const response = await fetch(`${API_BASE}/api/chat`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(reqPayload),
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
                    assistantMsgObj.setLatticeStatus('done');
                    scrollToBottom();
                } else if (eventName === 'session_error' || payload.event === 'session_error') {
                    assistantMsgObj.setLatticeStatus('error');
                    const errMsg = payload.error || 'A system error occurred while processing the analysis.';
                    assistantMsgObj.contentEl.innerHTML += `<div style="margin-top:10px; color:var(--badge-red-text); background:var(--badge-red-bg); padding:12px; border-radius:8px; border:1px solid var(--badge-red-border);">
                        <strong>Investigation Error:</strong> ${escapeHtml(errMsg)}
                    </div>`;
                    scrollToBottom();
                } else if (eventName === 'pdf_report_ready' || payload.event === 'pdf_report_ready') {
                    const filename = payload.filename || '';
                    if (filename && !assistantMsgObj.element.querySelector('.pdf-download-banner')) {
                        const downloadBtn = document.createElement('div');
                        downloadBtn.className = 'pdf-download-banner';
                        downloadBtn.innerHTML = `
                            <div style="display:flex; align-items:center; gap:10px; margin-top:12px; padding:10px 14px; background:var(--badge-green-bg,#e8f5e9); border:1px solid var(--badge-green-border,#a5d6a7); border-radius:8px; font-size:13px;">
                                <span style="font-size:20px;">📄</span>
                                <div style="flex:1">
                                    <strong style="color:var(--badge-green-text,#1b5e20);">${typeof currentLang !== 'undefined' && currentLang === 'id' ? 'Laporan PDF Audit Siap' : 'PDF Audit Report Ready'}</strong>
                                    <div style="font-family:monospace; font-size:11px; opacity:0.8; margin-top:2px;">${escapeHtml(filename)}</div>
                                </div>
                                <a href="${API_BASE}/api/reports/${encodeURIComponent(filename)}" download="${escapeHtml(filename)}" style="padding:6px 14px; background:var(--accent-primary,#1d47a0); color:#fff; border-radius:6px; text-decoration:none; font-size:12px; font-weight:600; white-space:nowrap;">⬇ Download PDF</a>
                            </div>
                        `;
                        const bodyEl = assistantMsgObj.element.querySelector('.message-body');
                        if (bodyEl) bodyEl.appendChild(downloadBtn);
                    }
                } else if (eventName === 'session_complete' || eventName === 'agent_message_complete') {
                    assistantMsgObj.setLatticeStatus('done');
                    loadLiveGraph(currentSessionId);

                    // Fallback populate content if chunks were not streamed
                    if (!fullText) {
                        const fallbackText = payload.content || payload.summary || payload.response || '';
                        if (fallbackText) {
                            fullText = fallbackText;
                            assistantMsgObj.contentEl.innerHTML = renderMarkdown(fullText);
                        }
                    }

                    // Fallback: regex detect PDF in fullText if pdf_report_ready event wasn't caught
                    const pdfMatch = fullText.match(/NISKAVA_[A-Z0-9_]+\.pdf/);
                    if (pdfMatch && !assistantMsgObj.element.querySelector('.pdf-download-banner')) {
                        const filename = pdfMatch[0];
                        const downloadBtn = document.createElement('div');
                        downloadBtn.className = 'pdf-download-banner';
                        downloadBtn.innerHTML = `
                            <div style="display:flex; align-items:center; gap:10px; margin-top:12px; padding:10px 14px; background:var(--badge-green-bg,#e8f5e9); border:1px solid var(--badge-green-border,#a5d6a7); border-radius:8px; font-size:13px;">
                                <span style="font-size:20px;">📄</span>
                                <div style="flex:1">
                                    <strong style="color:var(--badge-green-text,#1b5e20);">${typeof currentLang !== 'undefined' && currentLang === 'id' ? 'Laporan PDF Audit Siap' : 'PDF Audit Report Ready'}</strong>
                                    <div style="font-family:monospace; font-size:11px; opacity:0.8; margin-top:2px;">${escapeHtml(filename)}</div>
                                </div>
                                <a href="${API_BASE}/api/reports/${encodeURIComponent(filename)}" download="${escapeHtml(filename)}" style="padding:6px 14px; background:var(--accent-primary,#1d47a0); color:#fff; border-radius:6px; text-decoration:none; font-size:12px; font-weight:600; white-space:nowrap;">⬇ Download PDF</a>
                            </div>
                        `;
                        const bodyEl = assistantMsgObj.element.querySelector('.message-body');
                        if (bodyEl) bodyEl.appendChild(downloadBtn);
                    }
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

        if (!fullText && (!assistantMsgObj.getLatticeStatus || assistantMsgObj.getLatticeStatus() !== 'error')) {
            assistantMsgObj.contentEl.innerHTML = renderMarkdown(reactSteps.length > 0 ?
                `_Market intelligence analysis completed with ${reactSteps.length} observation steps._` :
                `_No text response received from AI model. Please verify your AI provider settings in Settings or retry your prompt._`);
        }

    } catch (err) {
        if (err.name !== 'AbortError') {
            assistantMsgObj.setLatticeStatus('error');
            const isNetworkErr = err.message && (err.message.includes('Failed to fetch') || err.message.includes('NetworkError'));
            assistantMsgObj.contentEl.innerHTML = `<div style="color:var(--badge-red-text); background:var(--badge-red-bg); padding:12px; border-radius:8px; border:1px solid var(--badge-red-border); line-height:1.5;">
                <strong>Investigation Connection Error:</strong> ${escapeHtml(err.message)}.<br>
                ${isNetworkErr ? '<small style="opacity:0.9;">Ensure the backend daemon is running in terminal: <code>niskava serve --port 8080</code></small>' : ''}
            </div>`;
        }
    } finally {
        isGenerating = false;
        updateSendButtonState();
        if (assistantMsgObj.getLatticeStatus && assistantMsgObj.getLatticeStatus() !== 'error') {
            assistantMsgObj.setLatticeStatus('done');
        }
        if (reactSteps.length === 0 && assistantMsgObj.element) {
            const sb = assistantMsgObj.element.querySelector('.react-accordion') || assistantMsgObj.element.querySelector('.react-steps-box');
            if (sb) sb.style.display = 'none';
        }
        loadChatSessions(false);
        loadLiveGraph(currentSessionId);
        scrollToBottom();
    }
}

function appendUserMessage(text, attachments) {
    const safeText = (text || '').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    const msgDiv = document.createElement('div');
    msgDiv.className = 'chat-message user-message';

    let attachmentsHTML = '';
    if (attachments && attachments.length > 0) {
        const cardsHTML = attachments.map(att => {
            const fname = att.filename || att.name || 'document';
            const fsize = formatFileSize(att.file_size || att.size || 0);
            return `
                <div class="user-attachment-card" title="${escapeHtml(fname)} (${fsize})">
                    <span class="attachment-icon">${getFileIconSVG(fname)}</span>
                    <span class="attachment-filename">${escapeHtml(fname)}</span>
                    <span class="attachment-meta">${fsize}</span>
                </div>
            `;
        }).join('');
        attachmentsHTML = `<div class="user-attachments-grid">${cardsHTML}</div>`;
    }

    msgDiv.innerHTML = `
        <div class="message-avatar user">G</div>
        <div class="message-body">
            <div class="message-header-row">
                <span class="message-author">User</span>
                <span class="message-time">${t('just_now')}</span>
            </div>
            ${attachmentsHTML}
            <div class="markdown-rendered"><p>${safeText}</p></div>
        </div>
    `;
    chatView.appendChild(msgDiv);
}

function createAssistantMessageElement(sessionId, msgId) {
    let activeSessionId = sessionId || currentSessionId;
    let activeMsgId = msgId;
    const msgDiv = document.createElement('div');
    msgDiv.className = 'chat-message assistant-message';

    const thinkingTitle = t('react_thinking_process') || (typeof currentLang !== 'undefined' && currentLang === 'en' ? 'Thinking & Investigation Process' : 'Proses Berpikir & Investigasi');

    msgDiv.innerHTML = `
        <div class="message-avatar ai"><img src="data:image/png;base64,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" alt="NISKAVA Agent" class="ai-avatar-img"></div>
        <div class="message-body">
            <div class="message-header-row">
                <span class="message-author">NISKAVA Agent</span>
                <span class="message-time">${t('just_now')}</span>
            </div>
            
            <!-- Collapsible ReAct Execution Accordion -->
            <div class="react-accordion react-steps-box" role="region" aria-label="Investigation Steps">
                <div class="react-accordion-header react-steps-toggle" role="button" tabindex="0" title="${t('investigation_details_tooltip') || 'Klik untuk rincian proses investigasi'}">
                    <div class="react-accordion-header-left">
                        <span class="react-status-dot pulsing"></span>
                        <span class="react-accordion-title">${thinkingTitle}</span>
                        <span class="react-accordion-timer font-mono">0.0s</span>
                    </div>
                    <div class="react-accordion-header-right">
                        <span class="react-accordion-badge react-count-badge">0 ${t('steps')}</span>
                        <svg class="react-accordion-chevron steps-chevron" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"></polyline></svg>
                    </div>
                </div>
                <div class="react-accordion-content react-steps-content"></div>
            </div>

            <!-- Markdown Content -->
            <div class="markdown-rendered"></div>

            <!-- Dedicated 3-Tier Evidence Matrix (Law 2 Non-Advisory) -->
            <div class="evidence-matrix-container" style="display:none; margin-top:14px;"></div>

            <!-- Quick Action Bar -->
            <div class="message-action-row">
                <button type="button" class="btn-msg-action btn-copy-report" title="${t('copy_tooltip')}">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                    <span>${t('copy')}</span>
                </button>
                <button type="button" class="btn-msg-action btn-export-md" title="${t('export_tooltip')}">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                    <span>${t('export')}</span>
                </button>
                <button type="button" class="btn-msg-action btn-fork-chat" title="${t('fork_tooltip')}">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="18" r="3"></circle><circle cx="6" cy="6" r="3"></circle><path d="M6 9a9 9 0 0 0 9 9"></path></svg>
                    <span>${t('fork')}</span>
                </button>
            </div>
        </div>
    `;

    const accordionEl = msgDiv.querySelector('.react-accordion');
    const stepsToggle = msgDiv.querySelector('.react-accordion-header');
    const stepsContent = msgDiv.querySelector('.react-accordion-content');
    const countBadge = msgDiv.querySelector('.react-accordion-badge');
    const stepsChevron = msgDiv.querySelector('.react-accordion-chevron');
    const contentEl = msgDiv.querySelector('.markdown-rendered');
    const evidenceContainer = msgDiv.querySelector('.evidence-matrix-container');
    const statusDot = msgDiv.querySelector('.react-status-dot');
    const timerEl = msgDiv.querySelector('.react-accordion-timer');

    // Live stopwatch timer
    const startTime = performance.now();
    let timerInterval = setInterval(() => {
        const elapsed = ((performance.now() - startTime) / 1000).toFixed(1);
        if (timerEl) timerEl.textContent = `${elapsed}s`;
    }, 100);

    let currentStatus = 'working';
    function setLatticeStatus(status) {
        currentStatus = status;
        clearInterval(timerInterval);
        const finalElapsed = ((performance.now() - startTime) / 1000).toFixed(1);
        if (timerEl) timerEl.textContent = `${finalElapsed}s`;

        if (statusDot) {
            statusDot.classList.remove('pulsing', 'done', 'error');
            if (status === 'done') {
                statusDot.classList.add('done');
            } else if (status === 'error') {
                statusDot.classList.add('error');
            } else {
                statusDot.classList.add('pulsing');
            }
        }
    }

    stepsToggle.addEventListener('click', () => {
        const isShown = accordionEl.classList.contains('open');
        accordionEl.classList.toggle('open');
        if (stepsChevron) {
            stepsChevron.classList.toggle('open', !isShown);
        }
    });

    const btnCopy = msgDiv.querySelector('.btn-copy-report');
    btnCopy.addEventListener('click', async () => {
        const textToCopy = contentEl.innerText || contentEl.textContent || '';
        const ok = await copyToClipboard(textToCopy);
        if (ok) {
            const origHTML = btnCopy.innerHTML;
            btnCopy.innerHTML = `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg> <span style="color:#10B981;">${typeof currentLang !== 'undefined' && currentLang === 'en' ? 'Copied!' : 'Disalin!'}</span>`;
            setTimeout(() => { btnCopy.innerHTML = origHTML; }, 2000);
            showToast(t('toast_copied'));
        } else {
            showToast(typeof currentLang !== 'undefined' && currentLang === 'en' ? 'Failed to copy to clipboard' : 'Gagal menyalin ke papan klip', true);
        }
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
    if (!Array.isArray(steps)) steps = [steps];
    const unit = typeof currentLang !== 'undefined' && currentLang === 'en'
        ? (steps.length === 1 ? t('step') : t('steps'))
        : t('steps');
    if (msgObj.countBadge) {
        msgObj.countBadge.textContent = `${steps.length} ${unit}`;
    }
    if (msgObj.stepsContent) {
        msgObj.stepsContent.innerHTML = steps.map(s => `
            <div class="react-step-row">
                <span class="react-step-bullet">●</span>
                <span class="react-step-text">${String(s || '').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</span>
            </div>
        `).join('');
        msgObj.stepsContent.scrollTop = msgObj.stepsContent.scrollHeight;
    }
}

function scrollToBottom() {
    const workspaceContent = document.getElementById('workspaceContent');
    if (workspaceContent) {
        workspaceContent.scrollTo({
            top: workspaceContent.scrollHeight,
            behavior: 'smooth'
        });
    }
}

// Accidental Navigation Guard: Warn user before reloading if analysis is actively streaming/thinking
window.addEventListener('beforeunload', (e) => {
    if (isGenerating) {
        e.preventDefault();
        e.returnValue = '';
    }
});
