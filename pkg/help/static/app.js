// Help system JavaScript functionality

document.addEventListener('DOMContentLoaded', function() {
    // Initialize search functionality
    initSearch();
    
    // Initialize navigation
    initNavigation();
    
    // Initialize tooltips if Bootstrap is available
    if (typeof bootstrap !== 'undefined') {
        var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
        var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
            return new bootstrap.Tooltip(tooltipTriggerEl);
        });
    }
});

function initSearch() {
    const searchForm = document.getElementById('searchForm');
    const searchInput = document.getElementById('searchInput');
    const searchType = document.getElementById('searchType');
    
    if (searchForm && searchInput) {
        // Auto-submit search on type (with debounce)
        let searchTimeout;
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(function() {
                if (searchInput.value.length >= 2 || searchInput.value.length === 0) {
                    searchForm.submit();
                }
            }, 300);
        });
        
        // Submit on type change
        if (searchType) {
            searchType.addEventListener('change', function() {
                searchForm.submit();
            });
        }
    }
}

function initNavigation() {
    // Highlight current page in navigation
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-link');
    
    navLinks.forEach(function(link) {
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        }
    });
    
    // Add smooth scrolling for anchor links
    const anchorLinks = document.querySelectorAll('a[href^="#"]');
    anchorLinks.forEach(function(link) {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

// Utility function to filter sections by type
function filterSections(type) {
    const sections = document.querySelectorAll('.section-card');
    
    sections.forEach(function(section) {
        const sectionType = section.getAttribute('data-section-type');
        if (type === 'all' || sectionType === type) {
            section.style.display = 'block';
        } else {
            section.style.display = 'none';
        }
    });
    
    // Update filter buttons
    const filterButtons = document.querySelectorAll('.filter-btn');
    filterButtons.forEach(function(btn) {
        btn.classList.remove('active');
        if (btn.getAttribute('data-filter') === type) {
            btn.classList.add('active');
        }
    });
}

// Add filter functionality if filter buttons exist
document.addEventListener('DOMContentLoaded', function() {
    const filterButtons = document.querySelectorAll('.filter-btn');
    filterButtons.forEach(function(btn) {
        btn.addEventListener('click', function() {
            const filterType = this.getAttribute('data-filter');
            filterSections(filterType);
        });
    });
});

// Copy code functionality
function copyCode(button) {
    const codeBlock = button.parentElement.querySelector('code');
    if (codeBlock) {
        navigator.clipboard.writeText(codeBlock.textContent).then(function() {
            // Show feedback
            const originalText = button.textContent;
            button.textContent = 'Copied!';
            button.classList.add('btn-success');
            button.classList.remove('btn-outline-secondary');
            
            setTimeout(function() {
                button.textContent = originalText;
                button.classList.remove('btn-success');
                button.classList.add('btn-outline-secondary');
            }, 2000);
        });
    }
}

// Add copy buttons to code blocks
document.addEventListener('DOMContentLoaded', function() {
    const codeBlocks = document.querySelectorAll('pre code');
    codeBlocks.forEach(function(codeBlock) {
        const pre = codeBlock.parentElement;
        const copyButton = document.createElement('button');
        copyButton.className = 'btn btn-sm btn-outline-secondary copy-btn';
        copyButton.textContent = 'Copy';
        copyButton.onclick = function() { copyCode(this); };
        
        // Position the button
        pre.style.position = 'relative';
        copyButton.style.position = 'absolute';
        copyButton.style.top = '0.5rem';
        copyButton.style.right = '0.5rem';
        
        pre.appendChild(copyButton);
    });
});
