// NISKAVA Web Workspace Client Architecture
        (function() {
            // State
            const API_BASE = (window.location.protocol === 'file:' || ['5500', '3000', '5173', '8000'].includes(window.location.port))
                ? 'http://127.0.0.1:20128'
                : '';
            let currentSessionId = 'WEB-' + new Date().toISOString().slice(0,10).replace(/-/g,'') + '-' + Math.floor(1000 + Math.random() * 9000);
            let isGenerating = false;
            let currentAbortController = null;

            // DOM Elements

        // ==========================================
