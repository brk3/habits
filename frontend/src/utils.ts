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

export { toTitleCase, getHabitFromURL, computeDaysThisMonthAsPercentage, intToMonth };
