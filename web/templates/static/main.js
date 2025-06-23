let eventSource = null;
const fallbackSvg = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <path fill="currentColor" d="M22.204.01A2 2 0 0 1 24 2v20l-.01.204a2 2 0 0 1-1.786 1.785L22 24H2l-.204-.01A2 2 0 0 1 .01 22.203L0 22V2A2 2 0 0 1 1.796.01L2 0h20l.204.01ZM2 22h20V2H2v20Zm11.5-2h-3v-3h3v3ZM15 4a2.5 2.5 0 0 1 2.5 2.5v3.253a2.5 2.5 0 0 1-1.918 2.432l-2.082.498V15.5h-3v-3.21a2.5 2.5 0 0 1 1.918-2.433l2.082-.499V7H7V4h8Z"/>
                    </svg>`;
const arrowUp = 'M11.354 23a1 1 0 0 1-2 0V3.004l-7.19 7.19A.854.854 0 0 1 .957 8.986L9.594.35a.998.998 0 0 1 1.464-.06l8.692 8.692a.854.854 0 0 1-1.207 1.207L11.353 3v20Z'
const arrowDown = 'M9 1a1 1 0 0 1 2 0v19.996l7.19-7.19a.854.854 0 0 1 1.206 1.208L10.76 23.65a.998.998 0 0 1-1.464.06L.604 15.018A.854.854 0 1 1 1.81 13.81L9 21V1Z'
const elementList = {
    homeButton: document.getElementById('home-link'),
    settingButton: document.getElementById('setting-link'),
    addFeedButton: document.getElementById('add-feed-link'),
    helpButton: document.getElementById('help-link'),
    showAllNews: document.getElementById('show-all-news'),
    uniqueLink: document.querySelector('.unique-link-list'),
    themeToggle: document.querySelector('.theme-toggle input[type="checkbox"]'),
    feedView: document.querySelector('.feed-view'),
    mainView: document.querySelector('.panel-main'),
    count: document.querySelector('.count'),
    newTitle: document.querySelector('.panel-header'),
    searchInput: document.getElementById('searchInput'),
    sortTime: document.querySelectorAll('.filter-popup input[type="radio"]'),
    sortAscDesc: document.getElementById('sort-asc-desc'),
};
const API_ENDPOINTS = {
    LOAD_NEWS:        '/load-news',
    FILTER_BY_LINK:   '/filter-by-link',
    FILTER_BY_SEARCH: '/filter-by-search',
    HOME_VIEW:        '/home-view',
    SETTING_VIEW:     '/setting-view',
    ADD_FEED:         '/add-feed',
    HELP_VIEW:        '/help-view',
    SORT_NEWS:        '/sort-news',
}
const MESSAGES = {
    SSE_NEW: 'New SSE connection initiated',
    SSE_PREV: 'Previous SSE connection closed',
    SSE_INIT: 'SSE connection established',
    SSE_PING: 'Ping event received',
    SSE_UPDATE: 'Received update event',
    SSS_ERROR: 'SSE error occurred',
    NETWORK_ERROR: 'Network response was not ok',
}
//start
document.addEventListener('DOMContentLoaded', function() {
    setupSSE();
    const currentTheme = localStorage.getItem('theme');
    if (currentTheme) {
        document.documentElement.setAttribute('data-theme', currentTheme);
        if (currentTheme === 'dark') {
            elementList.themeToggle.checked = true;
        }
    }
});
//sse
function setupSSE() {
    if (eventSource) {
        eventSource.close();
        console.log(MESSAGES.SSE_PREV);
    }

    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    const initialBackoff = 1000;
    const maxBackoff = 30000;

    function connect() {
        eventSource = new EventSource('/sse');
        console.log(MESSAGES.SSE_NEW);

        eventSource.addEventListener('init', () => {
            console.log(MESSAGES.SSE_INIT);
            loadAllNews();
            reconnectAttempts = 0;
        });

        eventSource.addEventListener('update', () => {
            console.log(MESSAGES.SSE_UPDATE);
            loadAllNews();
        });

        eventSource.addEventListener('ping', () => {
            console.log(MESSAGES.SSE_PING);
        });

        eventSource.onerror = (err) => {
            console.error(MESSAGES.SSS_ERROR, err);
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
function setActiveButton(activeButton) {
    document.querySelectorAll('.menu-header a').forEach(button => {
        button.removeAttribute('class');
    });
    activeButton.classList.add('active');
}
//home
elementList.homeButton.addEventListener('click', function (e) {
    e.preventDefault();
    setActiveButton(this);
    HomeView();
});
async function HomeView() {
    try {
        const response = await fetch(API_ENDPOINTS.HOME_VIEW)
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.text();
        elementList.mainView.innerHTML = data;
    }
    catch (error) {
        console.error('Error HomeView:', error);
    }
};
//setting
elementList.settingButton.addEventListener('click', function (e) {
    e.preventDefault();
    setActiveButton(this);
    SettingView();
});
async function SettingView() {
    try {
        const response = await fetch(API_ENDPOINTS.SETTING_VIEW)
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.text();
        elementList.mainView.innerHTML = data;
    }
    catch (error) {
        console.error('Error SettingView:', error);
    }
};
//add
elementList.addFeedButton.addEventListener('click', function (e) {
    e.preventDefault();
    setActiveButton(this);
    AddNewFeedView();
});
async function AddNewFeedView() {
    try {
        const response = await fetch(API_ENDPOINTS.ADD_FEED)
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.text();
        elementList.mainView.innerHTML = data;
    }
    catch (error) {
        console.error('Error AddNewFeedView:', error);
    }
};
//help
elementList.helpButton.addEventListener('click', function (e) {
    e.preventDefault();
    setActiveButton(this);
    HelpView();
});
async function HelpView() {
    try {
        const response = await fetch(API_ENDPOINTS.HELP_VIEW)
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.text();
        elementList.mainView.innerHTML = data;
    }
    catch (error) {
    }
}
//theme switch
elementList.themeToggle.addEventListener('change', function(e) {
    e.preventDefault();
    if (e.target.checked) {
        document.documentElement.setAttribute('data-theme', 'dark');
        localStorage.setItem('theme', 'dark');
    } else {
        document.documentElement.setAttribute('data-theme', 'light');
        localStorage.setItem('theme', 'light');
    }
})
// show all news
elementList.showAllNews.addEventListener('click', function (e) {
    e.preventDefault();
    loadAllNews();
});
//load news
async function loadAllNews() {
    elementList.showAllNews.classList.add('active');
    try {
        const response = await fetch(API_ENDPOINTS.LOAD_NEWS)
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.json();
        const timeFilterValue = data.timeFilterValue;
        elementList.feedView.innerHTML = data.feedViewHTML;
        elementList.count.textContent = data.totalCount;
        elementList.newTitle.innerHTML = '';

        const faviconSpan = document.createElement('span');
        faviconSpan.className = 'favicon';
        faviconSpan.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" class="icon" fill="none" viewBox="0 0 24 24">
        <path fill="currentColor" fill-rule="evenodd" d="M24 0H0v3h24V0Zm-6 7H0v3h18V7ZM0 14h18v3H0v-3Zm24 7H0v3h24v-3Z" clip-rule="evenodd"/>
            </svg>`
        elementList.newTitle.appendChild(faviconSpan);
        elementList.newTitle.appendChild(document.createTextNode(`All news for the last ` + data.timeFilterValue + ` hours`));
        document.querySelectorAll('.filter-popup input[type="radio"]').forEach((radio) => {
            radio.checked = parseInt(radio.value, 10) === timeFilterValue;
        });
        elementList.uniqueLink.innerHTML = '';
        data.uniqueItems.forEach((item) => {
            const link = document.createElement('a');
            link.href = item.channelLink;
            link.dataset.channel = item.channelLink;
            const faviconSpan = document.createElement('span');
            faviconSpan.className = 'favicon';
            const faviconImg = document.createElement('img');
            faviconImg.src = data.uniqueFaviconURLs[item.channelLink] || fallbackSvg;
            faviconImg.alt = 'favicon';
            faviconImg.onerror = () => {
                faviconImg.remove();
                faviconSpan.innerHTML = fallbackSvg;
            };
            faviconSpan.appendChild(faviconImg);
            link.innerHTML = `${item.channelTitle} <div class="info"><span class="category">${item.category}</span> <span class="count">${data.uniqueCounts[item.channelLink]}</span></div>`;
            link.prepend(faviconSpan);
            elementList.uniqueLink.appendChild(link);
            const svg = document.getElementById('sort-icon');
            if (data.sortFilter === 'desc') {
               svg.querySelector('path').setAttribute('d', arrowDown);
            }
            else {
                svg.querySelector('path').setAttribute('d', arrowUp);
            }
            elementList.sortAscDesc.dataset.sort = data.sortFilter;
    });
}
    catch (error) {
        console.error('Error loadAllNews:', error);
    }
};
//show unique
elementList.uniqueLink.addEventListener('click', function(e) {
    e.preventDefault();
    const link = e.target.closest('a');
    if (link) {
        document.querySelectorAll('.unique-link-list a.active').forEach(item => {
            item.classList.remove('active');
        });

        link.classList.add('active');
        showUniqueNews(link);
    }
});
async function showUniqueNews(link) {
    const channelLink = link.dataset.channel || link.getAttribute('href');
    elementList.showAllNews.classList.remove('active');
    try {
        const response = await fetch(API_ENDPOINTS.FILTER_BY_LINK, {
            method: 'GET',
            headers: {
                'Link': channelLink
            }
        });
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.json();
        elementList.feedView.innerHTML = data.feedViewHTML;
        elementList.count.textContent = data.totalCount;
        elementList.newTitle.innerHTML = '';

        const faviconSpan = document.createElement('span');
        faviconSpan.className = 'favicon';
        const faviconImg = document.createElement('img');
        faviconImg.src = data.uniqueFaviconURLs[channelLink] || fallbackSvg;
        faviconImg.alt = 'favicon';
        faviconImg.onerror = () => {
            faviconImg.remove();
            faviconSpan.innerHTML = fallbackSvg;
        };
        faviconSpan.appendChild(faviconImg);
        elementList.newTitle.appendChild(faviconSpan);
        elementList.newTitle.appendChild(document.createTextNode(data.channelTitle));
        document.querySelectorAll('.unique-link-list a').forEach(anchor => {
            if (anchor.classList.length === 0) {
                anchor.removeAttribute('class');
            }
        });
    }
    catch (error) {
        console.error('Error loadAllNews:', error);
    }
};
//filter search
elementList.searchInput.addEventListener('input', function(e) {
    e.preventDefault();
    const searchValue = elementList.searchInput.value.trim();
    if (searchValue === '') {
        loadAllNews();
        return;
    }
    filterNewsBySearch(searchValue);
});
async function filterNewsBySearch(searchValue) {
    try {
        const response = await fetch(API_ENDPOINTS.FILTER_BY_SEARCH, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Search-Query': encodeURIComponent(searchValue),
            },
        });
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.json();
        elementList.feedView.innerHTML = data.feedViewHTML;
        elementList.count.textContent = data.totalCount;
        elementList.newTitle.textContent = `Search results for "${searchValue}"`;
    }
    catch (error) {
        console.error('Error filtering news by filterNewsBySearch', error);
    }
};
 //filter time
elementList.sortTime.forEach(radio => {
    radio.addEventListener('change', function() {
        const selectRadioValue = this.value
        filterNewsByTime(selectRadioValue);
    });
});
async function filterNewsByTime(timeValue) {
    try {
        const response = await fetch(API_ENDPOINTS.SORT_NEWS, {
            method: 'GET',
            headers: {
                'timeFilter': timeValue,
            },
        });
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.json();
        elementList.feedView.innerHTML = data.feedViewHTML;
        elementList.count.textContent = data.totalCount;
        elementList.newTitle.innerHTML = '';
        const faviconSpan = document.createElement('span');
        faviconSpan.className = 'favicon';
        faviconSpan.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" class="icon" fill="none" viewBox="0 0 24 24">
        <path fill="currentColor" fill-rule="evenodd" d="M24 0H0v3h24V0Zm-6 7H0v3h18V7ZM0 14h18v3H0v-3Zm24 7H0v3h24v-3Z" clip-rule="evenodd"/>
            </svg>`
        elementList.newTitle.appendChild(faviconSpan);
        elementList.newTitle.appendChild(document.createTextNode(`All news for the last ` + data.timeFilterValue + ` hours`));
        elementList.uniqueLink.innerHTML = '';
        data.uniqueItems.forEach((item) => {
            const link = document.createElement('a');
            link.href = item.channelLink;
            link.target = '_blank';
            const faviconSpan = document.createElement('span');
            faviconSpan.className = 'favicon';
            const faviconImg = document.createElement('img');
            faviconImg.src = data.uniqueFaviconURLs[item.channelLink] || fallbackSvg;
            faviconImg.alt = 'favicon';
            faviconImg.onerror = () => {
                faviconImg.remove();
                faviconSpan.innerHTML = fallbackSvg;
            };
            faviconSpan.appendChild(faviconImg);
            link.innerHTML = `${item.channelTitle} <div class="info"><span class="category">${item.category}</span> <span class="count">${data.uniqueCounts[item.channelLink]}</span></div>`;
            link.prepend(faviconSpan);
            elementList.uniqueLink.appendChild(link);

    });
}
    catch (error) {
        console.error('Error filtering news by filterNewsByTime', error);
    }
};
//filter asc/desc
elementList.sortAscDesc.addEventListener('click', function(e) {
    e.preventDefault();
    filterNewsByAscDesc();
});
async function filterNewsByAscDesc() {
    const currentSort = elementList.sortAscDesc.dataset.sort;
    const newSort = currentSort === 'desc' ? 'asc' : 'desc';
    try {
        const response = await fetch(API_ENDPOINTS.SORT_NEWS, {
            method: 'GET',
            headers: {
                'sortFilter': newSort,
            },
        });
        if (!response.ok) {
            throw new Error(MESSAGES.NETWORK_ERROR);
        }
        const data = await response.json();
        elementList.feedView.innerHTML = data.feedViewHTML;
        elementList.count.textContent = data.totalCount;
        const svg = document.getElementById('sort-icon');
        if (newSort === 'desc') {
           svg.querySelector('path').setAttribute('d', arrowDown);
        }
        else {
            svg.querySelector('path').setAttribute('d', arrowUp);
        }
        elementList.sortAscDesc.dataset.sort = newSort;
    }
    catch (error) {
        console.error('Error filtering news by filterNewsByAscDesc', error);
    }
};
