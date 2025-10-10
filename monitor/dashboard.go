package monitor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// DashboardServer serves the aesthetic monitoring dashboard
type DashboardServer struct {
	server   *Server
	template *template.Template
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(server *Server) *DashboardServer {
	return &DashboardServer{
		server:   server,
		template: template.Must(template.New("dashboard").Parse(dashboardHTML)),
	}
}

// ServeHTTP handles HTTP requests for the dashboard
func (d *DashboardServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		d.serveDashboard(w, r)
	case "/api/stats":
		d.serveStats(w, r)
	case "/api/recipients":
		d.serveRecipients(w, r)
	case "/styles.css":
		d.serveCSS(w, r)
	case "/script.js":
		d.serveJS(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (d *DashboardServer) serveDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := d.template.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (d *DashboardServer) serveStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := d.server.GetStats()

	// Calculate real-time metrics
	now := time.Now()
	var emailsPerMinute float64
	if !stats.StartTime.IsZero() {
		duration := now.Sub(stats.StartTime).Minutes()
		if duration > 0 {
			emailsPerMinute = float64(stats.SentCount+stats.FailedCount) / duration
		}
	}

	// Build response
	response := map[string]any{
		"campaignStats": stats,
		"realTimeMetrics": map[string]any{
			"emailsPerMinute": emailsPerMinute,
			"successRate":     calculateSuccessRate(int64(stats.SentCount), int64(stats.FailedCount)),
			"currentTime":     now.Format(time.RFC3339),
			"uptime":          formatDuration(now.Sub(stats.StartTime)),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (d *DashboardServer) serveRecipients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	recipients := d.server.GetRecipients()
	json.NewEncoder(w).Encode(recipients)
}

func (d *DashboardServer) serveCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write([]byte(dashboardCSS))
}

func (d *DashboardServer) serveJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")

	// Real-time JavaScript with instant updates and job tracking
	simpleJS := `
class MonitoringDashboard {
    constructor() {
        this.chart = null;
        this.refreshInterval = 250; // 250ms for real-time updates
        this.init();
    }

    init() {
        this.setupChart();
        this.startDataPolling();
        this.setupEventListeners();
        this.loadInitialData();
    }

    setupChart() {
        const ctx = document.getElementById('performanceChart');
        if (!ctx) return;

        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Emails Sent',
                    data: [],
                    borderColor: '#10b981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
    }

    async fetchStats() {
        try {
            const response = await fetch('/api/stats');
            const data = await response.json();

            // Data received successfully

            this.updateMetrics(data);
            this.updateChart(data);
            this.updateCampaignDetails(data);
        } catch (error) {
            console.error('Error fetching stats:', error);
        }
    }

    updateMetrics(data) {

        if (data.campaignStats) {
            const sentCount = data.campaignStats.sent_count || 0;
            const failedCount = data.campaignStats.failed_count || 0;
            const totalRecipients = data.campaignStats.total_recipients || 0;


            // Update counter elements directly first, then animate
            const sentElement = document.getElementById('emailsSent');
            const failedElement = document.getElementById('emailsFailed');

            if (sentElement) {
                sentElement.textContent = sentCount;
            } else {
                console.error('emailsSent element not found!');
            }

            if (failedElement) {
                failedElement.textContent = failedCount;
            } else {
                console.error('emailsFailed element not found!');
            }

            // Calculate and display success rate
            const successRate = totalRecipients > 0 ? ((sentCount / totalRecipients) * 100).toFixed(1) : 0;
            const successRateElement = document.getElementById('successRate');
            if (successRateElement) {
                successRateElement.textContent = successRate + '%';
            }

            // Calculate emails per minute
            const emailsPerMinute = data.realTimeMetrics?.emailsPerMinute || 0;
            const emailRateElement = document.getElementById('emailRate');
            if (emailRateElement) {
                emailRateElement.textContent = emailsPerMinute.toFixed(1) + '/min';
            }

            // Update last updated timestamp
            const lastUpdatedElement = document.getElementById('lastUpdated');
            if (lastUpdatedElement) {
                lastUpdatedElement.textContent = new Date().toLocaleTimeString();
            }
        } else {
            console.error('No campaignStats in data!');
        }
    }


    updateChart(data) {
        if (!this.chart || !data.campaignStats) return;

        const now = new Date().toLocaleTimeString();
        this.chart.data.labels.push(now);
        this.chart.data.datasets[0].data.push(data.campaignStats.sent_count || 0);

        if (this.chart.data.labels.length > 20) {
            this.chart.data.labels.shift();
            this.chart.data.datasets[0].data.shift();
        }

        this.chart.update('none');
    }

    updateCampaignDetails(data) {
        if (data.campaignStats) {
            const stats = data.campaignStats;

            // Update campaign details
            document.getElementById('jobId').textContent = stats.job_id || '-';
            document.getElementById('totalRecipients').textContent = stats.total_recipients || '-';
            document.getElementById('csvFile').textContent = stats.csv_file || '-';
            document.getElementById('templateFile').textContent = stats.template_file || '-';

            // Update uptime from realTimeMetrics
            if (data.realTimeMetrics) {
                document.getElementById('uptime').textContent = data.realTimeMetrics.uptime || '-';
            }
        }
    }

    startDataPolling() {
        setInterval(() => this.fetchStats(), this.refreshInterval);
    }

    setupEventListeners() {
        // Add event listeners for search and filters
    }

    async loadInitialData() {
        await this.fetchStats();
    }
}

document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new MonitoringDashboard();
});

function refreshData() {
    if (window.dashboard) {
        window.dashboard.fetchStats();
    }
}

`

	w.Write([]byte(simpleJS))
}

func calculateSuccessRate(sent, failed int64) float64 {
	total := sent + failed
	if total == 0 {
		return 0
	}
	return (float64(sent) / float64(total)) * 100
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}
