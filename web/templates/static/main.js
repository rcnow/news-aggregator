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
            paths[0].setAttribute('d', 'M9 24C9 14.8571 9 6.85714 9 4');
            paths[1].setAttribute('d', 'M2 12.45L9.04286 2.45L16.5 12.45');
        } else {
            paths[0].setAttribute('d', 'M9 20C9 10.8571 9 2.85714 9 0');
            paths[1].setAttribute('d', 'M2 11.45L9.04286 21.45L16.5 11.45');
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