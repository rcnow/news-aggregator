let timeoutId;
let eventSource = null;

function setupLoadingState() {
    const feedView = document.querySelector('.feed-view');
    if (feedView.children.length === 0) {
        feedView.innerHTML = `
            <div class="loading">
                <h3>Loading news...</h3>
            </div>
        `;
    }
}

function setupSSE() {
    if (eventSource) {
        eventSource.close();
        console.log("Previous SSE connection closed.");
    }

    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    const initialBackoff = 1000;
    const maxBackoff = 30000;

    function connect() {
        eventSource = new EventSource('/sse');
        console.log("New SSE connection initiated.");

        eventSource.addEventListener('init', () => {
            console.log('SSE connection established');
            loadNews();
            reconnectAttempts = 0;
        });

        eventSource.addEventListener('update', () => {
            console.log('Received update event');
            loadNews();
        });

        eventSource.addEventListener('ping', () => {
            console.log('Received ping');
        });

        eventSource.onerror = (err) => {
            console.error("SSE error:", err);
            if (eventSource) {
                eventSource.close();
            }
            if (reconnectAttempts < maxReconnectAttempts) {
                const backoffTime = Math.min(initialBackoff * Math.pow(2, reconnectAttempts), maxBackoff);
                console.log(`Reconnecting in ${backoffTime / 1000} seconds...`);
                reconnectAttempts++;
                setTimeout(connect, backoffTime);
            } else {
                console.error("Max reconnect attempts reached. Manual refresh required.");
            }
        };
    }

    connect();

    window.addEventListener('beforeunload', () => {
        if (eventSource) {
            eventSource.close();
            console.log("SSE connection closed on page unload.");
        }
    });
}


function loadNews() {
    fetch('/load-news')
        .then(response => response.json())
        .then(data => {
            const feedContainer = document.getElementsByClassName('feed-view')[0];
            feedContainer.innerHTML = data.feedViewHTML;
            document.querySelector('.count').textContent = data.totalCount;
            const timeFilterValue = data.timeFilterValue;
            document.querySelectorAll('.filter-popup input[type="radio"]').forEach((radio) => {
                radio.checked = parseInt(radio.value, 10) === timeFilterValue;
            });
            const channelMenu = document.querySelector('.channel-menu');
            channelMenu.innerHTML = '';
            data.uniqueItems.forEach((item) => {
                const link = document.createElement('a');
                link.href = item.channelLink;
                link.dataset.channel = item.channelLink;
                const faviconSpan = document.createElement('span');
                faviconSpan.className = 'favicon';
                const faviconImg = document.createElement('img');
                faviconImg.src = `${new URL(item.channelLink).origin}/favicon.ico`;
                faviconImg.alt = 'favicon';
                faviconImg.onerror = () => {
                    faviconImg.remove();
                    const fallbackSvg = `
                        <svg class="icon-q" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M5 7.8799C5 5.14387 7.11752 3.77585 7.11752 3.77585C7.6469 2.79807 13.1456 2.86383 15.1607 3.31984C17.1757 3.77585 17.7051 5.59986 17.8076 6.05586C17.91 6.51187 18.2345 7.87988 17.7051 9.2479C17.1757 10.6159 16.3219 10.4844 15.1607 11.0062C13.9995 11.5279 12.0869 12.3084 11.9844 13.2862C11.8819 14.264 11.8819 17 11.8819 17" stroke="black" stroke-width="3"/>
                            <circle cx="12" cy="22" r="2" fill="#0B0B0B"/>
                        </svg>`;
                    faviconSpan.innerHTML = fallbackSvg;
                };
                faviconSpan.appendChild(faviconImg);
                link.innerHTML = `${item.channelTitle} <span class="count">${data.uniqueCounts[item.channelLink]}</span>`;
                link.prepend(faviconSpan);
                channelMenu.appendChild(link);
            });
            const svg = document.getElementById('sort-icon');
            const paths = svg.querySelectorAll('path');
            if (data.sortFilter === 'asc') {
                paths[0].setAttribute('d', 'M9 24C9 14.8571 9 6.85714 9 4');
                paths[1].setAttribute('d', 'M2 12.45L9.04286 2.45L16.5 12.45');
            } else {
                paths[0].setAttribute('d', 'M9 20C9 10.8571 9 2.85714 9 0');
                paths[1].setAttribute('d', 'M2 11.45L9.04286 21.45L16.5 11.45');
            }
        })
        .catch(error => {
            console.error('Error loading news:', error);
        });
}

document.addEventListener('DOMContentLoaded', function() {
    setupSSE();
    setupLoadingState();
});


document.getElementById('searchInput').addEventListener('input', filterNews);
function filterNews() {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => {
        const searchText = document.getElementById('searchInput').value.trim();
        const loading = document.getElementById('loading');
        if (loading)
            loading.style.display = 'block';
        fetch('/filter-by-search', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Search-Query': searchText,
            },
        })
        .then(response => response.json())
        .then(data => {
            console.log('Filtered data:', data);
            const feedContainer = document.getElementsByClassName('feed-view')[0];
            feedContainer.innerHTML = data.feedViewHTML;
        })
        .catch(error => {
            console.error('Error:', error);
        })
        .finally(() => {
            if(loading)
                loading.style.display = 'none';
        });
    }, 300);
}

document.getElementById('sort-feed').addEventListener('click', sortNews);
function sortNews() {
    const sortfeedElement = document.getElementById('sort-feed');
    const currentSort = sortfeedElement.dataset.sort;
    const newSort = currentSort === 'desc' ? 'asc' : 'desc';

    fetch('/sort-news', {
        method: 'GET',
        headers: {
            'sortFilter': newSort
        }
    })
    .then(response => response.json())
    .then(data => {
        const feedContainer = document.getElementsByClassName('feed-view')[0];
        feedContainer.innerHTML = data.feedViewHTML;
        document.querySelector('.count').textContent = data.totalCount;

        const svg = document.getElementById('sort-icon');
        const paths = svg.querySelectorAll('path');
        if (newSort === 'asc') {
            paths[0].setAttribute('d', 'M9 24C9 14.8571 9 6.85714 9 4');
            paths[1].setAttribute('d', 'M2 12.45L9.04286 2.45L16.5 12.45');
        } else {
            paths[0].setAttribute('d', 'M9 20C9 10.8571 9 2.85714 9 0');
            paths[1].setAttribute('d', 'M2 11.45L9.04286 21.45L16.5 11.45');
        }
        sortfeedElement.dataset.sort = newSort;
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

document.getElementById('addFeedButton').addEventListener('click', function() {
    fetch('/add-feed')
        .then(response => response.text())
        .then(html => {
            document.querySelector('.main-view').innerHTML = html;
        })
        .catch(error => console.error('Error loading form:', error));
});

document.getElementById('allNewsLoading').addEventListener('click', function(event) {
    event.preventDefault();
    loadNews();
});

document.querySelector('.channel-menu').addEventListener('click', function(event) {
    const link = event.target.closest('a');
    if (link) {
        event.preventDefault();
        const channelLink = link.dataset.channel || link.getAttribute('href');
        console.log('Channel link:', channelLink);
        fetch('/filter-by-link', {
            method: 'GET',
            headers: {
                'Link': channelLink
            }
        })
        .then(response => response.json())
        .then(data => {
            document.querySelector('.feed-view').innerHTML = data.feedViewHTML;
            document.querySelector('.count').textContent = data.totalCount;
        })
        .catch(error => console.error('Error', error));
    }
});

document.querySelectorAll('.filter-popup input[type="radio"]').forEach(radio => {
    radio.addEventListener('change', function() {
        const selectedValue = this.value;
        fetch('/sort-news', {
            method: 'GET',
            headers: {
                'timeFilter': selectedValue
            }
        })
        .then(response => response.json())
        .then(data => {
            document.querySelector('.feed-view').innerHTML = data.feedViewHTML;
            document.querySelector('.count').textContent = data.totalCount;
            const feedContainer = document.getElementsByClassName('feed-view')[0];
            feedContainer.innerHTML = data.feedViewHTML;
            document.querySelector('.count').textContent = data.totalCount;
            const channelMenu = document.querySelector('.channel-menu');
            channelMenu.innerHTML = '';
            data.uniqueItems.forEach((item) => {
                const link = document.createElement('a');
                link.href = item.channelLink;
                link.target = '_blank';
                const faviconSpan = document.createElement('span');
                faviconSpan.className = 'favicon';
                const faviconImg = document.createElement('img');
                faviconImg.src = `${new URL(item.channelLink).origin}/favicon.ico`;
                faviconImg.alt = 'favicon';
                faviconImg.onerror = () => {
                    faviconImg.remove();
                    const fallbackSvg = `
                        <svg class="icon-q" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M5 7.8799C5 5.14387 7.11752 3.77585 7.11752 3.77585C7.6469 2.79807 13.1456 2.86383 15.1607 3.31984C17.1757 3.77585 17.7051 5.59986 17.8076 6.05586C17.91 6.51187 18.2345 7.87988 17.7051 9.2479C17.1757 10.6159 16.3219 10.4844 15.1607 11.0062C13.9995 11.5279 12.0869 12.3084 11.9844 13.2862C11.8819 14.264 11.8819 17 11.8819 17" stroke="black" stroke-width="3"/>
                            <circle cx="12" cy="22" r="2" fill="#0B0B0B"/>
                        </svg>`;
                    faviconSpan.innerHTML = fallbackSvg;
                };
                faviconSpan.appendChild(faviconImg);
                link.innerHTML = `${item.channelTitle} <span class="count">${data.uniqueCounts[item.channelLink]}</span>`;
                link.prepend(faviconSpan);
                channelMenu.appendChild(link);
            });
        })
        .catch(error => console.error('Error:', error));
    });
});