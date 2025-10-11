// @ts-expect-error: No type definitions for 'cal-heatmap'
import CalHeatmap from 'cal-heatmap';
// @ts-expect-error: No type definitions for 'cal-heatmap'
import Tooltip from 'cal-heatmap/plugins/Tooltip';
import { fetchHabit, fetchHabitEntries, type HabitEntry } from './api';
import { getCurrentTheme } from './theme';

function formatDate(timestamp: number): string {
  return new Date(timestamp).toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
}

async function showEntriesForDate(habit: string, timestamp: number) {
  const entriesContainer = document.querySelector('#habit-entries');
  if (!entriesContainer) return;

  // Show the container with card styling
  entriesContainer.className = 'bg-white dark:bg-gray-700 p-6 rounded-lg shadow-lg dark:shadow-xl mt-4';
  entriesContainer.innerHTML = '<div class="text-gray-600 dark:text-gray-400">Loading...</div>';

  try {
    const allEntries = await fetchHabitEntries(habit);

    // Filter entries for this specific day (timestamp is in seconds from API)
    const dayStart = timestamp / 1000; // Convert to seconds
    const dayEnd = dayStart + 86400; // Add 24 hours

    const entriesForDay = allEntries.filter((entry: HabitEntry) =>
      entry.timestamp >= dayStart && entry.timestamp < dayEnd
    );

    if (entriesForDay.length === 0) {
      entriesContainer.innerHTML = `
        <div class="text-gray-600 dark:text-gray-400">
          No entries for this date
        </div>
      `;
      return;
    }

    entriesContainer.innerHTML = `
      <div class="space-y-2">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
          ${formatDate(timestamp)}
        </h3>
        ${entriesForDay.map((entry: HabitEntry) => {
          const entryTime = new Date(entry.timestamp * 1000).toLocaleTimeString('en-US', {
            hour: 'numeric',
            minute: '2-digit',
            hour12: true
          });
          return `
            <div class="bg-gray-50 dark:bg-gray-600 p-3 rounded">
              <div class="flex items-start gap-3">
                <span class="text-gray-500 dark:text-gray-400 text-sm font-medium min-w-[60px]">${entryTime}</span>
                <div class="flex-1">
                  ${entry.note && entry.note.trim() !== ''
                    ? `<p class="text-gray-900 dark:text-white">${entry.note}</p>`
                    : `<p class="text-gray-500 dark:text-gray-400 italic">No note</p>`
                  }
                </div>
              </div>
            </div>
          `;
        }).join('')}
      </div>
    `;
  } catch (error) {
    console.error('Failed to fetch entries:', error);
    entriesContainer.innerHTML = `
      <div class="text-red-600 dark:text-red-400">
        Failed to load entries. Please try again.
      </div>
    `;
  }
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
            ? ['#374151', '#166534', '#22c55e', '#86efac']
            : ['#e5e7eb', '#bbf7d0', '#4ade80', '#22c55e'],
          type: 'threshold',
          domain: [1, 2, 3],
        },
      },
    },
    [
      [
        Tooltip,
        {
          text: function (timestamp: number, value: number): string {
            if (!value) {
              return new Date(timestamp).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric'
              });
            }
            const date = new Date(timestamp).toLocaleDateString('en-US', {
              month: 'short',
              day: 'numeric',
              year: 'numeric'
            });
            return `${date}: ${value} ${value === 1 ? 'entry' : 'entries'}`;
          },
        },
      ],
    ]
  );

  // Add click handler to heatmap cells
  cal.on('click', (event: any, timestamp: number, value: number) => {
    if (value > 0) {
      showEntriesForDate(habit, timestamp);
    }
  });
}

export { drawHabitHeatmap };