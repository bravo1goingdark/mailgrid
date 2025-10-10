package monitor

const dashboardCSS = `
:root {
    --primary-color: #3b82f6;
    --primary-dark: #2563eb;
    --success-color: #10b981;
    --warning-color: #f59e0b;
    --error-color: #ef4444;
    --info-color: #06b6d4;

    --bg-primary: #ffffff;
    --bg-secondary: #f8fafc;
    --bg-tertiary: #f1f5f9;
    --bg-dark: #0f172a;

    --text-primary: #1e293b;
    --text-secondary: #64748b;
    --text-muted: #94a3b8;

    --border-color: #e2e8f0;
    --border-light: #f1f5f9;

    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);

    --radius: 8px;
    --radius-lg: 12px;
    --spacing: 1rem;
}

* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
    color: var(--text-primary);
    line-height: 1.5;
}

.dashboard {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

/* Header */
.header {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    border-bottom: 1px solid var(--border-light);
    position: sticky;
    top: 0;
    z-index: 100;
}

.header-content {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.logo i {
    font-size: 2rem;
    color: var(--primary-color);
    animation: float 3s ease-in-out infinite;
}

.logo h1 {
    font-size: 1.75rem;
    font-weight: 700;
    background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
}

.version {
    background: var(--primary-color);
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
}

.header-actions {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.status-indicator {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
}

.status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--success-color);
    animation: pulse 2s ease-in-out infinite;
}

.refresh-btn {
    background: var(--primary-color);
    color: white;
    border: none;
    padding: 0.5rem;
    border-radius: var(--radius);
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
}

.refresh-btn:hover {
    background: var(--primary-dark);
    transform: translateY(-1px);
}

.refresh-btn i {
    font-size: 1rem;
}

/* Main Content */
.main-content {
    flex: 1;
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    width: 100%;
}

/* Metrics Grid */
.metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
}

.metric-card {
    background: white;
    border-radius: var(--radius-lg);
    padding: 1.5rem;
    box-shadow: var(--shadow-md);
    border: 1px solid var(--border-light);
    transition: all 0.3s;
    position: relative;
    overflow: hidden;
}

.metric-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, var(--primary-color), var(--primary-dark));
}

.metric-card:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-lg);
}

.metric-card {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
}

.metric-icon {
    width: 48px;
    height: 48px;
    border-radius: var(--radius);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
    color: white;
}

.metric-icon.success { background: var(--success-color); }
.metric-icon.warning { background: var(--warning-color); }
.metric-icon.info { background: var(--info-color); }
.metric-icon.primary { background: var(--primary-color); }

.metric-content {
    flex: 1;
}

.metric-content h3 {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
    margin-bottom: 0.25rem;
}

.metric-value {
    font-size: 2rem;
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1;
    margin-bottom: 0.25rem;
}

.metric-change {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 0.125rem 0.5rem;
    border-radius: 4px;
    display: inline-block;
}

.metric-change.positive {
    background: rgba(16, 185, 129, 0.1);
    color: var(--success-color);
}

.metric-change.negative {
    background: rgba(239, 68, 68, 0.1);
    color: var(--error-color);
}

.metric-subtitle {
    font-size: 0.75rem;
    color: var(--text-muted);
}

/* Dashboard Grid */
.dashboard-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 1.5rem;
}

.dashboard-card {
    background: white;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
    border: 1px solid var(--border-light);
    overflow: hidden;
}

.recipients-card {
    grid-column: 1 / -1;
}

.card-header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-light);
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-secondary);
}

.card-header h2 {
    font-size: 1.125rem;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.card-header i {
    color: var(--primary-color);
}

.card-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.card-content {
    padding: 1.5rem;
}

.chart-card .card-content {
    padding: 1rem;
}

/* Detail Grid */
.detail-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
}

.detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.detail-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.025em;
}

.detail-value {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-primary);
}

/* Table */
.table-container {
    overflow-x: auto;
}

.recipients-table {
    width: 100%;
    border-collapse: collapse;
}

.recipients-table th,
.recipients-table td {
    text-align: left;
    padding: 0.75rem;
    border-bottom: 1px solid var(--border-light);
}

.recipients-table th {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.025em;
    background: var(--bg-tertiary);
}

.recipients-table td {
    font-size: 0.875rem;
}

.recipients-table tr:hover {
    background: var(--bg-secondary);
}

/* Status badges */
.status-badge {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.025em;
}

.status-pending { background: rgba(148, 163, 184, 0.2); color: var(--text-secondary); }
.status-sending { background: rgba(6, 182, 212, 0.2); color: var(--info-color); }
.status-sent { background: rgba(16, 185, 129, 0.2); color: var(--success-color); }
.status-failed { background: rgba(239, 68, 68, 0.2); color: var(--error-color); }
.status-retry { background: rgba(245, 158, 11, 0.2); color: var(--warning-color); }

/* Form Elements */
.search-input,
.status-filter {
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    font-size: 0.875rem;
    background: white;
    transition: all 0.2s;
}

.search-input:focus,
.status-filter:focus {
    outline: none;
    border-color: var(--primary-color);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn-secondary {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    padding: 0.5rem;
    border-radius: var(--radius);
    cursor: pointer;
    transition: all 0.2s;
    color: var(--text-secondary);
}

.btn-secondary:hover {
    background: var(--border-light);
    color: var(--text-primary);
}

/* Loading */
.loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 2rem;
    color: var(--text-muted);
}

.loading i {
    animation: spin 1s linear infinite;
}

/* Animations */
@keyframes float {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
}

@keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
}

@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

/* Responsive */
@media (max-width: 768px) {
    .main-content {
        padding: 1rem;
    }

    .header-content {
        padding: 1rem;
    }

    .dashboard-grid {
        grid-template-columns: 1fr;
    }

    .metrics-grid {
        grid-template-columns: 1fr;
    }

    .metric-card {
        padding: 1rem;
    }

    .card-header {
        padding: 1rem;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.5rem;
    }

    .card-actions {
        width: 100%;
        justify-content: space-between;
    }
}
`
