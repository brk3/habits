import './style.css'
// @ts-expect-error: No type definitions for 'cal-heatmap'
import CalHeatmap from 'cal-heatmap';
import 'cal-heatmap/cal-heatmap.css';

// Set the body background to adapt to dark mode
document.body.className = 'bg-gray-50 dark:bg-gray-900 min-h-screen transition-colors duration-200';

// Add custom CSS for cal-heatmap month labels in dark mode
const style = document.createElement('style');
style.textContent = `
  @media (prefers-color-scheme: dark) {
    .ch-domain-text {
      fill: #ffffff !important;
      color: #ffffff !important;
    }
  }
  
  @media (prefers-color-scheme: light) {
    .ch-domain-text {
      fill: #374151 !important;
      color: #374151 !important;
    }
  }
`;
document.head.appendChild(style);

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div class="max-w-5xl mx-auto p-6">
    <!-- Title -->
    <div id="title" class="text-3xl font-bold mb-6 text-gray-900 dark:text-white"></div>

    <!-- Top row of 3 cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
      <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30">
        <div class="text-lg text-gray-700 dark:text-gray-300">🔥 Current Streak</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">7 days</div>
      </div>
      <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30">
        <div class="text-lg text-gray-700 dark:text-gray-300">🏅 Longest Streak</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">14 days</div>
      </div>
      <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30">
        <div class="text-lg text-gray-700 dark:text-gray-300">📅 This Month: 15 / 31</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">48%</div>
      </div>
    </div>

    <!-- Second row of 3 cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
      <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30">
        <div class="text-lg text-gray-700 dark:text-gray-300">Total Days Done</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">212</div>
      </div>
      <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30">
        <div class="text-lg text-gray-700 dark:text-gray-300">Best Month</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">July 2025</div>
      </div>
      <div class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow dark:shadow-gray-700/30">
        <div class="text-lg text-gray-700 dark:text-gray-300">First Logged</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">Jan 14, 2025</div>
      </div>
    </div>

    <div id="cal-heatmap" class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md dark:shadow-gray-700/30 flex justify-center"></div>
  </div>
`;

type HeatmapDatum = {
  t: number;       // timestamp in milliseconds (start of day)
  p: number;       // count
  v: string;       // habit name
};

async function fetchHabitData(habit: string): Promise<HeatmapDatum[]> {
  console.log(`Fetching data for habit: ${habit}`);
  const res = await fetch(`/api/habits/${habit}`);
  console.log("Response status:", res);
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

function isDarkMode(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

async function drawHabitHeatmap(habit: string) {
  const data = await fetchHabitData(habit);
  console.log("Data for heatmap:", data);

  const timestamps = data.map(d => d.t);
  const earliest = new Date(Math.min(...timestamps));
  console.log("Start date for heatmap:", earliest.toISOString());

  const darkMode = isDarkMode();
  
  const cal = new CalHeatmap();
  cal.paint({
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
      start: new Date(earliest),
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
          ? ['#374151', '#22c55e'] // Dark mode: darker gray to green
          : ['#e5e7eb', '#22c55e'], // Light mode: light gray to green
        domain: [0, 1],
      },
    },
  });
}

function toTitleCase(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function getHabitFromURL(): string | null {
  const parts = window.location.pathname.split('/');
  console.log("URL parts:", parts);
  if (parts.length >= 3 && parts[1] === "habits") {
    return parts[2];
  }
  return null;
}

// Listen for changes in color scheme preference
function setupDarkModeListener() {
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  
  mediaQuery.addEventListener('change', (e) => {
    console.log('Color scheme changed:', e.matches ? 'dark' : 'light');
    // Re-render the heatmap with new colors
    const habit = getHabitFromURL();
    if (habit) {
      // Clear existing heatmap
      document.querySelector('#cal-heatmap')!.innerHTML = '';
      drawHabitHeatmap(habit);
    }
  });
}

const habit = getHabitFromURL();
if (!habit) {
  console.error("No habit found in URL");
} else {
  document.querySelector<HTMLHeadingElement>('#title')!.innerHTML = `
    ${toTitleCase(habit)}
  `;
  drawHabitHeatmap(habit);
  setupDarkModeListener();
}