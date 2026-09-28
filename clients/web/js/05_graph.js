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

            // 9. Markdown Parser supporting Headings, Tables, Blockquotes, Badges, Lists
            function renderMarkdown(txt) {
                if (!txt) return '';

                let html = txt;

                // Escape HTML tags to prevent XSS except formatted blocks
                html = html.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

                // Restore markdown blockquotes (> ...)
                html = html.replace(/^&gt; (.*$)/gim, '<blockquote><p>$1</p></blockquote>');

                // Markdown Headings: #, ##, ###, ####
                html = html.replace(/^#### (.*$)/gim, '<h4>$1</h4>');
                html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
                html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
                html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');

                // Horizontal Rule
                html = html.replace(/^---$/gim, '<hr>');

                // Bold & Italic
                html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
                html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');

                // Inline code
                html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

                // Badges
                html = html.replace(/\[SUPPORTED\]/g, '<span class="md-badge green">SUPPORTED</span>');
                html = html.replace(/\[LIKELY_CATALYST\]/g, '<span class="md-badge blue">LIKELY CATALYST</span>');
                html = html.replace(/\[EXTREME SURGE\]/g, '<span class="md-badge green">EXTREME SURGE</span>');
                html = html.replace(/\[DIVERGENT\]/g, '<span class="md-badge amber">DIVERGENT</span>');
                html = html.replace(/\[INSTITUTIONAL BUY\]/g, '<span class="md-badge blue">INSTITUTIONAL BUY</span>');

                // GFM Tables: detect lines with |
                const lines = html.split('\n');
                let inTable = false;
                let tableHtml = '';
                let resultLines = [];

                for (let i = 0; i < lines.length; i++) {
                    const line = lines[i].trim();
                    if (line.startsWith('|') && line.endsWith('|')) {
                        const cells = line.split('|').map(c => c.trim()).slice(1, -1);
                        if (!inTable) {
                            inTable = true;
                            tableHtml = '<table><thead><tr>' + cells.map(c => `<th>${c}</th>`).join('') + '</tr></thead><tbody>';
                        } else if (line.includes('---')) {
                            // Separator row, skip
                            continue;
                        } else {
                            tableHtml += '<tr>' + cells.map(c => `<td>${c}</td>`).join('') + '</tr>';
                        }
                    } else {
                        if (inTable) {
                            inTable = false;
                            tableHtml += '</tbody></table>';
                            resultLines.push(tableHtml);
                            tableHtml = '';
                        }
                        resultLines.push(lines[i]);
                    }
                }
                if (inTable) {
                    tableHtml += '</tbody></table>';
                    resultLines.push(tableHtml);
                }
                html = resultLines.join('\n');

                // Lists
                html = html.replace(/^\* (.*$)/gim, '<li>$1</li>');
                html = html.replace(/^[0-9]+\. (.*$)/gim, '<li>$1</li>');
                // Wrap list items
                html = html.replace(/(<li>.*<\/li>(\s*<li>.*<\/li>)*)/gim, '<ul>$1</ul>');

                // Paragraph breaks
                html = html.replace(/\n\n+/g, '</p><p>');

                return html;
            }
