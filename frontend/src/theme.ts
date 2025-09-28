// Theme management
function getStoredTheme(): string {
  return localStorage.getItem('theme') || 'auto';
}

function setStoredTheme(theme: string) {
  localStorage.setItem('theme', theme);
}

function applyTheme(theme: string) {
  const html = document.documentElement;

  // Always remove dark class first to ensure clean state
  html.classList.remove('dark');

  if (theme === 'dark') {
    html.classList.add('dark');
  } else if (theme === 'light') {
    // Light mode uses default (no dark class)
  } else {
    // Auto mode - follow system preference
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      html.classList.add('dark');
    }
  }
}

function getCurrentTheme(): string {
  return document.documentElement.classList.contains('dark') ? 'dark' : 'light';
}

function createThemeToggle(): string {
  const currentTheme = getStoredTheme();
  const icons = {
    light: '☀️',
    dark: '🌙',
    auto: '💻'
  };

  return `
    <div class="relative">
      <button id="theme-toggle"
              class="flex items-center gap-2 p-2 rounded-lg bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150 text-gray-700 dark:text-gray-300">
        <span id="theme-icon" class="text-lg">${icons[currentTheme as keyof typeof icons]}</span>
        <span id="theme-text" class="text-sm capitalize text-gray-700 dark:text-gray-300 hidden">${currentTheme}</span>
        <svg id="theme-chevron" class="w-4 h-4 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
        </svg>
      </button>
      <div id="theme-menu" class="hidden absolute right-0 mt-2 w-32 bg-white/95 dark:bg-gray-800/95 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl backdrop-blur-md z-50">
        <button data-theme="light" class="w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-t-lg text-gray-700 dark:text-gray-300 flex items-center gap-2">
          ☀️ Light
        </button>
        <button data-theme="dark" class="w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 flex items-center gap-2">
          🌙 Dark
        </button>
        <button data-theme="auto" class="w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-b-lg text-gray-700 dark:text-gray-300 flex items-center gap-2">
          💻 Auto
        </button>
      </div>
    </div>
  `;
}

function setupThemeToggle(onThemeChange?: () => void) {
  const toggle = document.getElementById('theme-toggle');
  const menu = document.getElementById('theme-menu');
  const icon = document.getElementById('theme-icon');
  const text = document.getElementById('theme-text');
  const chevron = document.getElementById('theme-chevron');

  if (!toggle || !menu || !icon || !text || !chevron) return;

  const icons = { light: '☀️', dark: '🌙', auto: '💻' };

  // Toggle menu visibility
  toggle.addEventListener('click', (e) => {
    e.stopPropagation();
    const isHidden = menu.classList.contains('hidden');

    if (isHidden) {
      // Opening menu - show text and rotate chevron
      menu.classList.remove('hidden');
      text.classList.remove('hidden');
      chevron.style.transform = 'rotate(180deg)';
    } else {
      // Closing menu - hide text and reset chevron
      menu.classList.add('hidden');
      text.classList.add('hidden');
      chevron.style.transform = 'rotate(0deg)';
    }
  });

  // Close menu when clicking outside
  document.addEventListener('click', () => {
    menu.classList.add('hidden');
    text.classList.add('hidden');
    chevron.style.transform = 'rotate(0deg)';
  });

  // Handle theme selection
  menu.addEventListener('click', (e) => {
    const target = e.target as HTMLElement;
    const button = target.closest('[data-theme]') as HTMLElement;
    if (!button) return;

    const theme = button.dataset.theme!;
    setStoredTheme(theme);
    applyTheme(theme);

    icon.textContent = icons[theme as keyof typeof icons];
    text.textContent = theme.charAt(0).toUpperCase() + theme.slice(1);
    menu.classList.add('hidden');
    text.classList.add('hidden');
    chevron.style.transform = 'rotate(0deg)';

    // Callback for theme changes (e.g., refresh heatmap)
    if (onThemeChange) {
      onThemeChange();
    }
  });
}

export { getStoredTheme, setStoredTheme, applyTheme, getCurrentTheme, createThemeToggle, setupThemeToggle };