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

            // Toast helper
            function showToast(msg) {
                appToast.textContent = msg;
                appToast.classList.add('show');
                setTimeout(() => {
                    appToast.classList.remove('show');
                }, 2200);
            }

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
