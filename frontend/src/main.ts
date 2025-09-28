import './style.css'
// @ts-expect-error: No type definitions for 'cal-heatmap'
import CalHeatmap from 'cal-heatmap';
// @ts-expect-error: No type definitions for 'cal-heatmap'
import Tooltip from 'cal-heatmap/plugins/Tooltip';
import 'cal-heatmap/cal-heatmap.css';

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

function initializeBodyStyles() {
  document.body.className = 'bg-gray-50 dark:bg-gray-900 min-h-screen transition-colors duration-200';
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

function setupThemeToggle() {
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

    // Refresh heatmap if on habit page
    const habit = getHabitFromURL();
    if (habit) {
      drawHabitHeatmap(habit);
    }
  });
}

function drawHabitSummary(habit: string) {
  document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
    <div class="max-w-5xl mx-auto p-6">
      <div class="flex justify-between items-center mb-6">
        <div id="title" class="text-3xl font-bold text-gray-900 dark:text-white">${toTitleCase(habit)}</div>
        ${createThemeToggle()}
      </div>

      <!-- Top row of 3 cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
        <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30" data-stat="current-streak">
          <div class="text-lg text-gray-700 dark:text-gray-300">Current Streak</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30" data-stat="longest-streak">
          <div class="text-lg text-gray-700 dark:text-gray-300">Longest Streak</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30" data-stat="month-progress">
          <div class="text-lg text-gray-700 dark:text-gray-300">This Month</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
      </div>

      <!-- Second row of 3 cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30" data-stat="total-days">
          <div class="text-lg text-gray-700 dark:text-gray-300">Total Days Done</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30" data-stat="best-month">
          <div class="text-lg text-gray-700 dark:text-gray-300">Best Month</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30" data-stat="first-logged">
          <div class="text-lg text-gray-700 dark:text-gray-300">First Logged</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
      </div>

      <div id="cal-heatmap" class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md dark:shadow-gray-700/30 flex justify-center"></div>
    </div>
  `;

  drawSummaryStats(habit);
  drawHabitHeatmap(habit);
  setupThemeToggle();
}

type HeatmapDatum = {
  t: number;       // timestamp in milliseconds (start of day)
  p: number;       // count
  v: string;       // habit name
};

async function fetchHabit(habit: string): Promise<HeatmapDatum[]> {
  const res = await fetch(`/api/habits/${habit}`, { credentials: 'include' });
  const json = await res.json();

  const counts: Record<number, number> = {};
  for (const entry of json.entries) {
    const timestamp = entry.timestamp * 1000; // Convert to milliseconds
    counts[timestamp] = (counts[timestamp] || 0) + 1;
  }

  const result: HeatmapDatum[] = Object.entries(counts).map(([t, p]) => ({
    t: Number(t),
    p,
    v: habit,
  }));

  return result;
}

async function fetchHabitSummary(habit: string): Promise<any> {
  const res = await fetch(`/api/habits/${habit}/summary`, { credentials: 'include' });
  if (!res.ok) {
    throw new Error(`Failed to fetch summary for habit ${habit}: ${res.statusText}`);
  }
  return res.json();
}

async function fetchHabits(): Promise<string[]> {
  const res = await fetch('/api/habits', { credentials: 'include' });
  if (!res.ok) {
    throw new Error(`Failed to fetch habits: ${res.statusText}`);
  }
  const data = await res.json();
  return data.habits;
}

async function fetchVersionInfo(): Promise<{ Version: string; BuildDate: string }> {
  const res = await fetch('/version');
  if (!res.ok) {
    throw new Error(`Failed to fetch version info: ${res.statusText}`);
  }
  return res.json();
}

async function drawHabitHeatmap(habit: string) {
  const data = await fetchHabit(habit);
  const darkMode = getCurrentTheme() === 'dark';

  // Clear existing heatmap before redrawing
  const heatmapContainer = document.querySelector('#cal-heatmap');
  if (heatmapContainer) {
    heatmapContainer.innerHTML = '';
  }

  const cal = new CalHeatmap();
  cal.paint(
    {
      itemSelector: "#cal-heatmap",
      range: 12,
      domain: {
        type: 'month',
        label: {
          position: 'top',
          text: 'MMM',
          textColor: darkMode ? '#ffffff' : '#374151',
        },
      },
      subDomain: {
        type: 'day',
        radius: 2,
        width: 13,
        height: 13,
      },
      date: {
        start: new Date(new Date().getFullYear(), 0, 1),
        end: new Date(new Date().getFullYear(), 11, 31),
      },
      data: {
        source: data,
        type: 'json',
        x: 't',
        y: 'p',
      },
      scale: {
        color: {
          range: darkMode
            ? ['#374151', '#22c55e']
            : ['#e5e7eb', '#22c55e'],
          domain: [0, 1],
        },
      },
    },
    [
      [
        Tooltip,
        {
          text: function (): string {
            return (
              'hello!'
            );
          },
        },
      ],
    ]
  );
}

function toTitleCase(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function getHabitFromURL(): string | null {
  const parts = window.location.pathname.split('/');
  if (parts.length >= 3 && parts[1] === "habits") {
    return parts[2];
  }
  return null;
}

async function drawSummaryStats(id: string) {
  const data = await fetchHabitSummary(id);

  const updateStat = (stat: string, value: string | number) => {
    const container = document.querySelector(`[data-stat="${stat}"]`);
    if (container) {
      const valueElement = container.querySelector('.text-2xl');
      if (valueElement) {
        valueElement.textContent = value.toString();
      }
    }
  };

  updateStat('current-streak', `${data.habit_summary.current_streak} days`);
  updateStat('longest-streak', `${data.habit_summary.longest_streak} days`);
  updateStat('month-progress', computeDaysThisMonthAsPercentage(data.habit_summary.this_month));
  updateStat('total-days', data.habit_summary.total_days_done);
  updateStat('best-month', intToMonth(data.habit_summary.best_month));
  updateStat('first-logged', new Date(data.habit_summary.first_logged * 1000).toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric'
  }));
}

function computeDaysThisMonthAsPercentage(daysThisMonth: number): string {
  const today = new Date();
  const totalDaysInMonth = new Date(today.getFullYear(), today.getMonth() + 1, 0).getDate();
  const percentage = (daysThisMonth / totalDaysInMonth) * 100;
  return `${Math.round(percentage)}%`;
}

function intToMonth(month: number): string {
  const months = [
    "Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"
  ];
  return months[month-1];
}

async function drawHabitsList() {
  document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
    <div class="max-w-5xl mx-auto p-6">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">My Habits</h1>
        ${createThemeToggle()}
      </div>
      <div class="grid gap-4">
        <div id="habits-list" class="bg-white dark:bg-gray-800 rounded-lg shadow-md dark:shadow-gray-700/30 divide-y dark:divide-gray-700">
        </div>
      </div>
    </div>
  `;

  try {
    const habits = await fetchHabits();
    const habitsList = document.querySelector('#habits-list')!;

    if (habits.length === 0) {
      habitsList.innerHTML = `
        <div class="p-4 text-gray-600 dark:text-gray-400">
          No habits tracked yet. Start by tracking your first habit!
        </div>
      `;
      return;
    }

    // Fetch summaries for all habits
    const summaries = await Promise.all(
      habits.map(async habit => ({
        name: habit,
        summary: await fetchHabitSummary(habit)
      }))
    );

    habitsList.innerHTML = summaries
      .sort((a, b) => a.name.localeCompare(b.name))
      .map(({ name, summary }) => `
        <a href="/habits/${name}"
           class="block p-4 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150">
          <div class="flex items-center justify-between">
            <span class="text-lg font-medium text-gray-900 dark:text-white">
              ${toTitleCase(name)}
              ${summary.habit_summary.current_streak > 1 ? '🔥' : ''}
            </span>
            <svg class="w-5 h-5 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </a>
      `)
      .join('');

  } catch (error) {
    console.error('Failed to fetch habits:', error);
    document.querySelector('#habits-list')!.innerHTML = `
      <div class="p-4 text-red-600 dark:text-red-400">
        Failed to load habits. Please try again later.
      </div>
    `;
  }

  setupThemeToggle();
}

async function drawHabitFooter() {
  try {
    const versionInfo = await fetchVersionInfo();
    const footer = document.createElement('div');
    var link = "https://github.com/brk3/habits/commits/main"
    footer.className = 'text-right max-w-5xl mx-auto mt-8 mb-4 px-6 text-xs text-gray-400 dark:text-gray-500';
    footer.innerHTML = `
      <a href="${link}" class="hover:text-gray-600 dark:hover:text-gray-300"
         target="_blank" rel="noopener noreferrer">
        ${versionInfo.Version}
      </a>`;
    document.querySelector('#app')?.appendChild(footer);
  } catch (error) {
    console.error('Failed to load version info:', error);
  }
}

async function main() {
  // Initialize theme before anything else
  const storedTheme = getStoredTheme();
  applyTheme(storedTheme);
  initializeBodyStyles();

  const habit = getHabitFromURL();
  if (!habit) {
    await drawHabitsList();
  } else {
    drawHabitSummary(habit);
  }

  await drawHabitFooter();
}

main()
