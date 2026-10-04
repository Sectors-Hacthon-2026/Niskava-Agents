// 1. Theme Management (Light / Dark with OS preference and LocalStorage)
            function initTheme() {
                const savedTheme = localStorage.getItem('niskava-theme');
                if (savedTheme) {
                    setTheme(savedTheme);
                } else {
                    const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
                    // Default to light theme matching media_1789825584687.jpg
                    setTheme(prefersDark ? 'dark' : 'light');
                }
            }

            function setTheme(theme) {
                document.documentElement.setAttribute('data-theme', theme);
                localStorage.setItem('niskava-theme', theme);
                if (theme === 'dark') {
                    if (themeIconMoon) themeIconMoon.style.display = 'none';
                    if (themeIconSun) themeIconSun.style.display = 'block';
                } else {
                    if (themeIconMoon) themeIconMoon.style.display = 'block';
                    if (themeIconSun) themeIconSun.style.display = 'none';
                }
            }

            // Language Switcher Handler
            const langToggleBtn = document.getElementById('langToggleBtn');
            if (langToggleBtn) {
                langToggleBtn.addEventListener('click', () => {
                    setLanguage(currentLang === 'id' ? 'en' : 'id', true);
                });
            }
            // Initialize Language from preference
            setLanguage(currentLang, false);

            btnThemeToggle.addEventListener('click', () => {
                const current = document.documentElement.getAttribute('data-theme') || 'light';
                setTheme(current === 'light' ? 'dark' : 'light');
                showToast(currentLang === 'en' ? `Switched to ${current === 'light' ? 'Dark' : 'Light'} theme` : `Beralih ke tema ${current === 'light' ? 'Gelap' : 'Terang'}`);
            });

            // 2. Sidebar Minimize/Expand Handlers & Interactive Border Resizer (<-> cursor)
            const sidebarResizer = document.getElementById('sidebarResizer');
            const sidebarSearchBox = document.getElementById('sidebarSearchBox');
            const sidebarBackdrop = document.getElementById('sidebarBackdrop');

            function isMobileView() {
                return window.innerWidth <= 768;
            }

            function openMobileSidebar() {
                appContainer.classList.add('sidebar-mobile-open');
                if (sidebar) {
                    sidebar.classList.add('mobile-open');
                }
                if (sidebarBackdrop) {
                    sidebarBackdrop.classList.add('active');
                }
            }

            function closeMobileSidebar() {
                appContainer.classList.remove('sidebar-mobile-open');
                if (sidebar) {
                    sidebar.classList.remove('mobile-open');
                }
                if (sidebarBackdrop) {
                    sidebarBackdrop.classList.remove('active');
                }
            }

            function toggleMobileSidebar() {
                if (appContainer.classList.contains('sidebar-mobile-open')) {
                    closeMobileSidebar();
                } else {
                    openMobileSidebar();
                }
            }

            if (sidebarBackdrop) {
                sidebarBackdrop.addEventListener('click', closeMobileSidebar);
            }

            window.addEventListener('resize', () => {
                if (!isMobileView()) {
                    closeMobileSidebar();
                }
            });

            function updateSidebarToggleIcon() {
                const isMin = appContainer.classList.contains('sidebar-collapsed');
                if (isMin) {
                    btnSidebarClose.innerHTML = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>';
                    btnSidebarClose.setAttribute('title', currentLang === 'en' ? 'Expand Sidebar' : 'Buka Sidebar');
                    btnSidebarClose.setAttribute('data-tooltip', currentLang === 'en' ? 'Expand Sidebar' : 'Buka Sidebar');
                } else {
                    btnSidebarClose.innerHTML = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>';
                    btnSidebarClose.setAttribute('title', currentLang === 'en' ? 'Collapse Sidebar' : 'Perkecil Sidebar');
                    btnSidebarClose.removeAttribute('data-tooltip');
                }
            }

            btnToggleSidebar.addEventListener('click', () => {
                if (isMobileView()) {
                    toggleMobileSidebar();
                } else {
                    appContainer.classList.toggle('sidebar-collapsed');
                    updateSidebarToggleIcon();
                    const isMin = appContainer.classList.contains('sidebar-collapsed');
                    showToast(isMin ? t('sidebar_collapsed_toast') : t('sidebar_expanded_toast'));
                }
            });

            btnSidebarClose.addEventListener('click', () => {
                if (isMobileView()) {
                    closeMobileSidebar();
                } else {
                    appContainer.classList.toggle('sidebar-collapsed');
                    updateSidebarToggleIcon();
                    const isMin = appContainer.classList.contains('sidebar-collapsed');
                    showToast(isMin ? t('sidebar_collapsed_toast') : t('sidebar_expanded_toast'));
                }
            });

            // Expand when clicking search box in minimized state
            if (sidebarSearchBox) {
                sidebarSearchBox.addEventListener('click', (e) => {
                    if (appContainer.classList.contains('sidebar-collapsed')) {
                        appContainer.classList.remove('sidebar-collapsed'); updateSidebarToggleIcon();
                        setTimeout(() => sidebarSearch.focus(), 150);
                    }
                });
            }

            // Open settings when clicking profile in minimized state
            if (sidebarProfile) {
                sidebarProfile.addEventListener('click', (e) => {
                    if (appContainer.classList.contains('sidebar-collapsed')) {
                        if (btnSettings) btnSettings.click();
                    }
                });
            }

            // Interactive boundary hover & drag resizer (<-> cursor)
            if (sidebarResizer) {
                let isResizing = false;
                let startX = 0;
                let hasMoved = false;

                sidebarResizer.addEventListener('mousedown', (e) => {
                    isResizing = true;
                    startX = e.clientX;
                    hasMoved = false;
                    sidebarResizer.classList.add('active');
                    document.body.style.cursor = 'col-resize';
                    document.body.style.userSelect = 'none';
                });

                window.addEventListener('mousemove', (e) => {
                    if (!isResizing) return;
                    if (Math.abs(e.clientX - startX) > 6) {
                        hasMoved = true;
                    }
                    if (e.clientX > 140) {
                        appContainer.classList.remove('sidebar-collapsed'); updateSidebarToggleIcon();
                    } else if (e.clientX < 110) {
                        appContainer.classList.add('sidebar-collapsed'); updateSidebarToggleIcon();
                    }
                });

                window.addEventListener('mouseup', () => {
                    if (!isResizing) return;
                    isResizing = false;
                    sidebarResizer.classList.remove('active');
                    document.body.style.cursor = '';
                    document.body.style.userSelect = '';
                    if (!hasMoved) {
                        // Single click on the border (<->): toggle minimize / expand
                        appContainer.classList.toggle('sidebar-collapsed');
                        updateSidebarToggleIcon();
                        const isMin = appContainer.classList.contains('sidebar-collapsed');
                        showToast(isMin ? t('sidebar_collapsed_toast') : t('sidebar_expanded_toast'));
                    }
                });
            }

            // Toast helper with multi-variant support & auto-dismiss
            let toastTimer = null;
            function showToast(msg, type = 'info', duration = 2600) {
                const toastEl = document.getElementById('appToast');
                const toastMsgEl = document.getElementById('appToastMessage');
                const toastIconEl = document.getElementById('appToastIcon');
                const toastCloseBtn = document.getElementById('appToastClose');
                if (!toastEl) return;

                // Handle legacy boolean type (e.g. showToast('Error', true))
                if (type === true) type = 'error';
                if (!type || type === false) type = 'info';

                // Select SVG icon based on variant
                let iconSvg = '';
                if (type === 'success') {
                    iconSvg = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>';
                } else if (type === 'error') {
                    iconSvg = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>';
                } else if (type === 'warning') {
                    iconSvg = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>';
                } else {
                    iconSvg = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>';
                }

                if (toastIconEl) toastIconEl.innerHTML = iconSvg;
                if (toastMsgEl) {
                    toastMsgEl.textContent = msg;
                } else {
                    toastEl.textContent = msg;
                }

                // Clean existing variant classes and add current
                toastEl.classList.remove('toast-success', 'toast-error', 'toast-warning', 'toast-info');
                toastEl.classList.add(`toast-${type}`);
                toastEl.classList.add('show');

                if (toastTimer) clearTimeout(toastTimer);
                toastTimer = setTimeout(() => {
                    toastEl.classList.remove('show');
                }, duration);

                if (toastCloseBtn) {
                    toastCloseBtn.onclick = () => {
                        if (toastTimer) clearTimeout(toastTimer);
                        toastEl.classList.remove('show');
                    };
                }
            }

            // Modern Confirmation Dialog Modal (Promise-based, accessible, dual-theme)
            function showConfirmDialog({ title, message, confirmText, cancelText, type = 'danger' } = {}) {
                return new Promise((resolve) => {
                    const modal = document.getElementById('confirmModal');
                    const titleEl = document.getElementById('confirmDialogTitle');
                    const msgEl = document.getElementById('confirmDialogMessage');
                    const iconEl = document.getElementById('confirmDialogIcon');
                    const btnCancel = document.getElementById('btnConfirmCancel');
                    const btnAction = document.getElementById('btnConfirmAction');

                    if (!modal) {
                        // Fallback if modal DOM element is missing
                        const fallbackConfirm = window.confirm(message || title || 'Confirm action?');
                        return resolve(fallbackConfirm);
                    }

                    const isEn = (typeof currentLang !== 'undefined' && currentLang === 'en');
                    const defaultTitle = isEn ? 'Confirm Action' : 'Konfirmasi Tindakan';
                    const defaultCancel = isEn ? 'Cancel' : 'Batal';
                    const defaultConfirm = isEn ? 'Continue' : 'Lanjutkan';

                    if (titleEl) titleEl.textContent = title || defaultTitle;
                    if (msgEl) msgEl.textContent = message || '';
                    if (btnCancel) btnCancel.textContent = cancelText || defaultCancel;
                    if (btnAction) {
                        btnAction.textContent = confirmText || defaultConfirm;
                        btnAction.className = `btn-confirm-action ${type}`;
                    }

                    if (iconEl) {
                        iconEl.className = `confirm-icon-badge ${type}`;
                        if (type === 'danger') {
                            iconEl.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>';
                        } else if (type === 'warning') {
                            iconEl.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>';
                        } else {
                            iconEl.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>';
                        }
                    }

                    const inputWrap = document.getElementById('confirmInputWrap');
                    if (inputWrap) inputWrap.style.display = 'none';

                    modal.style.display = 'flex';
                    if (btnCancel) btnCancel.focus();

                    function cleanup(result) {
                        modal.style.display = 'none';
                        window.removeEventListener('keydown', handleKey);
                        modal.removeEventListener('click', handleBackdrop);
                        if (btnCancel) btnCancel.onclick = null;
                        if (btnAction) btnAction.onclick = null;
                        resolve(result);
                    }

                    function handleKey(e) {
                        if (e.key === 'Escape') {
                            e.preventDefault();
                            cleanup(false);
                        } else if (e.key === 'Enter') {
                            if (document.activeElement === btnCancel) {
                                e.preventDefault();
                                cleanup(false);
                            } else {
                                e.preventDefault();
                                cleanup(true);
                            }
                        }
                    }

                    function handleBackdrop(e) {
                        if (e.target === modal) {
                            cleanup(false);
                        }
                    }

                    if (btnCancel) btnCancel.onclick = () => cleanup(false);
                    if (btnAction) btnAction.onclick = () => cleanup(true);
                    modal.addEventListener('click', handleBackdrop);
                    window.addEventListener('keydown', handleKey);
                });
            }

            // Modern Accessible Input Dialog Modal (Promise-based, supports text input, Esc/Enter, dual-theme)
            function showInputDialog({ title, message, defaultValue = '', placeholder = '', confirmText, cancelText, type = 'primary' } = {}) {
                return new Promise((resolve) => {
                    const modal = document.getElementById('confirmModal');
                    const titleEl = document.getElementById('confirmDialogTitle');
                    const msgEl = document.getElementById('confirmDialogMessage');
                    const iconEl = document.getElementById('confirmDialogIcon');
                    const inputWrap = document.getElementById('confirmInputWrap');
                    const inputEl = document.getElementById('confirmDialogInput');
                    const btnCancel = document.getElementById('btnConfirmCancel');
                    const btnAction = document.getElementById('btnConfirmAction');

                    if (!modal || !inputEl || !inputWrap) {
                        const fallbackVal = window.prompt(message || title || 'Enter value:', defaultValue);
                        return resolve(fallbackVal);
                    }

                    const isEn = (typeof currentLang !== 'undefined' && currentLang === 'en');
                    const defaultTitle = isEn ? 'Input Required' : 'Masukkan Data';
                    const defaultCancel = isEn ? 'Cancel' : 'Batal';
                    const defaultConfirm = isEn ? 'Save' : 'Simpan';

                    if (titleEl) titleEl.textContent = title || defaultTitle;
                    if (msgEl) msgEl.textContent = message || '';
                    if (btnCancel) btnCancel.textContent = cancelText || defaultCancel;
                    if (btnAction) {
                        btnAction.textContent = confirmText || defaultConfirm;
                        btnAction.className = `btn-confirm-action ${type}`;
                    }

                    if (iconEl) {
                        iconEl.className = `confirm-icon-badge ${type}`;
                        iconEl.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>';
                    }

                    inputWrap.style.display = 'block';
                    inputEl.value = defaultValue || '';
                    inputEl.placeholder = placeholder || (isEn ? 'Enter value...' : 'Masukkan nilai...');

                    modal.style.display = 'flex';
                    setTimeout(() => {
                        inputEl.focus();
                        inputEl.select();
                    }, 50);

                    function cleanup(result) {
                        modal.style.display = 'none';
                        inputWrap.style.display = 'none';
                        window.removeEventListener('keydown', handleKey);
                        modal.removeEventListener('click', handleBackdrop);
                        if (btnCancel) btnCancel.onclick = null;
                        if (btnAction) btnAction.onclick = null;
                        resolve(result);
                    }

                    function handleKey(e) {
                        if (e.key === 'Escape') {
                            e.preventDefault();
                            cleanup(null);
                        } else if (e.key === 'Enter') {
                            if (document.activeElement === btnCancel) {
                                e.preventDefault();
                                cleanup(null);
                            } else {
                                e.preventDefault();
                                cleanup(inputEl.value);
                            }
                        }
                    }

                    function handleBackdrop(e) {
                        if (e.target === modal) {
                            cleanup(null);
                        }
                    }

                    if (btnCancel) btnCancel.onclick = () => cleanup(null);
                    if (btnAction) btnAction.onclick = () => cleanup(inputEl.value);
                    modal.addEventListener('click', handleBackdrop);
                    window.addEventListener('keydown', handleKey);
                });
            }

            window.showConfirmDialog = showConfirmDialog;
            window.showInputDialog = showInputDialog;
            window.showToast = showToast;

            // Keyboard shortcut Cmd/Ctrl + K to focus search
            window.addEventListener('keydown', (e) => {
                if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
                    e.preventDefault();
                    if (appContainer.classList.contains('sidebar-collapsed')) {
                        appContainer.classList.remove('sidebar-collapsed'); updateSidebarToggleIcon();
                    }
                    sidebarSearch.focus();
                }
            });

            // View visibility state controllers (Clean synchronization without conflicting inline styles)
            function showChatView() {
                const graphPageView = document.getElementById('graphPageView');
                const invPageView = document.getElementById('investigationsPageView');
                const composerContainer = document.getElementById('composerContainer') || document.querySelector('.composer-container');
                const globalComplianceBox = document.getElementById('globalComplianceBox');
                const workspaceContent = document.getElementById('workspaceContent');

                if (graphPageView) graphPageView.style.display = 'none';
                if (invPageView) invPageView.style.display = 'none';
                if (workspaceContent) workspaceContent.classList.remove('graph-mode');
                if (globalComplianceBox) globalComplianceBox.style.display = 'block';
                if (composerContainer) composerContainer.style.display = 'flex';

                document.querySelectorAll('.nav-link-item').forEach(item => {
                    item.classList.toggle('active', item.getAttribute('data-nav') === 'chat');
                });

                if (heroView) {
                    heroView.classList.add('hidden');
                    heroView.style.display = 'none';
                }
                if (chatView) {
                    chatView.classList.add('active');
                    chatView.style.display = 'flex';
                }
                const chips = document.getElementById('quickActionChips');
                if (chips) chips.style.display = 'none';
            }

            function showHeroView() {
                const graphPageView = document.getElementById('graphPageView');
                const invPageView = document.getElementById('investigationsPageView');
                const composerContainer = document.getElementById('composerContainer') || document.querySelector('.composer-container');
                const globalComplianceBox = document.getElementById('globalComplianceBox');
                const workspaceContent = document.getElementById('workspaceContent');

                if (graphPageView) graphPageView.style.display = 'none';
                if (invPageView) invPageView.style.display = 'none';
                if (workspaceContent) workspaceContent.classList.remove('graph-mode');
                if (globalComplianceBox) globalComplianceBox.style.display = 'block';
                if (composerContainer) composerContainer.style.display = 'flex';

                document.querySelectorAll('.nav-link-item').forEach(item => {
                    item.classList.toggle('active', item.getAttribute('data-nav') === 'chat');
                });

                if (chatView) {
                    chatView.classList.remove('active');
                    chatView.style.display = 'none';
                    chatView.innerHTML = '';
                }
                if (heroView) {
                    heroView.classList.remove('hidden');
                    heroView.style.display = 'flex';
                }
                const chips = document.getElementById('quickActionChips');
                if (chips) chips.style.display = 'flex';
            }

            // Helper to escape HTML characters
            function escapeHtml(str) {
                if (!str) return '';
                return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
            }

            // Universal Clipboard Copy with Fallback for non-HTTPS / LAN IP environments
            async function copyToClipboard(text) {
                if (!text) return false;
                if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
                    try {
                        await navigator.clipboard.writeText(text);
                        return true;
                    } catch (e) {
                        console.warn('navigator.clipboard.writeText failed, falling back to execCommand:', e);
                    }
                }
                try {
                    const textarea = document.createElement('textarea');
                    textarea.value = text;
                    textarea.style.position = 'fixed';
                    textarea.style.left = '-9999px';
                    textarea.style.top = '-9999px';
                    textarea.style.opacity = '0';
                    document.body.appendChild(textarea);
                    textarea.focus();
                    textarea.select();
                    const success = document.execCommand('copy');
                    document.body.removeChild(textarea);
                    return success;
                } catch (err) {
                    console.error('execCommand copy fallback failed:', err);
                    return false;
                }
            }
            window.copyToClipboard = copyToClipboard;

            // ==========================================================================
