// PMaaS Platform JavaScript
document.addEventListener('DOMContentLoaded', function() {
    console.log('PMaaS Platform loaded');

    // Add click handlers for navigation
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            // Remove active class from all links
            navLinks.forEach(l => l.classList.remove('active'));
            // Add active class to clicked link
            this.classList.add('active');
        });
    });

    // Add click handlers for action buttons
    const actionButtons = document.querySelectorAll('.btn');
    actionButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            // Only add click effect for buttons that don't have actual functionality yet
            if (this.textContent.includes('Add') || 
                this.textContent.includes('Edit') || 
                this.textContent.includes('Delete')) {
                
                // Show a placeholder alert
                const action = this.textContent.trim();
                alert(`${action} functionality coming soon!`);
            }
        });
    });

    // Add search functionality for properties table
    const searchInput = document.querySelector('input[placeholder*="Search"]');
    if (searchInput) {
        searchInput.addEventListener('input', function(e) {
            const searchTerm = e.target.value.toLowerCase();
            const tableRows = document.querySelectorAll('.table tbody tr');
            
            tableRows.forEach(row => {
                const address = row.querySelector('td:first-child')?.textContent.toLowerCase();
                if (address && address.includes(searchTerm)) {
                    row.style.display = '';
                } else {
                    row.style.display = 'none';
                }
            });
        });
    }

    // Add status filter functionality
    const statusSelect = document.querySelector('select');
    if (statusSelect && statusSelect.querySelector('option[value="All Status"]')) {
        statusSelect.addEventListener('change', function(e) {
            const selectedStatus = e.target.value.toLowerCase();
            const tableRows = document.querySelectorAll('.table tbody tr');
            
            tableRows.forEach(row => {
                const statusElement = row.querySelector('.status');
                if (statusElement) {
                    const rowStatus = statusElement.textContent.toLowerCase();
                    if (selectedStatus === 'all status' || rowStatus.includes(selectedStatus)) {
                        row.style.display = '';
                    } else {
                        row.style.display = 'none';
                    }
                }
            });
        });
    }

    // Animate stat cards on dashboard
    const statCards = document.querySelectorAll('.stat-card');
    statCards.forEach((card, index) => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(20px)';
        
        setTimeout(() => {
            card.style.transition = 'all 0.5s ease';
            card.style.opacity = '1';
            card.style.transform = 'translateY(0)';
        }, index * 100);
    });

    // Add hover effect to property cards
    const propertyCards = document.querySelectorAll('.property-card');
    propertyCards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-2px)';
            this.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.15)';
            this.style.transition = 'all 0.3s ease';
        });
        
        card.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
            this.style.boxShadow = '0 1px 3px rgba(0, 0, 0, 0.1)';
        });
    });

    // Auto-refresh data every 5 minutes (for future real-time updates)
    // setInterval(function() {
    //     console.log('Auto-refresh would happen here');
    // }, 300000);
});
