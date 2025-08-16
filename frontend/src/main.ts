import './style.css'
// @ts-expect-error: No type definitions for 'cal-heatmap'
import CalHeatmap from 'cal-heatmap';
import 'cal-heatmap/cal-heatmap.css';

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div class="max-w-5xl mx-auto p-6">
    <!-- Title -->
    <div id="title" class="text-3xl font-bold mb-6"></div>

    <!-- Top row of 3 cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-lg">🔥 Current Streak</div>
        <div class="text-2xl font-bold">7 days</div>
      </div>
      <div class="bg-white  p-4 rounded-lg shadow">
        <div class="text-lg">🏅 Longest Streak</div>
        <div class="text-2xl font-bold">14 days</div>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-lg">📅 This Month: 15 / 31</div>
        <div class="text-2xl font-bold">48%</div>
      </div>
    </div>

    <!-- Second row of 3 cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-lg">Total Days Done</div>
        <div class="text-2xl font-bold">212</div>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-lg">Best Month</div>
        <div class="text-2xl font-bold">July 2025</div>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-lg">First Logged</div>
        <div class="text-2xl font-bold">Jan 14, 2025</div>
      </div>
    </div>

    <div id="cal-heatmap" class="bg-white p-6 rounded-lg shadow-md flex justify-center"></div>
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

async function drawHabitHeatmap(habit: string) {
  const data = await fetchHabitData(habit);
  console.log("Data for heatmap:", data);

  const timestamps = data.map(d => d.t);
  const earliest = new Date(Math.min(...timestamps));
  console.log("Start date for heatmap:", earliest.toISOString());

  const cal = new CalHeatmap();
  cal.paint({
    itemSelector: "#cal-heatmap",
    range: 12,
    domain: {
      type: 'month',
      label: {
        position: 'top',
        text: 'MMM',
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
        range: ['#e5e7eb', '#22c55e'], // missed, done
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

const habit = getHabitFromURL();
if (!habit) {
  console.error("No habit found in URL");
} else {
  document.querySelector<HTMLHeadingElement>('#title')!.innerHTML = `
    ${toTitleCase(habit)}
  `;
  drawHabitHeatmap(habit);
}