// API types
export type HeatmapDatum = {
  t: number;       // timestamp in milliseconds (start of day)
  p: number;       // count
  v: string;       // habit name
};

// API functions
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

export { fetchHabit, fetchHabitSummary, fetchHabits, fetchVersionInfo };