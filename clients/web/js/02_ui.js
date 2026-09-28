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
                    themeIconMoon.style.display = 'none';
                    themeIconSun.style.display = 'block';
                } else {
                    themeIconMoon.style.display = 'block';
                    themeIconSun.style.display = 'none';
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
                appContainer.classList.toggle('sidebar-collapsed');
                updateSidebarToggleIcon();
                const isMin = appContainer.classList.contains('sidebar-collapsed');
                showToast(isMin ? t('sidebar_collapsed_toast') : t('sidebar_expanded_toast'));
            });

            btnSidebarClose.addEventListener('click', () => {
                appContainer.classList.toggle('sidebar-collapsed');
                updateSidebarToggleIcon();
                const isMin = appContainer.classList.contains('sidebar-collapsed');
                showToast(isMin ? t('sidebar_collapsed_toast') : t('sidebar_expanded_toast'));
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

            // Helper to escape HTML characters
            function escapeHtml(str) {
                if (!str) return '';
                return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
            }

            // ==========================================================================
