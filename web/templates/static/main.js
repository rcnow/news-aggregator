let timeoutId;
let isSorting = true;


function filterNews() {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => {
        const searchText = document.getElementById('searchInput').value.trim();
        console.log(searchText)
        const loading = document.getElementById('loading');
        if (loading)
            loading.style.display = 'block';
        fetch('/filter-news', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: 'query=' + encodeURIComponent(searchText),
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.text();
        })
        .then(html => {
            const feedContainer = document.getElementsByClassName('feed-view')[0];
            feedContainer.innerHTML = html;
        })
        .catch(error => {
            console.error('Error:', error);
        })
        .finally(() => {
            console.log(searchText)
            if(loading)
                loading.style.display = 'none';
        });
    }, 300);
}

document.getElementById('sort-feed').addEventListener('click', sortNews);
function sortNews() {
    console.log(isSorting);
    fetch('/sort-news', {
        method: 'GET',
        headers: {
            'Sort-Order': isSorting ? 'asc' : 'desc'
        }
    })
    .then(response => response.text())
    .then(html => {
        const feedContainer = document.getElementsByClassName('feed-view')[0];
        feedContainer.innerHTML = html;
        const svg = document.getElementById('sort-icon');
        const paths = svg.querySelectorAll('path');
        if (isSorting) {
            paths[0].setAttribute('d', 'M12 21C12 12.7714 12 5.57143 12 3');
            paths[1].setAttribute('d', 'M7 8L12 3L17 8');
        } else {
            paths[0].setAttribute('d', 'M12 3C12 12.2286 12 19.4286 12 21');
            paths[1].setAttribute('d', 'M7 16L12 21L17 16');
        }
        isSorting = !isSorting;
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

document.getElementById('searchInput').addEventListener('input', filterNews);

document.getElementById('addFeedButton').addEventListener('click', function() {
    fetch('/add-feed')
        .then(response => response.text())
        .then(html => {
            document.querySelector('.main-view').innerHTML = html;
        })
        .catch(error => console.error('Error loading form:', error));
});

document.querySelectorAll('.filter-popup input[type="radio"]').forEach(radio => {
    radio.addEventListener('change', function() {
        const selectedValue = this.value;
        fetch(`/sort-news?hours=${selectedValue}`, {
            method: 'GET',
        })
        .then(response => response.text())
        .then(html => {
            document.querySelector('.feed-view').innerHTML = html;
        })
        .catch(error => console.error('Error:', error));
    });
});