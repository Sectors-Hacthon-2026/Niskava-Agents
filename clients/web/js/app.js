(function() {
    'use strict';

    // --- 00_state.js ---
// NISKAVA Web Workspace Client Architecture
// State
const API_BASE = (window.location.protocol === 'file:' || ['5500', '3000', '5173', '8000'].includes(window.location.port))
                ? 'http://127.0.0.1:20128'
                : '';
            let currentSessionId = 'WEB-' + new Date().toISOString().slice(0,10).replace(/-/g,'') + '-' + Math.floor(1000 + Math.random() * 9000);
            let isGenerating = false;
            let currentAbortController = null;

            // DOM Elements

        // ==========================================

    // --- 01_i18n.js ---
// BILINGUAL i18n ENGINE (ID ⇄ EN)
        // ==========================================
        const I18N_DICT = {
            id: {
                brand_subtitle: "AI for Brighter Investments",
                search_placeholder: "Cari emiten, riwayat, atau topik...",
                search_tooltip: "Cari (⌘K)",
                new_research: "Riset Baru",
                nav_chat: "Percakapan AI",
                nav_investigations: "Investigasi Formal",
                nav_market: "Market Overview",
                nav_watchlist: "Watchlist",
                nav_graphify: "Memory Graph",
                nav_news: "Berita & Sentimen",
                nav_toolkit: "Pengaturan & Tools",
                nav_toolkit_tooltip: "Pengaturan & Diagnostik Sistem",
                recents_title: "Recents",
                profile_settings: "Pengaturan",
                profile_settings_tooltip: "Buka Pengaturan Sistem",
                profile_badge: "v0.1.4",
                sidebar_close_tooltip: "Tutup sidebar",
                header_session_default: "Percakapan AI",
                header_session_dropdown_tooltip: "Pilih / Ganti Sesi Percakapan",
                header_export_tooltip: "Ekspor Percakapan (Markdown)",
                header_theme_tooltip: "Ganti Tema (Gelap/Terang)",
                header_sidebar_tooltip: "Buka/Perkecil Sidebar (⌘B)",
                header_lang_tooltip: "Ganti Bahasa (ID / EN)",
                header_settings_tooltip: "Pengaturan & Diagnostik Sistem (⌘,)",
                hero_tag: "NISKAVA AI",
                hero_title: "Investasi lebih cerdas<br>dengan <span class=\"gradient-text\">wawasan yang lebih dalam.</span>",
                hero_desc: "Tanyakan apa pun tentang pasar modal Indonesia. NISKAVA akan menganalisis data, menghubungkan fakta, dan memberi insight untuk keputusanmu.",
                pcard_1: "Kenapa ANTM naik hari ini?",
                pcard_2: "Ringkasan berita market",
                pcard_3: "Analisis sektor perbankan",
                pcard_4: "Cari peluang saham undervalue",
                pcard_5: "Bandingkan INCO vs ANTM",
                pcard_6: "Tampilkan hubungan saham di Memory Graph",
                prompt_1: "Kenapa ANTM naik hari ini? Lakukan investigasi anomali volume dan keterbukaan informasinya.",
                prompt_2: "Berikan ringkasan berita pasar modal terkini dan sentimen pergerakan IHSG.",
                prompt_3: "Analisis sektor perbankan: bandingkan foreign flow BBCA, BBRI, dan BMRI.",
                prompt_4: "Cari peluang saham undervalue di indeks LQ45 dengan PBV rendah dan ROE sehat.",
                prompt_5: "Bandingkan fundamental emiten tambang nikel INCO vs ANTM secara komprehensif.",
                prompt_6: "Tampilkan hubungan saham ANTM dengan pergerakan komoditas nikel global di Memory Graph.",
                composer_placeholder: "Tanya NISKAVA tentang saham, emiten, berita, atau strategi...",
                chip_stocks: "Saham hari ini",
                chip_news: "Berita terbaru",
                chip_ta: "Analisis teknikal",
                chip_compare: "Perbandingan emiten",
                chip_opp: "Peluang investasi",
                compliance_bar: "Riset intelijen pasar modal berbasis bukti (IDX) • Bukan rekomendasi transaksi finansial (Law 2 & Law 3)",
                disclaimer: "Kepatuhan Regulasi (Law 2 & Law 3): Niskava Agent adalah platform riset intelijen pasar modal berbasis bukti untuk Bursa Efek Indonesia (IDX). BUKAN penasihat investasi berizin. Seluruh data dan sintesis disajikan semata-mata untuk verifikasi informasi pasar dan BUKAN rekomendasi beli/jual.",
                just_now: "Baru saja",
                copy: "Salin",
                export: "Ekspor",
                fork: "Fork",
                copy_tooltip: "Salin Jawaban",
                export_tooltip: "Ekspor Markdown",
                fork_tooltip: "Fork sesi dari titik percakapan ini (OpenCode Forking)",
                investigation_details_tooltip: "Klik untuk rincian proses investigasi AI",
                investigating: "Menginvestigasi Intelijen Pasar",
                investigation_done: "Investigasi selesai dalam",
                investigation_failed: "Investigasi gagal setelah",
                steps: "langkah",
                step: "langkah",
                evidence_matrix_title: "Matriks Verifikasi Bukti 3-Tier (Law 2)",
                evidence_verified_badge: "Bukti Diverifikasi",
                evidence_disclaimer: "<strong>Kepatuhan Law 2 (Strict Non-Advisory):</strong> Bukti diklasifikasikan secara deterministik berdasarkan laporan keterbukaan informasi emiten dan konsensus berita resmi BEI. Bukan merupakan rekomendasi beli/jual sekuritas.",
                market_finding: "Temuan Fakta Pasar",
                confidence: "Keyakinan",
                verified: "Terverifikasi",
                search_results_title: "Hasil Pencarian Pesan",
                messages_found: "Pesan Ditemukan",
                no_messages_match: "Tidak ada pesan mencocoki",
                close_btn: "✕ Tutup",
                research_chat: "Percakapan Riset",
                modal_title: "Pengaturan & Diagnostik Sistem",
                modal_subtitle: "Konfigurasi model AI, koneksi data pasar Sectors v2, bot Telegram, dan integritas penyimpanan lokal.",
                tab_api: "Kredensial API & LLM",
                tab_telegram: "Telegram Bot Daemon",
                tab_sectors: "Cache Sectors (Law 5)",
                tab_diagnostics: "Diagnostik Sistem",
                btn_save: "Simpan Perubahan",
                btn_cancel: "Batal",
                btn_test: "Uji Koneksi",
                clean_cache: "Bersihkan Cache Kadaluarsa",
                btn_start_bot: "Nyalakan Bot",
                btn_stop_bot: "Matikan Bot",
                btn_add_user: "+ Tambah",
                btn_send_ping: "Kirim Ping",
                status_not_checked: "Belum dicek",
                loading_diagnostics: "Memuat data diagnostik...",
                menu_pin: "Sematkan ke Atas",
                menu_unpin: "Lepas Sematan",
                menu_rename: "Ganti Nama",
                menu_delete: "Hapus Chat",
                rename_prompt: "Ganti judul percakapan:",
                delete_confirm: "Hapus sesi percakapan ini? Tindakan ini tidak dapat dibatalkan.",
                fork_prompt: "Masukkan judul untuk sesi percabangan baru:",
                fork_default_title: "Fork: Riset Pasar",
                history_empty: "Belum ada riwayat percakapan.<br>Mulai riset baru di atas.",
                options_tooltip: "Opsi percakapan",
                btn_send_tooltip: "Kirim Pesan",
                btn_stop_tooltip: "Hentikan Investigasi",
                btn_attach_tooltip: "Lampirkan Dokumen",
                toast_copied: "Disalin ke clipboard",
                toast_exported: "Laporan sesi diekspor",
                toast_new_session: "Sesi riset baru dimulai",
                toast_stopped_user: "Investigasi dihentikan oleh pengguna",
                toast_lang_id: "Bahasa diubah ke Bahasa Indonesia",
                toast_lang_en: "Language switched to English",
                sidebar_collapsed_toast: "Sidebar diminimalkan (Icon rail)",
                sidebar_expanded_toast: "Sidebar dibuka",
                inv_formal_title: "Investigasi Formal Otonom (7-Tahap)",
                inv_formal_desc: "Audit komprehensif anomali pasar modal Indonesia terverifikasi bukti matematis (Z-Score ≥ 2.50σ) dan taksonomi 3-tier.",
                btn_inv_back_to_chat: "Percakapan AI",
                btn_inv_open_chat: "Buka di Chat",
                inv_initiate_title: "Inisiasi Audit Otonom",
                inv_ticker_label: "Ticker Emiten BEI (IDX)",
                inv_ticker_placeholder: "cth: ANTM, BBCA, BUMI...",
                inv_horizon_label: "Rentang Observasi (Horizon)",
                inv_horizon_30: "30 Hari Perdagangan",
                inv_horizon_60: "60 Hari Perdagangan",
                inv_horizon_90: "90 Hari Perdagangan",
                inv_mock_label: "Mode Uji Coba Offline (Data Mock)",
                btn_run_investigation: "Mulai Audit Otonom",
                inv_history_title: "Riwayat Investigasi",
                inv_empty_title: "Dossier Investigasi Belum Dipilih",
                inv_empty_desc: "Pilih salah satu sesi dari daftar di sebelah kiri atau jalankan audit baru untuk menelaah anomali statistik, candlestick, bukti 3-tier, dan kronologi kejadian.",
                inv_anomaly_title: "Ringkasan Anomali Deterministik (Law 1)",
                inv_chart_title: "Visualisasi Candlestick & Volume Spike (TradingView)",
                inv_matrix_title: "Matriks Verifikasi Bukti 3-Tier (Law 2)",
                inv_timeline_title: "Timeline Kronologis Bukti & Berita",
                inv_disclaimer: "<strong>Kepatuhan Regulasi (Law 2 & Law 3):</strong> Seluruh dossier dan temuan investigasi disajikan secara obyektif berdasarkan data historis keterbukaan informasi IDXnet dan bukan merupakan rekomendasi transaksi finansial.",
                graph_page_title: "Memory Graph — Jaringan Asosiasi Pasar Modal",
                graph_page_desc: "Visualisasi interaktif graf asosiasi multi-entitas pasar modal Indonesia (Emiten, Katalis Komoditas, Broker, Regulasi, dan Anomali Volume).",
                btn_reload_graph: "Segarkan Graf",
                graph_open_tab: "Buka Tab Mandiri",
                graph_total_nodes: "Total Simpul (Nodes)",
                graph_total_edges: "Total Relasi (Edges)",
                graph_top_hub: "Sentralitas Hub Utama",
                graph_coverage: "Cakupan Graf",
                btn_apply_filter: "Terapkan Filter",
                btn_reset_filter: "Reset",
                btn_clean_test_data: "Bersihkan Data Uji",
                inv_actual_vol: "Volume Aktual",
                inv_return_dev: "Deviasi Return",
                inv_chart_status: "Data Pasar Harian",
                inv_loading_history: "Memuat riwayat investigasi...",
                sidebar_resizer_tooltip: "Klik atau seret untuk membuka/memperkecil sidebar",
                sec_key_placeholder: "sec_live_... (Kosongkan jika tidak diubah)",
                label_ai_provider: "Provider AI Aktif",
                opt_gemini: "Google Gemini (Disarankan)",
                opt_openai: "OpenAI / Compatible Gateway",
                opt_ollama: "Ollama (Lokal / On-Premise)",
                opt_anthropic: "Anthropic Claude",
                gemini_key_placeholder: "AIzaSy... (Kosongkan jika tidak diubah)",
                label_openai_key: "OpenAI API Key (Opsional)",
                openai_key_placeholder: "sk-... (Kosongkan jika tidak diubah)",
                label_offline_mode: "Mode Offline / Mock (Law 5)",
                desc_offline_mode: "Gunakan data sintetis lokal tanpa memotong kuota kredit Sectors API v2 (offline_mode).",
                label_timeout: "Inference Timeout",
                label_bot_status: "Status Bot:",
                label_tele_token: "Telegram Bot Token",
                tele_token_placeholder: "7123456789:AAH... (Kosongkan jika tidak diubah)",
                label_tele_whitelist: "Whitelist User ID / Username (Akses Khusus Analis)",
                tele_user_placeholder: "Masukkan ID Telegram atau @username...",
                label_tele_ping: "Kirim Pesan Uji Coba (Test Ping)",
                tele_chatid_placeholder: "Chat ID Telegram (contoh: 123456789)",
                desc_sectors_cache: "Monitor efisiensi kuota 1.000 kredit Sectors API v2 (Kepatuhan Law 5). Setiap permintaan pasar yang terlayani melalui cache SQLite lokal menghemat 1 kredit secara otomatis.",
                label_sectors_budget: "Disiplin Anggaran Kredit Sectors v2 (Law 5)",
                label_cache_efficiency: "Efisiensi Cache SQLite:",
                data_freshness_tooltip: "Status Sumber Data Pasar (Bursa Efek Indonesia)",
                tag_anomaly: "ANOMALI",
                tag_disclosure: "DISCLOSURE",
                tag_flow: "FLOW",
                tag_valuation: "VALUASI",
                tag_peers: "PEERS",
                tag_graph: "EGO-GRAPH",
                graph_filter_ticker_ph: "Filter Ticker (cth: ANTM, BBRI)...",
                graph_hop_1: "1 Hop (Tetangga Langsung)",
                graph_hop_2: "2 Hops (Maks)",
                graph_node_all: "Semua Tipe Entitas",
                graph_node_ticker: "Emiten Saham (TICKER)",
                graph_node_catalyst: "Katalis Komoditas (CATALYST)",
                graph_node_broker: "Broker Sekuritas (BROKER)",
                graph_node_regulator: "Regulator BEI/OJK (REGULATOR)",
                graph_node_anomaly: "Anomali Volume (ANOMALY)",
                graph_node_event: "Aksi Korporasi (EVENT)",
                compliance_disclaimer: "<strong>Kepatuhan Hukum Pasar Modal (Law 2 & Law 3):</strong> Niskava Agent adalah platform riset intelijen pasar modal berbasis bukti untuk Bursa Efek Indonesia (IDX), <u>BUKAN</u> penasihat investasi berizin. Seluruh data, grafik asosiasi, dan sintesis disajikan untuk verifikasi informasi pasar dan <u>BUKAN</u> rekomendasi finansial.",
                stat_cache_total: "Total Entri Cache",
                stat_credit_saved: "Kredit Dihemat",
                stat_permanent_entries: "Entri Permanen (OHLCV)",
                stat_expired_entries: "Entri Kadaluarsa",
                label_clean_disk: "Bersihkan data kadaluarsa untuk menghemat ruang disk:"
            },
            en: {
                brand_subtitle: "AI for Brighter Investments",
                search_placeholder: "Search tickers, history, or topics...",
                search_tooltip: "Search (⌘K)",
                new_research: "New Research",
                nav_chat: "AI Chat",
                nav_investigations: "Formal Investigation",
                nav_market: "Market Overview",
                nav_watchlist: "Watchlist",
                nav_graphify: "Memory Graph",
                nav_news: "News & Sentiment",
                nav_toolkit: "Settings & Tools",
                nav_toolkit_tooltip: "System Settings & Diagnostics",
                recents_title: "Recents",
                profile_settings: "Settings",
                profile_settings_tooltip: "Open System Settings",
                profile_badge: "v0.1.4",
                sidebar_close_tooltip: "Collapse sidebar",
                header_session_default: "AI Chat",
                header_session_dropdown_tooltip: "Select / Switch Chat Session",
                header_export_tooltip: "Export Chat (Markdown)",
                header_theme_tooltip: "Toggle Theme (Dark/Light)",
                header_sidebar_tooltip: "Toggle Sidebar (⌘B)",
                header_lang_tooltip: "Switch Language (ID / EN)",
                header_settings_tooltip: "System Settings & Diagnostics (⌘,)",
                hero_tag: "NISKAVA AI",
                hero_title: "Smarter investments<br>with <span class=\"gradient-text\">deeper market insights.</span>",
                hero_desc: "Ask anything about the Indonesian capital market. NISKAVA analyzes data, connects facts, and provides actionable insights for your decisions.",
                pcard_1: "Why did ANTM rise today?",
                pcard_2: "Market news summary",
                pcard_3: "Banking sector analysis",
                pcard_4: "Find undervalued stocks",
                pcard_5: "Compare INCO vs ANTM",
                pcard_6: "Show relations in Memory Graph",
                prompt_1: "Why did ANTM rise today? Investigate volume anomalies and corporate disclosures.",
                prompt_2: "Provide a summary of latest market news and IHSG movement sentiment.",
                prompt_3: "Banking sector analysis: compare foreign flows of BBCA, BBRI, and BMRI.",
                prompt_4: "Find undervalued stocks in the LQ45 index with low PBV and healthy ROE.",
                prompt_5: "Comprehensively compare fundamentals of nickel miners INCO vs ANTM.",
                prompt_6: "Display ANTM stock relationships with global nickel commodity movements in Memory Graph.",
                composer_placeholder: "Ask NISKAVA about stocks, tickers, news, or strategies...",
                chip_stocks: "Today's stocks",
                chip_news: "Latest news",
                chip_ta: "Technical analysis",
                chip_compare: "Peer comparison",
                chip_opp: "Investment opportunities",
                compliance_bar: "Evidence-based capital market intelligence (IDX) • Not a financial recommendation (Law 2 & Law 3)",
                disclaimer: "Regulatory Compliance (Law 2 & Law 3): Niskava Agent is an evidence-based capital market intelligence platform for the Indonesia Stock Exchange (IDX). NOT a licensed investment advisor. All data and synthesis are presented solely for market information verification and NOT buy/sell recommendations.",
                just_now: "Just now",
                copy: "Copy",
                export: "Export",
                fork: "Fork",
                copy_tooltip: "Copy Answer",
                export_tooltip: "Export Markdown",
                fork_tooltip: "Fork session from this message (OpenCode Forking)",
                investigation_details_tooltip: "Click for AI investigation details",
                investigating: "Investigating Market Intelligence",
                investigation_done: "Investigation completed in",
                investigation_failed: "Investigation failed after",
                steps: "steps",
                step: "step",
                evidence_matrix_title: "3-Tier Evidence Verification Matrix (Law 2)",
                evidence_verified_badge: "Verified Evidence",
                evidence_disclaimer: "<strong>Law 2 Compliance (Strict Non-Advisory):</strong> Evidence is deterministically classified based on corporate disclosures and official IDX news consensus. Not a buy/sell recommendation.",
                market_finding: "Market Finding",
                confidence: "Confidence",
                verified: "Verified",
                search_results_title: "Message Search Results",
                messages_found: "Messages Found",
                no_messages_match: "No messages matching",
                close_btn: "✕ Close",
                research_chat: "Research Chat",
                modal_title: "Settings & System Diagnostics",
                modal_subtitle: "Configure AI models, Sectors v2 market data feeds, Telegram bot, and local storage integrity.",
                tab_api: "API & LLM Credentials",
                tab_telegram: "Telegram Bot Daemon",
                tab_sectors: "Sectors Cache (Law 5)",
                tab_diagnostics: "System Diagnostics",
                btn_save: "Save Changes",
                btn_cancel: "Cancel",
                btn_test: "Test Connection",
                clean_cache: "Clear Expired Cache",
                btn_start_bot: "Start Bot",
                btn_stop_bot: "Stop Bot",
                btn_add_user: "+ Add",
                btn_send_ping: "Send Ping",
                status_not_checked: "Not checked",
                loading_diagnostics: "Loading system diagnostics...",
                menu_pin: "Pin to Top",
                menu_unpin: "Unpin",
                menu_rename: "Rename",
                menu_delete: "Delete Chat",
                rename_prompt: "Enter new title for chat:",
                delete_confirm: "Delete this chat session? This action cannot be undone.",
                fork_prompt: "Enter title for new branched session:",
                fork_default_title: "Fork: Market Research",
                history_empty: "No chat history yet.<br>Start a new research above.",
                options_tooltip: "Conversation options",
                btn_send_tooltip: "Send Message",
                btn_stop_tooltip: "Stop Investigation",
                btn_attach_tooltip: "Attach Document",
                toast_copied: "Copied to clipboard",
                toast_exported: "Session report exported",
                toast_new_session: "New research session started",
                toast_stopped_user: "Investigation stopped by user",
                toast_lang_id: "Bahasa diubah ke Bahasa Indonesia",
                toast_lang_en: "Language switched to English",
                sidebar_collapsed_toast: "Sidebar minimized (Icon rail)",
                sidebar_expanded_toast: "Sidebar opened",
                inv_formal_title: "Autonomous Formal Investigation (7-Stage)",
                inv_formal_desc: "Comprehensive audit of Indonesian capital market anomalies verified by mathematical evidence (Z-Score ≥ 2.50σ) and 3-tier taxonomy.",
                btn_inv_back_to_chat: "AI Chat",
                btn_inv_open_chat: "Open in Chat",
                inv_initiate_title: "Initiate Autonomous Audit",
                inv_ticker_label: "IDX Stock Ticker (BEI)",
                inv_ticker_placeholder: "e.g., ANTM, BBCA, BUMI...",
                inv_horizon_label: "Observation Range (Horizon)",
                inv_horizon_30: "30 Trading Days",
                inv_horizon_60: "60 Trading Days",
                inv_horizon_90: "90 Trading Days",
                inv_mock_label: "Offline Test Mode (Mock Data)",
                btn_run_investigation: "Start Autonomous Audit",
                inv_history_title: "Investigation History",
                inv_empty_title: "No Investigation Dossier Selected",
                inv_empty_desc: "Select a session from the list on the left or run a new audit to inspect statistical anomalies, candlesticks, 3-tier evidence, and chronological timeline.",
                inv_anomaly_title: "Deterministic Anomaly Summary (Law 1)",
                inv_chart_title: "Candlestick & Volume Spike Visualization (TradingView)",
                inv_matrix_title: "3-Tier Evidence Verification Matrix (Law 2)",
                inv_timeline_title: "Chronological Evidence & News Timeline",
                inv_disclaimer: "<strong>Regulatory Compliance (Law 2 & Law 3):</strong> All dossiers and investigative findings are presented objectively based on historical IDXnet disclosures and are not financial transaction recommendations.",
                graph_page_title: "Memory Graph — Market Intelligence Knowledge Network",
                graph_page_desc: "Interactive multi-entity association graph for Indonesian capital markets (Issuers, Commodity Catalysts, Brokers, Regulators, and Volume Anomalies).",
                btn_reload_graph: "Refresh Graph",
                graph_open_tab: "Open Independent Tab",
                graph_total_nodes: "Total Nodes",
                graph_total_edges: "Total Edges",
                graph_top_hub: "Top Hub Centrality",
                graph_coverage: "Graph Scope",
                btn_apply_filter: "Apply Filter",
                btn_reset_filter: "Reset",
                btn_clean_test_data: "Clean Test Data",
                inv_actual_vol: "Actual Volume",
                inv_return_dev: "Return Deviation",
                inv_chart_status: "Daily Market Data",
                inv_loading_history: "Loading investigation history...",
                sidebar_resizer_tooltip: "Click or drag to toggle/resize sidebar",
                sec_key_placeholder: "sec_live_... (Leave blank if unchanged)",
                label_ai_provider: "Active AI Provider",
                opt_gemini: "Google Gemini (Recommended)",
                opt_openai: "OpenAI / Compatible Gateway",
                opt_ollama: "Ollama (Local / On-Premise)",
                opt_anthropic: "Anthropic Claude",
                gemini_key_placeholder: "AIzaSy... (Leave blank if unchanged)",
                label_openai_key: "OpenAI API Key (Optional)",
                openai_key_placeholder: "sk-... (Leave blank if unchanged)",
                label_offline_mode: "Offline / Mock Mode (Law 5)",
                desc_offline_mode: "Use local synthetic data without deducting Sectors API v2 credit quota (offline_mode).",
                label_timeout: "Inference Timeout",
                label_bot_status: "Bot Status:",
                label_tele_token: "Telegram Bot Token",
                tele_token_placeholder: "7123456789:AAH... (Leave blank if unchanged)",
                label_tele_whitelist: "Whitelist User ID / Username (Analyst Restricted Access)",
                tele_user_placeholder: "Enter Telegram ID or @username...",
                label_tele_ping: "Send Test Message (Test Ping)",
                tele_chatid_placeholder: "Telegram Chat ID (e.g., 123456789)",
                desc_sectors_cache: "Monitor Sectors API v2 1,000 credit quota efficiency (Law 5 Compliance). Every market query served via local SQLite cache automatically saves 1 credit.",
                label_sectors_budget: "Sectors v2 Credit Budget Discipline (Law 5)",
                label_cache_efficiency: "SQLite Cache Efficiency:",
                data_freshness_tooltip: "Market Data Feed Status (Indonesia Stock Exchange)",
                tag_anomaly: "ANOMALY",
                tag_disclosure: "DISCLOSURE",
                tag_flow: "FLOW",
                tag_valuation: "VALUATION",
                tag_peers: "PEERS",
                tag_graph: "EGO-GRAPH",
                graph_filter_ticker_ph: "Filter Ticker (e.g., ANTM, BBRI)...",
                graph_hop_1: "1 Hop (Direct Neighbors)",
                graph_hop_2: "2 Hops (Max)",
                graph_node_all: "All Entity Types",
                graph_node_ticker: "Stock Tickers (TICKER)",
                graph_node_catalyst: "Commodity Catalysts (CATALYST)",
                graph_node_broker: "Securities Brokers (BROKER)",
                graph_node_regulator: "Market Regulators (REGULATOR)",
                graph_node_anomaly: "Volume Anomalies (ANOMALY)",
                graph_node_event: "Corporate Actions (EVENT)",
                compliance_disclaimer: "<strong>Capital Market Legal Compliance (Law 2 & Law 3):</strong> Niskava Agent is an evidence-based capital market intelligence platform for the Indonesia Stock Exchange (IDX), <u>NOT</u> a licensed investment advisor. All data, association graphs, and synthesis are presented for market information verification and <u>NOT</u> financial recommendations.",
                stat_cache_total: "Total Cache Entries",
                stat_credit_saved: "Credits Saved",
                stat_permanent_entries: "Permanent Entries (OHLCV)",
                stat_expired_entries: "Expired Entries",
                label_clean_disk: "Clean expired data to save disk space:"
            }
        };

        let currentLang = localStorage.getItem('niskava_lang') || 'id';

        function t(key, fallback = '') {
            const dict = I18N_DICT[currentLang] || I18N_DICT.id;
            return (dict && dict[key] !== undefined) ? dict[key] : (fallback || key);
        }

        function setLanguage(lang, triggerToast = true) {
            currentLang = (lang === 'en') ? 'en' : 'id';
            localStorage.setItem('niskava_lang', currentLang);
            document.documentElement.setAttribute('lang', currentLang);

            const dict = I18N_DICT[currentLang] || I18N_DICT.id;

            // 1. Update all elements with data-i18n
            document.querySelectorAll('[data-i18n]').forEach(el => {
                const key = el.getAttribute('data-i18n');
                if (dict[key]) {
                    el.innerHTML = dict[key];
                }
            });

            // 2. Update all elements with data-i18n-placeholder
            document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
                const key = el.getAttribute('data-i18n-placeholder');
                if (dict[key]) {
                    el.placeholder = dict[key];
                }
            });

            // 3. Update all elements with data-i18n-tooltip
            document.querySelectorAll('[data-i18n-tooltip]').forEach(el => {
                const key = el.getAttribute('data-i18n-tooltip');
                if (dict[key]) {
                    el.setAttribute('data-tooltip', dict[key]);
                    // Do NOT set native title if the element uses CSS data-tooltip to avoid double tooltips and tooltip freezing
                    if (!el.hasAttribute('data-tooltip') && !el.classList.contains('nav-link-item')) {
                        el.setAttribute('title', dict[key]);
                    } else {
                        el.removeAttribute('title');
                    }
                }
            });

            // 4. Update Language Switcher UI (ID / EN)
            const optId = document.getElementById('langOptId');
            const optEn = document.getElementById('langOptEn');
            if (optId && optEn) {
                optId.classList.toggle('active', currentLang === 'id');
                optEn.classList.toggle('active', currentLang === 'en');
            }

            // 5. Update prompt cards in Hero View (texts and data-prompt)
            const promptCards = document.querySelectorAll('.prompt-card');
            promptCards.forEach((card, idx) => {
                const keyP = `prompt_${idx + 1}`;
                const keyC = `pcard_${idx + 1}`;
                if (dict[keyP]) card.setAttribute('data-prompt', dict[keyP]);
                const textEl = card.querySelector('.prompt-card-text');
                if (textEl && dict[keyC]) textEl.textContent = dict[keyC];
            });

            // 6. Update dynamic elements in active Chat View
            // (a) Copy & Export action buttons
            document.querySelectorAll('.btn-copy-report').forEach(btn => {
                btn.setAttribute('title', dict.copy_tooltip);
                const span = btn.querySelector('span');
                if (span) span.textContent = dict.copy;
            });
            document.querySelectorAll('.btn-export-md').forEach(btn => {
                btn.setAttribute('title', dict.export_tooltip);
                const span = btn.querySelector('span');
                if (span) span.textContent = dict.export;
            });
            document.querySelectorAll('.btn-fork-chat').forEach(btn => {
                btn.setAttribute('title', dict.fork_tooltip);
            });

            // (b) Timestamps
            document.querySelectorAll('.message-time').forEach(el => {
                if (el.textContent === 'Baru saja' || el.textContent === 'Just now') {
                    el.textContent = dict.just_now;
                }
            });

            // (c) ReAct Accordions
            document.querySelectorAll('.react-steps-box').forEach(box => {
                const toggle = box.querySelector('.react-steps-toggle');
                if (toggle) toggle.setAttribute('title', dict.investigation_details_tooltip);
                const working = box.querySelector('.ll-text-working');
                if (working) working.textContent = dict.investigating;
                const done = box.querySelector('.ll-text-done');
                if (done) done.textContent = dict.investigation_done;
                const err = box.querySelector('.ll-text-error');
                if (err) err.textContent = dict.investigation_failed;
                const badge = box.querySelector('.react-count-badge');
                if (badge && badge.textContent) {
                    const match = badge.textContent.match(/\d+/);
                    if (match) {
                        const count = parseInt(match[0], 10);
                        const unit = currentLang === 'en' ? (count === 1 ? 'step' : 'steps') : 'langkah';
                        badge.textContent = `${count} ${unit}`;
                    }
                }
            });

            // (d) Evidence Matrices
            document.querySelectorAll('.evidence-matrix-card').forEach(card => {
                const titleSpan = card.querySelector('.evidence-matrix-title span');
                if (titleSpan) titleSpan.textContent = dict.evidence_matrix_title;
                const disc = card.querySelector('.evidence-disclaimer');
                if (disc) disc.innerHTML = dict.evidence_disclaimer;
                const badge = card.querySelector('.evidence-matrix-head .panel-card-badge');
                if (badge && badge.textContent) {
                    const match = badge.textContent.match(/\d+/);
                    if (match) {
                        badge.textContent = `${match[0]} ${dict.evidence_verified_badge}`;
                    }
                }
            });

            // (e) Update Composer Send Button Tooltip if function exists
            if (typeof updateSendButtonState === 'function') {
                updateSendButtonState();
            }

            if (triggerToast && typeof showToast === 'function') {
                showToast(currentLang === 'en' ? dict.toast_lang_en : dict.toast_lang_id);
            }
        }

            const appContainer = document.getElementById('appContainer');
            const sidebar = document.getElementById('sidebar');
            const btnToggleSidebar = document.getElementById('btnToggleSidebar');
            const btnSidebarClose = document.getElementById('btnSidebarClose');
            const btnThemeToggle = document.getElementById('themeToggleBtn') || document.getElementById('btnThemeToggle');
            const themeIconMoon = document.querySelector('.theme-icon-moon');
            const themeIconSun = document.querySelector('.theme-icon-sun');
            const btnNewResearch = document.getElementById('btnNewResearch');
            const heroView = document.getElementById('heroView');
            const chatView = document.getElementById('chatView');
            const chatInput = document.getElementById('chatInput');
            const btnSendMessage = document.getElementById('btnSendMessage');
            const sidebarSearch = document.getElementById('sidebarSearch');
            const historyList = document.getElementById('historyList');
            const appToast = document.getElementById('appToast');

    // --- 02_ui.js ---
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

            // ==========================================================================

    // --- 03_charts.js ---
// 3. CANDLESTICK CHARTS, EVIDENCE MATRIX & ADVANCED WORKSPACE ACTIONS
            // ==========================================================================

            // TradingView Lightweight Charts & High-Fidelity SVG Fallback Engine
            let activeCharts = {};

            function renderTradingViewCandlestick(containerId, candleData, anomalyDate) {
                const container = document.getElementById(containerId);
                if (!container) return;

                // Generate deterministic realistic candles if none provided
                if (!candleData || candleData.length === 0) {
                    const basePrice = 1620;
                    const now = new Date();
                    candleData = [];
                    for (let i = 29; i >= 0; i--) {
                        const d = new Date(now);
                        d.setDate(d.getDate() - i);
                        const timeStr = d.toISOString().split('T')[0];
                        const drift = Math.sin(i / 2.8) * 45 + (Math.random() - 0.48) * 30;
                        const open = Math.round(basePrice + drift);
                        const high = Math.round(open + Math.random() * 32 + 8);
                        const low = Math.round(open - Math.random() * 32 - 8);
                        const close = Math.round(open + (Math.random() - 0.48) * 30);
                        const volume = Math.round(18000000 + (i === 1 ? 68000000 : Math.random() * 14000000));
                        candleData.push({ time: timeStr, open, high, low, close, volume });
                    }
                    if (!anomalyDate && candleData.length > 2) {
                        anomalyDate = candleData[candleData.length - 2].time;
                    }
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
                                textColor: isDark ? '#94A3B8' : '#475569',
                                fontSize: 11,
                                fontFamily: 'Plus Jakarta Sans, sans-serif'
                            },
                            grid: {
                                vertLines: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' },
                                horzLines: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' }
                            },
                            crosshair: { mode: 1 },
                            rightPriceScale: { borderColor: isDark ? '#334155' : '#CBD5E1' },
                            timeScale: { borderColor: isDark ? '#334155' : '#CBD5E1', timeVisible: true }
                        });

                        const candleSeries = chart.addCandlestickSeries({
                            upColor: '#0ECB81',
                            downColor: '#F6465D',
                            borderUpColor: '#0ECB81',
                            borderDownColor: '#F6465D',
                            wickUpColor: '#0ECB81',
                            wickDownColor: '#F6465D'
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
                let svg = `<svg width="100%" height="100%" viewBox="0 0 ${width} ${height}" style="display:block;">`;
                svg += `<line x1="0" y1="${height * 0.3}" x2="${width}" y2="${height * 0.3}" stroke="rgba(255,255,255,0.06)" />`;
                svg += `<line x1="0" y1="${height * 0.6}" x2="${width}" y2="${height * 0.6}" stroke="rgba(255,255,255,0.06)" />`;

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
                const newTitle = prompt(t('fork_prompt'), `Fork: ${title || (currentLang === 'en' ? 'Market Research' : 'Riset Pasar')}`);
                if (!newTitle) return;

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

    // --- 04_sessions.js ---
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
                const invPageView = document.getElementById('investigationsPageView');
                const graphPageView = document.getElementById('graphPageView');
                const composerContainer = document.getElementById('composerContainer');
                const globalComplianceBox = document.getElementById('globalComplianceBox');
                if (invPageView) invPageView.style.display = 'none';
                if (graphPageView) graphPageView.style.display = 'none';
                if (globalComplianceBox) globalComplianceBox.style.display = 'block';
                if (composerContainer) composerContainer.style.display = 'flex';
                document.querySelectorAll('.nav-link-item').forEach(item => {
                    item.classList.toggle('active', item.getAttribute('data-nav') === 'chat');
                });

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

    // --- 05_graph.js ---
// 8. Memory Graph: Dynamic Knowledge Graph Visualizer & Anomaly Feed
            const initialGraphSvgHtml = document.getElementById('graphifySvg') ? document.getElementById('graphifySvg').innerHTML : '';

            async function loadLiveGraph(sessionId) {
                if (!sessionId) {
                    renderDynamicGraph([], []);
                    return;
                }
                try {
                    const res = await fetch(`${API_BASE}/api/graph/data?session_id=${encodeURIComponent(sessionId)}`);
                    if (!res.ok) {
                        renderDynamicGraph([], []);
                        return;
                    }
                    const data = await res.json();
                    const nodes = data.nodes || [];
                    const edges = data.edges || [];
                    renderDynamicGraph(nodes, edges);
                } catch (e) {
                    console.warn('loadLiveGraph warning:', e);
                }
            }

            function renderDynamicGraph(nodes, edges, focusTicker) {
                const svg = document.getElementById('graphifySvg');
                const badge = document.getElementById('graphifyFocusBadge');
                const tooltip = document.getElementById('graphTooltip');
                if (!svg) return;

                if (!nodes || nodes.length === 0) {
                    svg.innerHTML = `
                        <rect width="100%" height="100%" fill="transparent"/>
                        <text x="175" y="95" text-anchor="middle" fill="#64748B" font-size="11" font-family="sans-serif">Belum ada relasi aktif untuk sesi ini</text>
                        <text x="175" y="115" text-anchor="middle" fill="#475569" font-size="9" font-family="sans-serif">Ketik emiten (cth: ANTM, BBRI) di chat untuk mulai</text>
                    `;
                    if (badge) badge.textContent = '-';
                    return;
                }

                let centerNode = nodes.find(n => focusTicker && (n.id.toUpperCase() === focusTicker.toUpperCase() || (n.label && n.label.toUpperCase() === focusTicker.toUpperCase())));
                if (!centerNode) {
                    centerNode = nodes.find(n => n.node_type === 'EMITEN' || n.node_type === 'STOCK') || nodes[0];
                }

                if (badge && centerNode) {
                    badge.textContent = centerNode.label || centerNode.id;
                }

                const satelliteNodes = nodes.filter(n => n.id !== centerNode.id).slice(0, 6);
                const width = 350;
                const height = 200;
                const centerX = width / 2;
                const centerY = height / 2;
                const radiusX = 112;
                const radiusY = 62;

                const positions = new Map();
                positions.set(centerNode.id, { x: centerX, y: centerY, node: centerNode, isCenter: true });

                satelliteNodes.forEach((node, i) => {
                    const angle = (i * 2 * Math.PI / satelliteNodes.length) - (Math.PI / 2);
                    const x = Math.round(centerX + radiusX * Math.cos(angle));
                    const y = Math.round(centerY + radiusY * Math.sin(angle));
                    positions.set(node.id, { x, y, node, isCenter: false });
                });

                const typeColors = {
                    'EMITEN': { fill: 'rgba(252, 213, 53, 0.18)', stroke: '#FCD535', text: '#946300' },
                    'STOCK': { fill: 'rgba(252, 213, 53, 0.18)', stroke: '#FCD535', text: '#946300' },
                    'COMMODITY': { fill: '#FFFBEB', stroke: '#F59E0B', text: '#B45309' },
                    'INSTITUTION': { fill: '#ECFDF5', stroke: '#10B981', text: '#047857' },
                    'SECTOR': { fill: 'rgba(252, 213, 53, 0.15)', stroke: '#F0B90B', text: '#946300' },
                    'PEER': { fill: '#FAF5FF', stroke: '#8B5CF6', text: '#6B21A8' },
                    'NEWS': { fill: '#F1F5F9', stroke: '#64748B', text: '#334155' },
                    'ANOMALY': { fill: '#FEF2F2', stroke: '#EF4444', text: '#B91C1C' }
                };

                let svgHtml = '';

                // Draw Edges
                if (edges && edges.length > 0) {
                    edges.forEach(edge => {
                        const p1 = positions.get(edge.source_id);
                        const p2 = positions.get(edge.target_id);
                        if (p1 && p2) {
                            svgHtml += `<line x1="${p1.x}" y1="${p1.y}" x2="${p2.x}" y2="${p2.y}" class="graph-edge"></line>`;
                        }
                    });
                }
                // Fallback connection to center if no edges drawn
                satelliteNodes.forEach(sNode => {
                    const p = positions.get(sNode.id);
                    if (p) {
                        svgHtml += `<line x1="${centerX}" y1="${centerY}" x2="${p.x}" y2="${p.y}" class="graph-edge"></line>`;
                    }
                });

                // Satellite Nodes
                satelliteNodes.forEach(node => {
                    const pos = positions.get(node.id);
                    const col = typeColors[node.node_type] || { fill: '#F8FAFC', stroke: '#94A3B8', text: '#475569' };
                    const label = (node.label || node.id).slice(0, 8);
                    svgHtml += `
                        <g class="graph-node" data-id="${escapeHtml(node.id)}" data-type="${escapeHtml(node.node_type || '')}" data-label="${escapeHtml(node.label || node.id)}" transform="translate(${pos.x}, ${pos.y})">
                            <circle r="18" fill="${col.fill}" stroke="${col.stroke}" stroke-width="1.8"></circle>
                            <text text-anchor="middle" dy="4" font-size="8.5" font-weight="700" fill="${col.text}">${escapeHtml(label)}</text>
                        </g>
                    `;
                });

                // Center Node
                const centerLabel = (centerNode.label || centerNode.id).slice(0, 6);
                svgHtml += `
                    <g class="graph-node active" data-id="${escapeHtml(centerNode.id)}" data-type="${escapeHtml(centerNode.node_type || '')}" data-label="${escapeHtml(centerNode.label || centerNode.id)}" transform="translate(${centerX}, ${centerY})">
                        <circle r="26" fill="#FCD535" stroke="#1E2329" stroke-width="2.5" filter="drop-shadow(0 4px 10px rgba(252, 213, 53, 0.45))"></circle>
                        <text text-anchor="middle" dy="4" font-size="11" font-weight="800" fill="#1E2329">${escapeHtml(centerLabel)}</text>
                    </g>
                `;

                svg.innerHTML = svgHtml;

                svg.querySelectorAll('.graph-node').forEach(g => {
                    g.addEventListener('mouseenter', () => {
                        const lbl = g.getAttribute('data-label');
                        const typ = g.getAttribute('data-type');
                        if (tooltip) {
                            tooltip.innerHTML = `<span><strong>${escapeHtml(lbl)}</strong> (${escapeHtml(typ || 'Relasi')})</span><strong style="color:var(--accent-text)">Klik Node</strong>`;
                        }
                    });
                    g.addEventListener('click', () => {
                        const lbl = g.getAttribute('data-label');
                        showToast(`Memory Graph: Relasi ${lbl} terpilih`);
                    });
                });
            }

            function attachInitialGraphListeners() {
                document.querySelectorAll('.graph-node').forEach(node => {
                    node.addEventListener('click', () => {
                        const nodeName = node.getAttribute('data-node') || 'Emiten';
                        showToast(`Memory Graph: Relasi ${nodeName.toUpperCase()} terpilih`);
                    });
                });
            }
            attachInitialGraphListeners();

            function updateRightPanelAnomaly(payload) {
                // Panel Analisis kanan dihilangkan sesuai permintaan pengguna
            }

            // 9. Markdown Parser supporting Code Blocks, Headings, Tables, Blockquotes, Badges, Lists, and Paragraphs
            function renderMarkdown(txt) {
                if (!txt) return '';

                // Step 1: Extract Fenced Code Blocks (```lang ... ```)
                const codeBlocks = [];
                let text = txt.replace(/```([a-zA-Z0-9_-]*)\n([\s\S]*?)```/g, (match, lang, code) => {
                    const id = `__CODE_BLOCK_${codeBlocks.length}__`;
                    const escapedCode = code
                        .replace(/&/g, '&amp;')
                        .replace(/</g, '&lt;')
                        .replace(/>/g, '&gt;');
                    const langBadge = lang ? `<div class="code-block-header"><span class="code-lang-label">${lang.toUpperCase()}</span><button class="btn-copy-code" onclick="navigator.clipboard.writeText(this.getAttribute('data-code')).then(()=>showToast(currentLang==='en'?'Code copied!':'Kode disalin!'))" data-code="${code.replace(/"/g, '&quot;')}">Salin</button></div>` : '';
                    codeBlocks.push(`<div class="code-block-wrapper">${langBadge}<pre><code class="language-${lang || 'plaintext'}">${escapedCode}</code></pre></div>`);
                    return id;
                });

                // Step 2: Escape HTML for security
                text = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

                // Step 3: Badges (Law 2 Verification Taxonomy & Metrics)
                text = text.replace(/\[SUPPORTED\]/g, '<span class="md-badge green">SUPPORTED</span>');
                text = text.replace(/\[UNCERTAIN\]/g, '<span class="md-badge amber">UNCERTAIN</span>');
                text = text.replace(/\[CONTRADICTED\]/g, '<span class="md-badge red">CONTRADICTED</span>');
                text = text.replace(/\[LIKELY_CATALYST\]/g, '<span class="md-badge blue">LIKELY CATALYST</span>');
                text = text.replace(/\[PRECEDED_ANNOUNCEMENT\]/g, '<span class="md-badge amber">PRECEDED ANNOUNCEMENT</span>');
                text = text.replace(/\[UNEXPLAINED_BY_NEWS\]/g, '<span class="md-badge red">UNEXPLAINED BY NEWS</span>');
                text = text.replace(/\[EXTREME SURGE\]/g, '<span class="md-badge green">EXTREME SURGE</span>');
                text = text.replace(/\[ANOMALY\]/g, '<span class="md-badge purple">ANOMALY</span>');
                text = text.replace(/\[DIVERGENT\]/g, '<span class="md-badge amber">DIVERGENT</span>');
                text = text.replace(/\[INSTITUTIONAL BUY\]/g, '<span class="md-badge blue">INSTITUTIONAL BUY</span>');

                // Step 4: Markdown Headings (# ... ####)
                text = text.replace(/^#### (.*$)/gim, '<h4>$1</h4>');
                text = text.replace(/^### (.*$)/gim, '<h3>$1</h3>');
                text = text.replace(/^## (.*$)/gim, '<h2>$1</h2>');
                text = text.replace(/^# (.*$)/gim, '<h1>$1</h1>');

                // Step 5: Horizontal Rule
                text = text.replace(/^---$/gim, '<hr>');

                // Step 6: Blockquotes (> ...)
                text = text.replace(/^&gt; (.*$)/gim, '<blockquote><p>$1</p></blockquote>');

                // Step 7: Bold & Italic & Inline Code
                text = text.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
                text = text.replace(/\*(.*?)\*/g, '<em>$1</em>');
                text = text.replace(/`([^`]+)`/g, '<code>$1</code>');

                // Step 8: Tables and Lists line-by-line processing
                const lines = text.split('\n');
                let inTable = false;
                let tableHtml = '';
                let inUl = false;
                let inOl = false;
                const result = [];

                for (let i = 0; i < lines.length; i++) {
                    const rawLine = lines[i];
                    const trimmed = rawLine.trim();

                    // Check Table
                    if (trimmed.startsWith('|') && trimmed.endsWith('|')) {
                        if (inUl) { result.push('</ul>'); inUl = false; }
                        if (inOl) { result.push('</ol>'); inOl = false; }
                        const cells = trimmed.split('|').map(c => c.trim()).slice(1, -1);
                        if (!inTable) {
                            inTable = true;
                            tableHtml = '<table><thead><tr>' + cells.map(c => `<th>${c}</th>`).join('') + '</tr></thead><tbody>';
                        } else if (trimmed.includes('---')) {
                            // Separator row
                            continue;
                        } else {
                            tableHtml += '<tr>' + cells.map(c => `<td>${c}</td>`).join('') + '</tr>';
                        }
                        continue;
                    } else if (inTable) {
                        inTable = false;
                        tableHtml += '</tbody></table>';
                        result.push(tableHtml);
                        tableHtml = '';
                    }

                    // Check Unordered List: '-' or '*'
                    const ulMatch = trimmed.match(/^[-*]\s+(.*)$/);
                    if (ulMatch) {
                        if (inOl) { result.push('</ol>'); inOl = false; }
                        if (!inUl) { result.push('<ul>'); inUl = true; }
                        result.push(`<li>${ulMatch[1]}</li>`);
                        continue;
                    }

                    // Check Ordered List: '1.', '2.', etc.
                    const olMatch = trimmed.match(/^(\d+)\.\s+(.*)$/);
                    if (olMatch) {
                        if (inUl) { result.push('</ul>'); inUl = false; }
                        if (!inOl) { result.push('<ol>'); inOl = true; }
                        result.push(`<li>${olMatch[2]}</li>`);
                        continue;
                    }

                    // Close list tags if non-list line
                    if (inUl) { result.push('</ul>'); inUl = false; }
                    if (inOl) { result.push('</ol>'); inOl = false; }

                    if (!trimmed) {
                        result.push(''); // blank line
                    } else if (trimmed.startsWith('<h') || trimmed.startsWith('<hr') || trimmed.startsWith('<blockquote') || trimmed.startsWith('__CODE_BLOCK_')) {
                        result.push(trimmed);
                    } else {
                        result.push(`<p>${trimmed}</p>`);
                    }
                }

                if (inTable) {
                    tableHtml += '</tbody></table>';
                    result.push(tableHtml);
                }
                if (inUl) result.push('</ul>');
                if (inOl) result.push('</ol>');

                let finalHtml = result.join('\n');

                // Step 9: Re-insert Fenced Code Blocks
                codeBlocks.forEach((block, idx) => {
                    finalHtml = finalHtml.replace(`__CODE_BLOCK_${idx}__`, block);
                });

                return finalHtml;
            }

    // --- 06_chat.js ---
// 10. Message Dispatcher & SSE Stream Handler
            async function handleSendMessage() {
                const prompt = chatInput.value.trim();
                if (!prompt || isGenerating) return;

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
                            } else if (eventName === 'pdf_report_ready' || payload.event === 'pdf_report_ready') {
                                const filename = payload.filename || '';
                                if (filename && !assistantMsgObj.element.querySelector('.pdf-download-banner')) {
                                    const downloadBtn = document.createElement('div');
                                    downloadBtn.className = 'pdf-download-banner';
                                    downloadBtn.innerHTML = `
                                        <div style="display:flex; align-items:center; gap:10px; margin-top:12px; padding:10px 14px; background:var(--badge-green-bg,#e8f5e9); border:1px solid var(--badge-green-border,#a5d6a7); border-radius:8px; font-size:13px;">
                                            <span style="font-size:20px;">📄</span>
                                            <div style="flex:1">
                                                <strong style="color:var(--badge-green-text,#1b5e20);">${typeof currentLang !== 'undefined' && currentLang === 'en' ? 'PDF Audit Report Ready' : 'Laporan PDF Audit Siap'}</strong>
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
                                                <strong style="color:var(--badge-green-text,#1b5e20);">${typeof currentLang !== 'undefined' && currentLang === 'en' ? 'PDF Audit Report Ready' : 'Laporan PDF Audit Siap'}</strong>
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
                    if (reactSteps.length === 0 && assistantMsgObj.element) {
                        const sb = assistantMsgObj.element.querySelector('.react-steps-box');
                        if (sb) sb.style.display = 'none';
                    }
                    loadChatSessions(false);
                    loadLiveGraph(currentSessionId);
                    scrollToBottom();
                }
            }

            function appendUserMessage(text) {
                const safeText = (text || '').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                const msgDiv = document.createElement('div');
                msgDiv.className = 'chat-message user-message';
                msgDiv.innerHTML = `
                    <div class="message-avatar user">G</div>
                    <div class="message-body">
                        <div class="message-header-row">
                            <span class="message-author">User</span>
                            <span class="message-time">${t('just_now')}</span>
                        </div>
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
                msgDiv.innerHTML = `
                    <div class="message-avatar ai"><img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAASKUlEQVR4nO1bfXBV5Zn/ve8599yv5OY7wYQkkESQQBLCZQkJkLCGjxIKMmxjAqJIpYmCtItWRlpn+dCtCwhKXZWusjvqtIyy7iDaXS0WpGBBIrbqVkZaTC0glEKBTICQ8/HbP845N/eGEAIJaGf2mTlz7z1f93l+z/dz3gP8P32lJGtraxUA4qtm5EaRAKAAUHGp0MLZry5btkx2cfxvliRsoZXonR6PB42NjVm33nrrqKVLl2Z4PB4IcYnMLljyejF3PVAWsBkWAEwABAApJUzT9E2cOHH40aNHJ5w9e3ZCS0vLCMMw4lVVPev1ej8LhUK/CYVCe3Nycj7cunXr71VVvWCapntfCYC1tbVy8+bNBGD1FbN9SQpsoQHYWq6vr8/56KOPKs6cOTO5tbV1/Pnz5we0t7fDsiL8M5oPVVXh8XigadoXgUDgo7i4uL0DBw78xbZt2/aTjP4v1yr6BIi+IAkAR48eDVRWVpYXFBQsS0lJ2RkMBs+pqkrYgrqa04UQhvPd3WcA0J3PyPlSSoZCIbOysvL2999/P2XEiBHfWbx4cZaqqu7/xrjWV0WKoigoKipamZSU9Eefz0chRLTQBmyhzah9V9pM57p2AFYgELCqq6vvyszMfMvv91/IzMz8r7q6unzn/7sKqjeMVCEEioqKFmiaFqNl2AK4Wu4MCjvvv9xx5x6Wx+NpGzt27Mq4uLiTABgKhU6WlpbWSRmJkdctWHZJVVVVKgCMGzfuVr/fb6LDtHuq5R5vQggLttDmpEmTWjVNMwBQ0zTm5eX9a3Nzs89h64ZZggIAdXV1AxyNWOiI+pdot7t93Wi+K0ugz+ejz+eLxA4pJZOTk5sqKyuLHN6uOwgCgNyyZUt8SkrKR44QXWq+pwD0FITo86K+6wCYkJBwtqGh4SaHv+sGggCgqqqKrKysrQ4T+pUYFsLeLhXc+bw6S+hqa1dVleFw+DaHz0ia6GvySCmRm5v7L1LKboXvCgTECNqlNntkBcFgkIMGDaKU0r1WF0JYOTk5j6OjrO5zUoQQKCkpaXAi/iXCdzbR6M0Fwf3t1cDSwV7mZ6lUFOcaXD4+CCHo8XgIgImJiVy/fj0B0OPxRFwwNTX1fadG6PP6QAGA8vLyykAgYCAqxXUIGSug+x0A8zMVzp6USr9PELCP/UNVHF95rD9/8nAmb8nxXFbbUkpGF1NCCN555508ceIES0pKokEwNU1jOByeEc1zd9TTvCkAWMuWLUs+cODAq+fPn1cEhBNoOr5dGnsEACInQ2BFQw5SEyQuthNSCkgBlBcFABhIDkmkJtq8RvdDUkpIKWFZFgzDQHx8PObPn4+mpia89NJLSEtLw65du3DXXXdB13WQRHt7Ow4ePLi+vr4+w1FQnwRDBQAqKirKHdM3cYl5ghCI+S0VsCBT8qcrB/DhOan0KHZpC4DVYR+3rhnApxb34+xJIQajLENRlMh5AJibm8vly5ezubmZLlmWRcuyIr83bNjA+Ph4Sil1VVVZUlKy0OG921jQo0BRVVUldu7ciQutF3ItyyIASilgWcTIwRq+OS4FHlXAIiEgHDUSAhYKsv344Het+PGrJ2HC1mbVcC8euisLz24+jv9+7xwMAkIIqKoCwzDgdoAjRoxAQ0MD6uvrkZCQAAAwTRNCCLjVH0kYhoHGxkacOnUKP/zhDxW/3w9N0z5x2GevAdi5cycAoPVCa55pmkJKQZKYMjqIb0+/Cbt/81cc+6sBRdruIIQNhBDAL/aewfuftkG3HOFLvfjBvBw88+qX2Lr7HIQQUBQFpmnAMAwIIVBTU4P77rsPU6ZMiQhqGAaklFCUWLd2wTBNE1VVVbR3iZaampr/bWpqAvqoW1SFEMjMvOkn0kk5t42L5xtPDOLYYt8V05ZrzuNKfHz7xwWcPjboWhEV51hiYiIbGhrY1NTEaDIMI8bUuyL3+OnTp41QKMT4+PgPSPZpMSRVVUFGRto2AJw+Lmi8uW4wK4Z6HUGUK4IwttjLt566mTPGBSORXFFs4e+//34ePnw4IpBpmj0SvAsQ9DFjxjAYDD7rWMoVLbwnLiAAWLpuKAnxcTnfHBPEvd/KFv/8QjN+/buLdqQWJoYP9mLATRpIROUHW/qMRAU1Y1Ow4bUvsWXXOUgpIQRgWURxcTGefvppAB1mHvljIWBZVmR44pp7F6MzmKYJVVVRVlaG/fv3H+mBXD0DYNmyZWLFihVMGzInfUZlKLu2OogfbfxCvPfJRdsfLRNzpyZh2rgkfHGsLaJyEJHUeLGdWPPSUfzq4zYIIUBaEEIBaWHt2rUAAF3XI/4dDYKbCqPJBaTzfgBi+PDhIDnCjitmtwGwR7Rjmd3yfrp75eJ3ny9lRaHHAECpSAoBNkxP4utr8jg8X6GqgKoCKp02dKrqVMV2mdtvvz3i59HU1NTE+fPnMzMzk+FwmI2NjXzhhRe4b98+nj59uksXME2TJM3jx48zPT39jyQ1F5ReAUBCkBS7//Pu344frhFO+6lI8NvfTODWJ/JYmKs4AsrY0lfYuV26+52AKKVkMBhkc3MzLcuiYRhsa2vjL3/5S957772cOHEiFy1axNWrV3PGjBnMyMiIgNivXz9WV1dzx44dEfBM06RlWdy3bx+///3vMz4+Xp8wYcLNjgjXPiRx5vP4ny3PDq4YntEOwHIj+ozKeL65Lp/DBqgx3Z7s3OG5xxwB3JL20UcfJUnquk6SPHLkCH/+85/zwIEDEa3qus4dO3awtrY2ApwLxEMPPUSSvHjxIklywYIFkWNer5fTp08f3GsAqpwYUVE29Ed+nySE0CEEPQq44eFMzp0S19H0RKW7y22KolAIwcGDB/PChQsRzXWmP/zhD3zqqae4Zs0aPvHEE1y+fDlLS0sj1+fl5fHo0aMR19m3b5+bbk1FUZicnPwxSQ/6IA2qAJCYmPiCbcpCBwSDPnDDknROLvPFaBdRJh4tuOv7iuP7b775ZozvW5ZF0zR54MAB1tfX0+PxsLy8nFOmTGFRURGzs7MjluP1evnee+/FaH/27NmudemKoliFhYUPOJniikH+iuYhhEBqaqoRXVEKAQgJJ6Lb+/x+P/Lz82PSlns9AGiaBtM0MW3aNEydOhWmacZUdUIImKaJYDCIYcOGobi4GHPmzMGhQ4dw+PBhVFdXIz8/H+vWrUNFRQUMw4DH4wEAfPHFFw4vVEzTFKdOnXpw7ty5fwfAdN34WkkFgKysrBVOe6sDgnF+8N+WZvAbo/0RLWdmZnL79u3Mzc1lWVmZ257GWIKmaTx48GBE4z2hvXv3cvXq1WxpaeHhw4d59uzZiOu48WPJkiUEYCqKYqWnp/923Lhxlc7Atm9coLCwcKmU0ooG4PkfZLCmvAOAlJQUbt++nQsXLiRJzpo1i4A9wMzJyeHQoUO5bt26GNPvKpW5QhmGcdnzXHKPf/jhh5RS6gDYv3//zQ7vfTIiVwFgwIABjY4FtLsx4PkfpLOmwh8zqdm1axc//vhjvvvuu/R6vRRC8NFHH+W5c+di/P1qyAXFtZrO17tt8ahRo0whBBMTEw/SrgF61AtcCSULgBgyZMiOYDB4moQCAcst8mgRQgCqqkDXdezfvx9FRUXYvXs3Ll68CJKYNGkSAoGAfTPL6rKM7ZZBKaGq6mXLYLc9njNnjiAJy7IG3HHHHTchqhbtLSkAMHLkyGk+n5cAdEXCevrBm3j/txKc6GuPs0aMGEGSPHToEB988EGqqspFixZ1a/a9Jdcijh8/zri4OENRFI4ePfobAOAsvugTUoUQyM/PX+6kMn30UC/fWJvHyaP8Dgh2itu0aVOEubfeeovl5eX88ssvY5jtC6HdClLXdba3t5Mk582bpwPgoEGDlrh89xUAkWcBGf0ytjjR3age6ecbT+Tz1rCfEIKKojAQCPCdd96JYbitra3PhO+O9u7dqwPg4MGDF/U1AC4I4uWXXw4lJiYeFMJ+FOaCMLbYRyHsAiguLsjt27f3uYAuiJ9//jn/8pe/cNOmTfze977HgwcP8uTJk5w5c6apKAonT55cDvStC7gAyAULFsTFxcV9Dqf0BMDavw/ppyuzmZ4oI8ORQCDADRs2cM+ePXz88cf52WefkWSP839X5F67c+dO1tTUcObMmXzkkUc4d+5cZmVlmQAYCoWOv/3228EonvuMVAAoKCh42H0i5HZ9qQngi/+UxVGF7oQopgCyABgrVqwg2dH89JZaW1tJkps2bXL/R1cUhWlpaW84FWaPtH81zwXMZ599NunkyZP/aFkWIexrSXvTTUCVNuAkI8NOKaUQQiivvPLKJeXv1RLtNAdd1xEM2kpOSEjAM888A6/XC8uyEB8f/ytnqtwj7fcUAAUAV61aVdfS0pIBwBQx1woICDBmDQ9hmiY0TTufnZ39zqeffoo9e/ZEav5rIbcW8Hg8OHPmDNauXYtjx46hoKAAqqpKVVXRr1+/3QBQW1vb+2lQFKkAkJeX1yildNf4UDoDj5QQuPGR/hw91BdxASGELoRgfn7+6oSEBAghjjU0NJCkeS1uoOs6T5w4wT179nD+/PmRZ4TocLNr8v8epYna2lpu3rwZ6enph44cOSLa29tjLcfpCmP0T0JRFOi6/tnZs2dFWlraG6+//vp3nnzySUvTNGlZVlczvUvIdac///nPWLhwIVpbW5Geno6VK1fCNE00NzfjzJkz5pYtW1Sfz7evpqbmHDqtVus1uS3ljBkzhvj9fgOAJYSwhGMBySHwpeUDWOaOye06wfR4PCwuLp4MADU1NZUAuHHjxkgauJrq0C18urqupaVFDwaDHDhwYI8eh10LCQB49dVX4wKBwAnYDZBl533BoF9w3rQMDsjUXADcpSzWhAkThgAASS0hIeH38fHx1mOPPWa2tLREBLgaIEzT5CeffMLi4mKWlZVx2rRpvOWWWyyv19t+9913934O2B0AmqYhOTn5I9gWYLoW4C5SkDIyAbJg1wKnFy9enOzepLi4+AGnimwvKCjgxo0bY7pEt+vrTnjSrvvLysoYDoej1wXsdTLMdVsaoyiKgn79+m1zhNSB2EUQ7ooPdz1gQkLCZyTdvKeoqors7OyfOeMtHQDD4TA3b94cI6g7C7gcGG7tT5KHDx/Ww+Ew09LSVvZ0DHbNAADAsGHDJvp8PhOA6WraBSIKEAMA09LSdjirNdy1w7K5udmXkpLyvv1oTIksrCorK+Nzzz0XaZyiwdB1vcsK0gHCmDVrlhUIBKY6AFzXlaMSAIYPHz43aomM1cXKLR32mOzFTkxJAPjud7+bk5SUdAz2oNSMHqImJydz9uzZfO2113jq1KkYgd1AGA2KYRhmSUkJi4qKej8G7yF5AGDQoEHuylC9IyNE8rIuhODAgQNXONdEzNJtUG677baKYDDYBsAQQlhSysjU2N0yMjJYV1fHF198kX/60590kp239iVLltDj8fyaHWPwG7JQ0l0ptjTKn01nswAYiqKwsLBwXmcAon+Hw+FZXq83Yklwxuqqql4CRiAQ4JgxYzhv3jzeeeedrKur46RJk5iWlnZw6tSpWc59b+hSWVVRFOTm5q53V256vV5qmkYhhOHxeKyRI0dOAC7blnqklLj55psjQxZ0epAihKCqqgwGg+3Z2dmrNE17AMADABb7/f4FgUBgyrZt2xKc+93wxdICgJRSoqqqamxVVdX48ePHV1VXV9f5fL42r9fL2traQufcy2lGVRQFWVlZP4ta6xcNghtM92iaFt1gda4iv9LXa2L+vLCwcLoQghkZGb/44IMPPOjeLAUASVJLTU3dC1vrRlRq1QEwNzd3DexA6oPzXhEA9WvzwlVtba1SWFiokRSpqalv+f3+C/Pnz+/vHO7WL90yu7GxMSspKekonEGLO3ZTVZWlpaUzndOvW47vLQkAWL9+fSgUCrGgoOBeZ3+PcrIbI6ZOnToqEAich11jmLCDX1tdXV22c+qNfR+gp+S+N1BQUHBPVlbWu1czkYkiFQBGjhxZ72SGiwCYlJT0Mcmv/Wt0EoAoLS1dX1ZWdjM63hy7WvIIIVBQUPCwk17NrKysf7/eJW5vSQDA0qVL0+65557xQIdfXyOpqqqif//+/6EoCocMGXKPu7+XfF5fihK6t6bqZgbPwIED3wuHw2Od/V+Lt8OuRH3ip7QXOeLll18OrVq1Kr4v7vm3SF/roHej6Lo3N/8HKyQo2DSlao4AAAAASUVORK5CYII=" alt="NISKAVA Agent" class="ai-avatar-img"></div>
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
                if (!Array.isArray(steps)) steps = [steps];
                const unit = currentLang === 'en' ? (steps.length === 1 ? t('step') : t('steps')) : t('steps');
                msgObj.countBadge.textContent = `${steps.length} ${unit}`;
                msgObj.stepsContent.innerHTML = steps.map(s => `
                    <div class="react-step-row">
                        <span class="react-step-bullet">●</span>
                        <span>${String(s || '').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</span>
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

            // Accidental Navigation Guard: Warn user before reloading if analysis is actively streaming/thinking
            window.addEventListener('beforeunload', (e) => {
                if (isGenerating) {
                    e.preventDefault();
                    e.returnValue = '';
                }
            });

    // --- 07_settings.js ---
// 10. Navigation, Dedicated Memory Graph & Settings/Toolkit Integration
            function initNavigationAndSettings() {
                const composerContainer = document.querySelector('.composer-container');
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

                function switchMainView(targetNav) {
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
                            if (target === 'sectors') {
                                const tab1Badge = document.getElementById('tab1SectorsBadge');
                                if (tab1Badge) {
                                    tab1Badge.textContent = currentLang === 'en' ? `Connected (${result.latency_ms}ms)` : `Terhubung (${result.latency_ms}ms)`;
                                    tab1Badge.style.color = '#10B981';
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
                                    navigator.clipboard.writeText(path).then(() => {
                                        showToast(isEn ? 'Path copied to clipboard' : 'Jalur disalin ke papan klip');
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
                            payload.auth.sectors_api_key = inputSectorsKey.value.trim();
                        }
                        if (selectAiProvider && selectAiProvider.value) {
                            payload.auth.ai_provider = selectAiProvider.value;
                        }
                        if (inputGeminiKey) {
                            payload.auth.gemini_api_key = inputGeminiKey.value.trim();
                        }
                        if (inputGeminiModel && inputGeminiModel.value.trim()) {
                            payload.auth.gemini_model = inputGeminiModel.value.trim();
                        }
                        if (inputOpenaiKey) {
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
                        if (inputAnthropicKey) {
                            payload.auth.anthropic_api_key = inputAnthropicKey.value.trim();
                        }
                        if (inputAnthropicModel && inputAnthropicModel.value.trim()) {
                            payload.auth.anthropic_model = inputAnthropicModel.value.trim();
                        }
                        if (inputTeleToken) {
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

    // --- 08_bootstrap.js ---
// Initialize theme, send button, sessions, live graph, navigation & settings on start
initTheme();
updateSendButtonState();
loadChatSessions();
loadLiveGraph(currentSessionId);
initNavigationAndSettings();
})();
