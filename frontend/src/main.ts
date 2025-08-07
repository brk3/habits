//import './style.css'
import CalHeatmap from 'cal-heatmap';
import 'cal-heatmap/cal-heatmap.css';


document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div id="cal-heatmap"></div>
`;

const cal = new CalHeatmap();
cal.paint({
  itemSelector: "#cal-heatmap"
});