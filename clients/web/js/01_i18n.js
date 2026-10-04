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
                nav_radar_audit: "Audit & Anomali",
                radar_title: "Pemindai Anomali & Audit Saham",
                radar_badge: "Bursa Efek Indonesia",
                radar_desc: "Pemindaian otomatis lonjakan volume perdagangan (Vz ≥ 2.50σ), foreign flow, dan cross-check keterbukaan informasi IDXnet.",
                radar_top_anomalies: "Sinyal Anomali Terdeteksi (Top Z-Score)",
                radar_empty: "Belum ada anomali terdeteksi melebihi ambang batas 2.0σ. Masukkan kode emiten di bawah untuk memeriksa.",
                radar_btn_inspect: "Audit Saham Ini",
                radar_search_title: "Pemeriksaan Saham",
                radar_search_placeholder: "Ketik kode emiten BEI (cth: ANTM, BBCA, BUMI)...",
                btn_refresh_data: "Perbarui Data",
                btn_inv_open_chat: "Buka di Percakapan AI",
                btn_inv_export_md: "Ekspor Laporan Markdown",
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
                btn_run_investigation: "Mulai Audit Otonom",
                inv_history_title: "Riwayat Investigasi",
                inv_empty_title: "Dossier Investigasi Belum Dipilih",
                inv_empty_desc: "Pilih salah satu sesi dari daftar di sebelah kiri atau jalankan audit baru untuk menelaah anomali statistik, candlestick, bukti 3-tier, dan kronologi kejadian.",
                inv_anomaly_title: "Ringkasan Anomali Deterministik (Law 1)",
                inv_chart_title: "Visualisasi Candlestick & Volume Spike (TradingView)",
                inv_matrix_title: "Matriks Verifikasi Bukti 3-Tier (Law 2)",
                inv_timeline_title: "Timeline Kronologis Bukti & Berita",
                inv_disclaimer: "<strong>Kepatuhan Regulasi (Law 2 & Law 3):</strong> Seluruh dossier dan temuan investigasi disajikan secara obyektif berdasarkan data historis keterbukaan informasi IDXnet dan bukan merupakan rekomendasi transaksi finansial.",
                graph_page_title: "Memory Graph - Jaringan Asosiasi Pasar Modal",
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
                nav_radar_audit: "Audit & Anomaly Screener",
                radar_title: "Market Anomaly Screener & Audit",
                radar_badge: "Indonesia Stock Exchange",
                radar_desc: "Automated screening for volume spikes (Vz ≥ 2.50σ), foreign flow, and IDXnet corporate disclosure cross-referencing.",
                radar_top_anomalies: "Detected Anomaly Signals (Top Z-Scores)",
                radar_empty: "No anomalies above 2.0σ threshold yet. Enter an IDX ticker below to inspect.",
                radar_btn_inspect: "Audit Stock",
                radar_search_title: "Stock Audit Inspection",
                radar_search_placeholder: "Enter IDX ticker (e.g. ANTM, BBCA, BUMI)...",
                btn_refresh_data: "Refresh Data",
                btn_inv_open_chat: "Open in AI Chat",
                btn_inv_export_md: "Export Markdown Report",
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
                btn_run_investigation: "Start Autonomous Audit",
                inv_history_title: "Investigation History",
                inv_empty_title: "No Investigation Dossier Selected",
                inv_empty_desc: "Select a session from the list on the left or run a new audit to inspect statistical anomalies, candlesticks, 3-tier evidence, and chronological timeline.",
                inv_anomaly_title: "Deterministic Anomaly Summary (Law 1)",
                inv_chart_title: "Candlestick & Volume Spike Visualization (TradingView)",
                inv_matrix_title: "3-Tier Evidence Verification Matrix (Law 2)",
                inv_timeline_title: "Chronological Evidence & News Timeline",
                inv_disclaimer: "<strong>Regulatory Compliance (Law 2 & Law 3):</strong> All dossiers and investigative findings are presented objectively based on historical IDXnet disclosures and are not financial transaction recommendations.",
                graph_page_title: "Memory Graph - Market Intelligence Knowledge Network",
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
