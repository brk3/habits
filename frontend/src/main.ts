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
      <div class="flex justify-between items-center mb-6">
        <div id="title" class="text-3xl font-bold text-gray-900 dark:text-white">${toTitleCase(habit)}</div>
        ${createThemeToggle()}
      </div>

      <!-- Top row of 3 cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
        <div class="bg-white dark:bg-gray-700 p-4 rounded-lg shadow-lg dark:shadow-xl" data-stat="current-streak">
          <div class="text-lg text-gray-700 dark:text-gray-300">Current Streak</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-700 p-4 rounded-lg shadow-lg dark:shadow-xl" data-stat="longest-streak">
          <div class="text-lg text-gray-700 dark:text-gray-300">Longest Streak</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-700 p-4 rounded-lg shadow-lg dark:shadow-xl" data-stat="month-progress">
          <div class="text-lg text-gray-700 dark:text-gray-300">This Month</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
      </div>

      <!-- Second row of 3 cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div class="bg-white dark:bg-gray-700 p-4 rounded-lg shadow-lg dark:shadow-xl" data-stat="total-days">
          <div class="text-lg text-gray-700 dark:text-gray-300">Total Days Done</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-700 p-4 rounded-lg shadow-lg dark:shadow-xl" data-stat="best-month">
          <div class="text-lg text-gray-700 dark:text-gray-300">Best Month</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
        <div class="bg-white dark:bg-gray-700 p-4 rounded-lg shadow-lg dark:shadow-xl" data-stat="first-logged">
          <div class="text-lg text-gray-700 dark:text-gray-300">First Logged</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
        </div>
      </div>

      <div id="cal-heatmap" class="bg-white dark:bg-gray-700 p-6 rounded-lg shadow-lg dark:shadow-xl flex justify-center"></div>
    </div>
  `;

  drawSummaryStats(habit);
  drawHabitHeatmap(habit);
  setupThemeToggle(() => {
    // Refresh heatmap when theme changes
    drawHabitHeatmap(habit);
  });
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

async function drawHabitsList() {
  document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
    <div class="max-w-5xl mx-auto p-6">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">My Habits</h1>
        ${createThemeToggle()}
      </div>
      <div class="grid gap-4">
        <div id="habits-list" class="bg-white dark:bg-gray-700 rounded-lg shadow-lg dark:shadow-xl divide-y dark:divide-gray-600">
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
