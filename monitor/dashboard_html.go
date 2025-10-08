package monitor

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mailgrid - Real-time Monitoring Dashboard</title>
    <link rel="stylesheet" href="/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
</head>
<body>
    <div class="dashboard">
        <!-- Header -->
        <header class="header">
            <div class="header-content">
                <div class="logo">
                    <i class="fas fa-paper-plane"></i>
                    <h1>Mailgrid</h1>
                    <span class="version">Real-time Monitor</span>
                </div>
                <div class="header-actions">
                    <div class="status-indicator" id="statusIndicator">
                        <div class="status-dot"></div>
                        <span>Monitoring Active</span>
                    </div>
                    <button class="refresh-btn" onclick="refreshData()">
                        <i class="fas fa-sync-alt"></i>
                    </button>
                </div>
            </div>
        </header>

        <!-- Main Content -->
        <main class="main-content">
            <!-- Key Metrics Cards -->
            <section class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-icon success">
                        <i class="fas fa-check-circle"></i>
                    </div>
                    <div class="metric-content">
                        <h3>Emails Sent</h3>
                        <div class="metric-value" id="emailsSent">0</div>
                        <div class="metric-change positive" id="sentChange">+0</div>
                    </div>
                </div>

                <div class="metric-card">
                    <div class="metric-icon warning">
                        <i class="fas fa-exclamation-triangle"></i>
                    </div>
                    <div class="metric-content">
                        <h3>Failed</h3>
                        <div class="metric-value" id="emailsFailed">0</div>
                        <div class="metric-change negative" id="failedChange">+0</div>
                    </div>
                </div>

                <div class="metric-card">
                    <div class="metric-icon info">
                        <i class="fas fa-percentage"></i>
                    </div>
                    <div class="metric-content">
                        <h3>Success Rate</h3>
                        <div class="metric-value" id="successRate">0%</div>
                        <div class="metric-trend" id="rateTrend">
                            <i class="fas fa-arrow-up"></i>
                        </div>
                    </div>
                </div>

                <div class="metric-card">
                    <div class="metric-icon primary">
                        <i class="fas fa-tachometer-alt"></i>
                    </div>
                    <div class="metric-content">
                        <h3>Rate</h3>
                        <div class="metric-value" id="emailRate">0/min</div>
                        <div class="metric-subtitle">emails per minute</div>
                    </div>
                </div>
            </section>

            <!-- Charts and Details -->
            <section class="dashboard-grid">
                <!-- Live Chart -->
                <div class="dashboard-card chart-card">
                    <div class="card-header">
                        <h2>
                            <i class="fas fa-chart-line"></i>
                            Real-time Performance
                        </h2>
                        <div class="card-actions">
                            <button class="btn-secondary" onclick="toggleChartType()">
                                <i class="fas fa-exchange-alt"></i>
                            </button>
                        </div>
                    </div>
                    <div class="card-content">
                        <canvas id="performanceChart" width="400" height="200"></canvas>
                    </div>
                </div>

                <!-- Campaign Details -->
                <div class="dashboard-card">
                    <div class="card-header">
                        <h2>
                            <i class="fas fa-info-circle"></i>
                            Campaign Details
                        </h2>
                    </div>
                    <div class="card-content">
                        <div class="detail-grid">
                            <div class="detail-item">
                                <span class="detail-label">Job ID</span>
                                <span class="detail-value" id="jobId">-</span>
                            </div>
                            <div class="detail-item">
                                <span class="detail-label">Total Recipients</span>
                                <span class="detail-value" id="totalRecipients">-</span>
                            </div>
                            <div class="detail-item">
                                <span class="detail-label">CSV File</span>
                                <span class="detail-value" id="csvFile">-</span>
                            </div>
                            <div class="detail-item">
                                <span class="detail-label">Template</span>
                                <span class="detail-value" id="templateFile">-</span>
                            </div>
                            <div class="detail-item">
                                <span class="detail-label">Uptime</span>
                                <span class="detail-value" id="uptime">-</span>
                            </div>
                            <div class="detail-item">
                                <span class="detail-label">Last Updated</span>
                                <span class="detail-value" id="lastUpdated">-</span>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Recipients Table -->
                <div class="dashboard-card recipients-card">
                    <div class="card-header">
                        <h2>
                            <i class="fas fa-users"></i>
                            Recipients Status
                        </h2>
                        <div class="card-actions">
                            <input type="text" placeholder="Search recipients..." id="recipientSearch" class="search-input">
                            <select id="statusFilter" class="status-filter">
                                <option value="">All Status</option>
                                <option value="pending">Pending</option>
                                <option value="sending">Sending</option>
                                <option value="sent">Sent</option>
                                <option value="failed">Failed</option>
                                <option value="retry">Retry</option>
                            </select>
                        </div>
                    </div>
                    <div class="card-content">
                        <div class="table-container">
                            <table class="recipients-table">
                                <thead>
                                    <tr>
                                        <th>Email</th>
                                        <th>Status</th>
                                        <th>Duration</th>
                                        <th>Message</th>
                                        <th>Timestamp</th>
                                    </tr>
                                </thead>
                                <tbody id="recipientsTable">
                                    <tr class="loading-row">
                                        <td colspan="5">
                                            <div class="loading">
                                                <i class="fas fa-spinner fa-spin"></i>
                                                Loading recipients...
                                            </div>
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    </div>

    <!-- Scripts -->
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="/script.js"></script>
</body>
</html>`