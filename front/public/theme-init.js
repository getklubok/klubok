(function () {
    const theme = localStorage.getItem('klubok-theme') || 'system';
    if (theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        document.documentElement.setAttribute('data-bs-theme', 'dark');
        document.documentElement.classList.add('theme-dark');
        document.documentElement.style.setProperty('color-scheme', 'dark');
    }
})();
