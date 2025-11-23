document.addEventListener('DOMContentLoaded', function() {
    const loadAnalyticsBtn = document.getElementById('loadAnalyticsBtn');
    const aliasInput = document.getElementById('aliasInput');
    const analyticsContainer = document.getElementById('analyticsContainer');
    const errorContainer = document.getElementById('errorContainer');
    let clicksChart = null;
    let userAgentsChart = null;

    loadAnalyticsBtn.addEventListener('click', async function() {
        const alias = aliasInput.value.trim();
        
        if (!alias) {
            showError('Пожалуйста, введите alias');
            return;
        }

        try {
            const response = await fetch(`/analytics/${alias}`);
            
            if (!response.ok) {
                if (response.status === 404) {
                    throw new Error('URL не найден');
                }
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка при загрузке аналитики');
            }

            const data = await response.json();
            renderAnalytics(data);
            analyticsContainer.classList.remove('hidden');
            errorContainer.classList.add('hidden');
            
        } catch (error) {
            showError(error.message);
            analyticsContainer.classList.add('hidden');
        }
    });

    function renderAnalytics(data) {
        // Общая статистика
        document.getElementById('totalClicks').textContent = data.total_clicks;
        
        // Сегодняшние клики (последние 24 часа)
        const today = new Date().toISOString().split('T')[0];
        const todayClicks = data.daily_stats[today] || 0;
        document.getElementById('todayClicks').textContent = todayClicks;
        
        // Месячные клики
        const currentMonth = new Date().toISOString().slice(0, 7);
        const monthClicks = data.monthly_stats[currentMonth] || 0;
        document.getElementById('monthClicks').textContent = monthClicks;
        
        // График переходов
        renderClicksChart(data.daily_stats);
        
        // График User-Agent
        renderUserAgentsChart(data.user_agent_stats);
        
        // Последние переходы
        renderClicksTable(data.clicks);
    }

    function renderClicksChart(dailyStats) {
        const ctx = document.getElementById('clicksChart').getContext('2d');
        
        // Сортируем даты
        const dates = Object.keys(dailyStats).sort();
        const counts = dates.map(date => dailyStats[date]);
        
        if (clicksChart) {
            clicksChart.destroy();
        }
        
        clicksChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: dates,
                datasets: [{
                    label: 'Переходов в день',
                    data: counts,
                    borderColor: '#6366f1',
                    backgroundColor: 'rgba(99, 102, 241, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.3
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            precision: 0
                        }
                    }
                }
            }
        });
    }

    function renderUserAgentsChart(userAgentStats) {
        const ctx = document.getElementById('userAgentsChart').getContext('2d');
        
        // Берем топ-5 User-Agent
        const entries = Object.entries(userAgentStats)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 5);
        
        const labels = entries.map(([ua]) => {
            // Упрощаем название User-Agent
            if (ua.includes('Chrome')) return 'Chrome';
            if (ua.includes('Firefox')) return 'Firefox';
            if (ua.includes('Safari')) return 'Safari';
            if (ua.includes('Edge')) return 'Edge';
            if (ua.includes('Mobile')) return 'Mobile';
            return 'Другие';
        });
        
        const counts = entries.map(([, count]) => count);
        
        if (userAgentsChart) {
            userAgentsChart.destroy();
        }
        
        userAgentsChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: labels,
                datasets: [{
                    data: counts,
                    backgroundColor: [
                        '#6366f1',
                        '#8b5cf6', 
                        '#ec4899',
                        '#f59e0b',
                        '#10b981'
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    }

    function renderClicksTable(clicks) {
        const tbody = document.getElementById('clicksTableBody');
        tbody.innerHTML = '';
        
        if (clicks.length === 0) {
            tbody.innerHTML = '<tr><td colspan="3" style="text-align: center; padding: 20px;">Нет данных о переходах</td></tr>';
            return;
        }
        
        clicks.slice(0, 10).forEach(click => {
            const row = document.createElement('tr');
            
            // Форматируем дату для отображения
            const date = new Date(click.clicked_at);
            const formattedDate = date.toLocaleString('ru-RU', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
            
            // Упрощаем User-Agent
            let userAgent = click.user_agent;
            if (userAgent.includes('Chrome')) userAgent = 'Chrome';
            else if (userAgent.includes('Firefox')) userAgent = 'Firefox';
            else if (userAgent.includes('Safari')) userAgent = 'Safari';
            else if (userAgent.includes('Mobile')) userAgent = 'Mobile';
            
            row.innerHTML = `
                <td>${formattedDate}</td>
                <td>${userAgent}</td>
                <td>${click.ip_address || 'Скрыт'}</td>
            `;
            tbody.appendChild(row);
        });
    }

    function showError(message) {
        errorContainer.textContent = message;
        errorContainer.classList.remove('hidden');
        setTimeout(() => {
            errorContainer.classList.add('hidden');
        }, 5000);
    }

    // Загрузка аналитики при открытии страницы с параметром alias
    const urlParams = new URLSearchParams(window.location.search);
    const aliasParam = urlParams.get('alias');
    if (aliasParam) {
        aliasInput.value = aliasParam;
        setTimeout(() => {
            loadAnalyticsBtn.click();
        }, 500);
    }
});