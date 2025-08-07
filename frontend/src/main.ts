//import './style.css'
// @ts-expect-error: No type definitions for 'cal-heatmap'
import CalHeatmap from 'cal-heatmap';
import 'cal-heatmap/cal-heatmap.css';

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div id="title"></div>
  <div id="cal-heatmap"></div>
`;

type HeatmapDatum = {
  t: number;       // timestamp in milliseconds (start of day)
  p: number;       // count
  v: string;       // habit name
};

async function fetchHabitData(habit: string): Promise<HeatmapDatum[]> {
  console.log(`Fetching data for habit: ${habit}`);
  const res = await fetch(`http://localhost:8080/habits/${habit}`);
  const json = await res.json();
  console.log("Fetched data:", json);

  const counts: Record<number, number> = {};
  for (const entry of json.entries) {
    const timestamp = entry.TimeStamp * 1000; // Convert to milliseconds
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
      width: 15,
      height: 15,
      gutter: 5,
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
        scheme: 'YlGn',
        domain: [0, 5],
      },
    },
  });
}

function toTitleCase(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

const habit = "guitar";
document.querySelector<HTMLDivElement>('#title')!.innerHTML = `<h1>${toTitleCase(habit)}</h1>`;
drawHabitHeatmap(habit);