import './style.css'
import 'cal-heatmap/cal-heatmap.css';
import { getStoredTheme, applyTheme, createThemeToggle, setupThemeToggle } from './theme';
import { fetchHabitSummary, fetchHabits, fetchVersionInfo } from './api';
import { drawHabitHeatmap } from './heatmap';
import { toTitleCase, getHabitFromURL, computeDaysThisMonthAsPercentage, intToMonth } from './utils';

function initializeBodyStyles() {
  document.body.className = 'bg-gray-50 dark:bg-gray-900 min-h-screen transition-colors duration-200';
}

function drawHabitSummary(habit: string) {
  document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
    <div class="max-w-5xl mx-auto p-6">
      <div class="flex justify-between items-center mb-4">
        <a href="/" class="inline-flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          <span>All Habits</span>
        </a>
        ${createThemeToggle()}
      </div>
      <div class="flex justify-between items-center mb-8">
        <div id="title" class="text-4xl font-bold text-gray-900 dark:text-white">${toTitleCase(habit)}</div>
        <button id="track-habit-btn" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors shadow-sm">
          Add Entry
        </button>
      </div>

      <!-- Top row of 3 cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
        <div class="stat-card-streak p-5 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800" data-stat="current-streak">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-2">Current Streak</div>
          <div class="text-3xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="stat-card-longest p-5 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800" data-stat="longest-streak">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-2">Longest Streak</div>
          <div class="text-3xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="stat-card-month p-5 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800" data-stat="month-progress">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-2">This Month</div>
          <div class="text-3xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
      </div>

      <!-- Second row of 3 cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div class="stat-card-total p-5 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800" data-stat="total-days">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-2">Total Days</div>
          <div class="text-3xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="stat-card-best p-5 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800" data-stat="best-month">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-2">Best Month</div>
          <div class="text-3xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="stat-card-last p-5 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800" data-stat="last-logged">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-2">Last Logged</div>
          <div class="text-3xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
      </div>

      <div id="cal-heatmap" class="flex justify-center mb-6"></div>

      <!-- Entries list shown when clicking a day -->
      <div id="habit-entries" class="hidden"></div>
    </div>
  `;

  drawSummaryStats(habit);
  drawHabitHeatmap(habit);
  setupThemeToggle(() => {
    // Refresh heatmap when theme changes
    drawHabitHeatmap(habit);
  });

  // Setup Track Habit button
  const trackBtn = document.querySelector('#track-habit-btn');
  if (trackBtn) {
    trackBtn.addEventListener('click', () => showTrackHabitForm(habit));
  }
}

function showTrackHabitForm(habit: string) {
  const isDark = document.documentElement.classList.contains('dark');

  const colors = isDark ? {
    cardBg: '#1f2937',
    cardBorder: '#374151',
    title: '#ffffff',
    label: '#d1d5db',
    inputBg: '#374151',
    inputBorder: '#4b5563',
    inputText: '#ffffff',
    cancelBg: '#374151',
    cancelText: '#d1d5db',
    cancelBorder: '#4b5563',
  } : {
    cardBg: '#ffffff',
    cardBorder: '#e5e7eb',
    title: '#111827',
    label: '#4b5563',
    inputBg: '#ffffff',
    inputBorder: '#d1d5db',
    inputText: '#111827',
    cancelBg: '#ffffff',
    cancelText: '#4b5563',
    cancelBorder: '#d1d5db',
  };

  const formHtml = `
    <div id="track-form-overlay" class="fixed inset-0 flex items-center justify-center z-50" style="background-color: rgba(0, 0, 0, 0.5) !important;">
      <div class="rounded-xl shadow-xl p-6 max-w-md w-full mx-4 border" style="background-color: ${colors.cardBg} !important; border-color: ${colors.cardBorder} !important;">
        <h3 class="text-2xl font-bold mb-4" style="color: ${colors.title} !important;">Track ${toTitleCase(habit)}</h3>

        <form id="track-habit-form">
          <div class="mb-4">
            <label for="note" class="block text-sm font-medium mb-2" style="color: ${colors.label} !important;">
              Note (optional)
            </label>
            <textarea
              id="note"
              name="note"
              rows="3"
              class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              style="background-color: ${colors.inputBg} !important; border-color: ${colors.inputBorder} !important; color: ${colors.inputText} !important;"
              placeholder="Add a note about this entry..."
            ></textarea>
          </div>

          <div class="flex gap-3 justify-end">
            <button
              type="button"
              id="cancel-btn"
              class="px-4 py-2 rounded-lg transition-colors border"
              style="color: ${colors.cancelText} !important; border-color: ${colors.cancelBorder} !important; background-color: ${colors.cancelBg} !important;"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-4 py-2 font-medium rounded-lg transition-colors"
              style="background-color: #2563eb !important; color: #ffffff !important;"
            >
              Track Now
            </button>
          </div>
        </form>
      </div>
    </div>
  `;

  document.body.insertAdjacentHTML('beforeend', formHtml);

  const overlay = document.querySelector('#track-form-overlay');
  const form = document.querySelector('#track-habit-form') as HTMLFormElement;
  const cancelBtn = document.querySelector('#cancel-btn');

  const closeForm = () => {
    overlay?.remove();
  };

  cancelBtn?.addEventListener('click', closeForm);
  overlay?.addEventListener('click', (e) => {
    if (e.target === overlay) closeForm();
  });

  form?.addEventListener('submit', async (e) => {
    e.preventDefault();

    const noteInput = document.querySelector('#note') as HTMLTextAreaElement;
    const note = noteInput?.value || '';

    try {
      const response = await fetch('/api/habits/', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          name: habit,
          note: note,
          timestamp: Math.floor(Date.now() / 1000), // Current time in seconds
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to track habit');
      }

      closeForm();

      window.location.reload();
    } catch (error) {
      console.error('Error tracking habit:', error);
      alert('Failed to track habit. Please try again.');
    }
  });
}

async function drawSummaryStats(id: string) {
  const data = await fetchHabitSummary(id);

  const updateStat = (stat: string, value: string | number) => {
    const container = document.querySelector(`[data-stat="${stat}"]`);
    if (container) {
      const valueElement = container.querySelector('.text-3xl');
      if (valueElement) {
        valueElement.textContent = value.toString();
      }
    }
  };

  updateStat('current-streak', `${data.habit_summary.current_streak} days`);
  updateStat('longest-streak', `${data.habit_summary.longest_streak} days`);
  updateStat('month-progress', computeDaysThisMonthAsPercentage(data.habit_summary.this_month));
  updateStat('total-days', data.habit_summary.total_days_done);
  updateStat('best-month', `${intToMonth(data.habit_summary.best_month)} ${new Date().getFullYear()}`);
  updateStat('last-logged', new Date(data.habit_summary.last_write * 1000).toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric'
  }));

  const fadeCard = (statName: string, isEmpty: boolean) => {
    const card = document.querySelector(`[data-stat="${statName}"]`);
    if (card) {
      if (isEmpty) {
        card.classList.add('opacity-50');
      } else {
        card.classList.remove('opacity-50');
      }
    }
  };

  fadeCard('current-streak', data.habit_summary.current_streak === 0);
  fadeCard('longest-streak', data.habit_summary.longest_streak === 0);
  fadeCard('month-progress', data.habit_summary.this_month === 0);
  fadeCard('total-days', data.habit_summary.total_days_done === 0);
  fadeCard('best-month', data.habit_summary.best_month === 0);
  fadeCard('last-logged', data.habit_summary.last_write === 0);
}

async function drawHabitsList() {
  document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
    <div class="max-w-5xl mx-auto p-6">
      <div class="flex justify-between items-center mb-8">
        <h1 class="text-4xl font-bold text-gray-900 dark:text-white">My Habits</h1>
        ${createThemeToggle()}
      </div>
      <div class="grid gap-4">
        <div id="habits-list" class="bg-white dark:bg-gray-800 rounded-xl shadow-md divide-y dark:divide-gray-700">
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
