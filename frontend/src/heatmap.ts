// @ts-expect-error: No type definitions for 'cal-heatmap'
import CalHeatmap from 'cal-heatmap';
// @ts-expect-error: No type definitions for 'cal-heatmap'
import Tooltip from 'cal-heatmap/plugins/Tooltip';
import { fetchHabit } from './api';
import { getCurrentTheme } from './theme';

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

export { drawHabitHeatmap };