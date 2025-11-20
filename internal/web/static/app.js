const dashboard = document.getElementById('dashboard');
const connectionStatus = document.getElementById('connection-status');

// Function to fetch all metrics
async function fetchData() {
    try {
        // We'll fetch the metrics endpoint to discover sources, 
        // but for the dashboard we want the rich JSON data.
        // Since we don't have a "list all pages" endpoint, we might need to 
        // rely on a known list or add an endpoint to list available pages.
        // For now, let's assume we can fetch a list of pages or just try common ones.
        // WAIT, the server stores pages in a map. 
        // Let's add a small endpoint to list available pages or just use /metrics to find them.
        // Actually, let's try to fetch the root "/" with format=json if we modify the server to support listing.
        // BUT, the current server implementation of ServeHTTP on "/" might not list pages.
        // Let's assume for this iteration that we will modify the server to return a list of pages at /api/pages or similar.
        // OR, we can just try to fetch /metrics and parse the sources.

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
            dashboard.innerHTML = '<div class="loading">No devices found.</div>';
            return;
        }

        connectionStatus.classList.add('connected');

        // Now fetch details for each source
        const sourceDataPromises = Array.from(sources).map(async source => {
            // The source name in metrics corresponds to the URL path suffix (e.g. "battery1")
            // The server stores pages at root + name.
            // We need to construct the URL. 
            // If source is "battery1", URL is "/battery1".
            // Note: The metrics source label is derived from the path relative to root.

            try {
                // source comes from metrics and is the full path (e.g. /battery/1)
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

function renderDashboard(devices) {
    // Clear loading state if it exists, but try to preserve existing cards to minimize jank?
    // For simplicity, let's rebuild or diff. Rebuilding is easier for now.
    dashboard.innerHTML = '';

    devices.sort((a, b) => a.name.localeCompare(b.name));

    devices.forEach(device => {
        const card = document.createElement('div');
        card.className = 'card';

        const lastUpdated = device.data.last_updated ? new Date(device.data.last_updated).toLocaleTimeString() : 'Unknown';

        // Filter out internal fields
        const entries = Object.entries(device.data).filter(([key]) => key !== 'last_updated');

        let gridHtml = '';
        entries.forEach(([key, value]) => {
            gridHtml += `
                <div class="data-item">
                    <span class="data-label">${key}</span>
                    <span class="data-value">${value}</span>
                </div>
            `;
        });

        card.innerHTML = `
            <div class="card-header">
                <span class="card-title">${device.name}</span>
                <span class="card-timestamp">${lastUpdated}</span>
            </div>
            <div class="data-grid">
                ${gridHtml}
            </div>
        `;

        dashboard.appendChild(card);
    });
}

// Initial fetch
fetchData();

// Poll every 5 seconds
setInterval(fetchData, 5000);
