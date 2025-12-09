const dashboard = document.getElementById('dashboard');
const connectionStatus = document.getElementById('connection-status');

// Helper to format numbers with precision
function formatNumber(val, decimals = 1) {
    const num = parseFloat(val);
    if (isNaN(num)) return val;
    return num.toFixed(decimals);
}

// Function to fetch all metrics
async function fetchData() {
    try {
        const response = await fetch('/metrics');
        if (!response.ok) throw new Error('Network response was not ok');
        const text = await response.text();

        // Parse metrics to find unique sources
        const sources = new Set();
        const lines = text.split('\n');
        for (const line of lines) {
            const match = line.match(/source="([^"]+)"/);
            if (match) {
                sources.add(match[1]);
            }
        }

        if (sources.size === 0) {
            dashboard.innerHTML = `
                <div class="loading-state">
                    <p>No devices found. Waiting for connection...</p>
                </div>`;
            return;
        }

        connectionStatus.classList.add('connected');

        // Now fetch details for each source
        const sourceDataPromises = Array.from(sources).map(async source => {
            try {
                const res = await fetch(`${source}?format=json`);
                if (!res.ok) return null;
                return { name: source, data: await res.json() };
            } catch (e) {
                console.error(`Error fetching ${source}:`, e);
                return null;
            }
        });

        const results = await Promise.all(sourceDataPromises);
        renderDashboard(results.filter(r => r !== null));

    } catch (error) {
        console.error('Fetch error:', error);
        connectionStatus.classList.remove('connected');
    }
}

function processDeviceData(data) {
    const processed = {
        type: 'UNKNOWN', // BATTERY, INVERTER
        meta: {},
        metrics: {},
        cells: [],
        other: {}
    };

    // Helper to extract numeric value from string with unit (e.g. "53.2V")
    const extractVal = (str) => {
        if (str === undefined || str === null) return NaN;
        if (typeof str === 'number') return str;
        if (typeof str === 'string') {
            // Remove everything except digits, minus sign, and dot
            const clean = str.replace(/[^\d.-]/g, '');
            const val = parseFloat(clean);
            return val;
        }
        return NaN;
    };

    Object.entries(data).forEach(([key, value]) => {
        if (key === 'last_updated') {
            processed.meta.lastUpdated = value;
            return;
        }

        const lowerKey = key.toLowerCase();

        // Identify Cell Voltages
        const cellMatch = key.match(/^cell_(\d+)_voltage$/i) ||
            key.match(/^Cell voltage (\d+)$/i) ||
            key.match(/^cell_voltage_(\d+)$/i);

        if (cellMatch) {
            processed.type = 'BATTERY'; // Definitely a battery
            processed.cells.push({
                id: parseInt(cellMatch[1]),
                value: extractVal(value),
                raw: value
            });
            return;
        }

        // Identify Key Metrics

        // Voltage
        if (lowerKey === 'pack_voltage' ||
            lowerKey === 'battery voltage' ||
            lowerKey === 'battery_voltage' ||
            lowerKey === 'voltage') {
            processed.metrics.voltage = extractVal(value);
        }
        // Current
        else if (lowerKey === 'pack_current' ||
            lowerKey === 'battery current' ||
            lowerKey === 'current') {
            processed.metrics.current = extractVal(value);
        }
        // Power
        else if (lowerKey === 'power' ||
            lowerKey === 'battery power') {
            processed.metrics.power = extractVal(value);
        }
        // SOC
        else if (lowerKey === 'soc' ||
            lowerKey === 'state of charge' ||
            (lowerKey === 'battery capacity' && typeof value === 'string' && value.includes('%'))) {
            processed.metrics.soc = extractVal(value);
        }
        // Cycles
        else if (lowerKey === 'cycle_counts' ||
            lowerKey === 'cycle count' ||
            lowerKey === 'number of cycles') {
            processed.other['Cycle count'] = value;
        }
        // Statistics
        else if (lowerKey === 'max_cell_voltage' ||
            lowerKey === 'maximum cell voltage (bms)') {
            processed.metrics.maxCellVoltage = extractVal(value);
        }
        else if (lowerKey === 'min_cell_voltage' ||
            lowerKey === 'minimum cell voltage (bms)') {
            processed.metrics.minCellVoltage = extractVal(value);
        }
        else if (lowerKey === 'diff_cell_voltage' ||
            lowerKey === 'cell voltage difference' ||
            key === 'CellVoltageDiff') {
            processed.metrics.cellDelta = extractVal(value);
        }
        // Inverter Specific
        else if (lowerKey.includes('grid') ||
            lowerKey.includes('output') ||
            lowerKey.includes('load') ||
            lowerKey.includes('pv') ||
            lowerKey.includes('scc') ||
            lowerKey.includes('inverter')) {
            if (processed.type === 'UNKNOWN') processed.type = 'INVERTER';
            processed.other[key] = value;

            // Capture specific inverter metrics
            if (lowerKey.includes('load') && lowerKey.includes('percentage')) {
                processed.metrics.loadPercent = extractVal(value);
            }
            if (lowerKey.includes('pv') && lowerKey.includes('power')) {
                // Sum PV power if multiple
                processed.metrics.pvPower = (processed.metrics.pvPower || 0) + extractVal(value);
            }
            if (lowerKey.includes('ac output') && lowerKey.includes('voltage')) {
                processed.metrics.acVoltage = extractVal(value);
            }
        }
        // Temperature
        // Try to match specific temperature fields
        else if (lowerKey === 'environment_temp' ||
            lowerKey === 'battery temperature' ||
            lowerKey === 'temperature' ||
            lowerKey === 'average_cell_temp') {

            // Convert to Celsius if Kelvin (ends with K or just heuristic > 200)
            let tempVal = extractVal(value);
            if (!isNaN(tempVal)) {
                if (typeof value === 'string' && value.toUpperCase().endsWith('K')) {
                    tempVal = tempVal - 273.15;
                } else if (tempVal > 200) {
                    // Assume K if > 200 (unless battery is on fire)
                    tempVal = tempVal - 273.15;
                }
                processed.other['Temperature'] = tempVal.toFixed(1) + '°C';
            } else {
                processed.other['Temperature'] = value;
            }
        }
        else {
            // Keep everything else
            processed.other[key] = value;
        }
    });

    // Sort cells by ID
    processed.cells.sort((a, b) => a.id - b.id);

    return processed;
}

function renderDashboard(devices) {
    const dashboard = document.getElementById('dashboard');
    dashboard.innerHTML = '';

    devices.sort((a, b) => a.name.localeCompare(b.name));

    devices.forEach(device => {
        const { type, meta, metrics, cells, other } = processDeviceData(device.data);
        console.log(`Rendering ${device.name} as ${type}`, { metrics, other });

        const card = document.createElement('div');
        card.className = 'card';

        // 1. Header
        const lastUpdated = meta.lastUpdated ? new Date(meta.lastUpdated).toLocaleTimeString() : 'Unknown';

        const headerHtml = `
            <div class="card-header">
                <span class="card-title">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="2" y="7" width="16" height="10" rx="2" ry="2"></rect>
                        <line x1="22" y1="11" x2="22" y2="13"></line>
                    </svg>
                    ${device.name.replace('/', '')} <small style="opacity:0.6; font-size:0.8em">(${type})</small>
                </span>
                <span class="card-timestamp">${lastUpdated}</span>
            </div>
        `;

        // 2. Content based on type
        let contentHtml = '';

        if (type === 'INVERTER') {
            // INVERTER LAYOUT
            const loadPercent = !isNaN(metrics.loadPercent) ? metrics.loadPercent : 0;
            const pvPower = !isNaN(metrics.pvPower) ? metrics.pvPower : 0;
            const acVoltage = !isNaN(metrics.acVoltage) ? metrics.acVoltage : 0;
            const voltage = !isNaN(metrics.voltage) ? metrics.voltage : 0; // Battery voltage often in inverters too

            // Load Color
            let loadClass = 'low'; // reusing soc classes for green/orange/red logic but inverted meaning usually
            if (loadPercent > 80) loadClass = 'low'; // "low" class is red in css (from SOC context)? check style. css usually 'low' soc is bad (red). 
            // actually for SOC: low (<20) is red, medium (<50) yellow, high green.
            // For load: low load is green, high load is red. 
            // So we can map: low load -> high soc class (green), high load -> low soc class (red).
            let loadColorClass = 'high';
            if (loadPercent > 50) loadColorClass = 'medium';
            if (loadPercent > 80) loadColorClass = 'low';

            contentHtml = ``;

        } else {
            // BATTERY / DEFAULT LAYOUT
            const soc = metrics.soc !== undefined && !isNaN(metrics.soc) ? metrics.soc : 0;
            const voltage = !isNaN(metrics.voltage) ? formatNumber(metrics.voltage, 2) : '--';
            const current = !isNaN(metrics.current) ? formatNumber(metrics.current, 2) : '--';
            const cycleCount = other['Cycle count'] || other['cycle_count'] || '--';

            // SOC Color
            let socClass = 'high';
            if (soc < 20) socClass = 'low';
            else if (soc < 50) socClass = 'medium';

            // Cells
            let cellsHtml = '';
            if (cells.length > 0) {
                const values = cells.map(c => c.value);
                let calcMax = 0, calcMin = 0, calcDiff = 0, mean = 0;
                if (values.length > 0) {
                    calcMax = Math.max(...values);
                    calcMin = Math.min(...values);
                    calcDiff = calcMax - calcMin;
                    mean = values.reduce((a, b) => a + b, 0) / values.length;
                }
                const maxVal = !isNaN(metrics.maxCellVoltage) ? metrics.maxCellVoltage : calcMax;
                const minVal = !isNaN(metrics.minCellVoltage) ? metrics.minCellVoltage : calcMin;
                const delta = !isNaN(metrics.cellDelta) ? metrics.cellDelta : calcDiff;

                let gridItems = cells.map(cell => {
                    let statusClass = '';
                    if (cell.value === calcMax) statusClass = 'highest';
                    if (cell.value === minVal) statusClass = 'lowest';
                    return `<div class="cell-item ${statusClass}"><span class="cell-id">#${cell.id}</span><span class="cell-value">${cell.value.toFixed(3)}</span></div>`;
                }).join('');

                cellsHtml = `
                    <div class="cells-container">
                        <div class="section-title">Cell Voltages</div>
                        <div class="cell-stats-grid" style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.5rem; margin-bottom: 1rem; text-align: center; font-size: 0.85rem;">
                            <div style="background: rgba(0,0,0,0.2); pad:5px; border-radius:5px"><div style="color:var(--text-secondary); font-size:0.75rem">Min</div><div style="color:var(--warning-color); font-weight:600">${minVal.toFixed(3)}V</div></div>
                            <div style="background: rgba(0,0,0,0.2); pad:5px; border-radius:5px"><div style="color:var(--text-secondary); font-size:0.75rem">Max</div><div style="color:var(--success-color); font-weight:600">${maxVal.toFixed(3)}V</div></div>
                            <div style="background: rgba(0,0,0,0.2); pad:5px; border-radius:5px"><div style="color:var(--text-secondary); font-size:0.75rem">Mean</div><div>${mean.toFixed(3)}V</div></div>
                            <div style="background: rgba(0,0,0,0.2); pad:5px; border-radius:5px"><div style="color:var(--text-secondary); font-size:0.75rem">Delta</div><div>${delta.toFixed(3)}V</div></div>
                        </div>
                        <div class="cells-grid">${gridItems}</div>
                    </div>
                `;
            }

            contentHtml = `
                <div class="key-metrics">
                    <div class="soc-container">
                        <span class="soc-label">${soc}%</span>
                        <div class="progress-bar-bg">
                            <div class="progress-bar-fill ${socClass}" style="width: ${soc}%"></div>
                        </div>
                    </div>
                    
                    <div class="metric-box">
                        <span class="metric-label">Voltage</span>
                        <div><span class="metric-value">${voltage}</span> <span class="metric-unit">V</span></div>
                    </div>
                    <div class="metric-box">
                        <span class="metric-label">Current</span>
                        <div><span class="metric-value">${current}</span> <span class="metric-unit">A</span></div>
                    </div>
                </div>
                ${cellsHtml}
            `;
        }

        // 3. Other Data (Common)
        const otherHtml = Object.entries(other)
            .filter(([k]) => {
                const lk = k.toLowerCase();
                return !lk.includes('cell_') &&
                    !lk.includes('cell voltage');
            })
            .map(([k, v]) => `
                <div class="data-row">
                    <span class="data-key">${k}</span>
                    <span class="data-val">${v}</span>
                </div>
            `).join('');

        card.innerHTML = `
            ${headerHtml}
            ${contentHtml}
            <div class="other-data-grid">${otherHtml}</div>
        `;

        dashboard.appendChild(card);
    });
}

// Initial fetch
fetchData();

// Poll every 2 seconds for snappier updates
setInterval(fetchData, 2000);
