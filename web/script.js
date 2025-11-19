document.addEventListener('DOMContentLoaded', () => {
    console.log("PrimGO Landing Page Loaded");

    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            document.querySelector(this.getAttribute('href')).scrollIntoView({
                behavior: 'smooth'
            });
        });
    });
    // Lightbox for history image
    const historyImage = document.querySelector('.history-image');
    if (historyImage) {
        historyImage.addEventListener('click', () => {
            // Create overlay
            const overlay = document.createElement('div');
            overlay.className = 'lightbox-overlay';

            // Create image
            const img = document.createElement('img');
            img.src = historyImage.src;
            img.className = 'lightbox-image';
            img.alt = historyImage.alt;

            // Append
            overlay.appendChild(img);
            document.body.appendChild(overlay);

            // Trigger animation
            requestAnimationFrame(() => {
                overlay.classList.add('active');
            });

            // Close on click
            overlay.addEventListener('click', () => {
                overlay.classList.remove('active');
                setTimeout(() => {
                    overlay.remove();
                }, 300); // Match transition duration
            });
        });
    }
    // Mobile Menu Toggle
    const menuToggle = document.querySelector('.mobile-menu-toggle');
    const navLinks = document.querySelector('.nav-links');

    if (menuToggle && navLinks) {
        menuToggle.addEventListener('click', () => {
            menuToggle.classList.toggle('active');
            navLinks.classList.toggle('active');
        });

        // Close menu when clicking a link
        navLinks.querySelectorAll('a').forEach(link => {
            link.addEventListener('click', () => {
                menuToggle.classList.remove('active');
                navLinks.classList.remove('active');
            });
        });
    }
});
