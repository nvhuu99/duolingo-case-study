<script setup>
import { ref, onMounted } from 'vue'
import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Title
} from 'chart.js'
import ChartDataLabels from 'chartjs-plugin-datalabels'
import IconHourGlass from '@/components/icons/IconHourGlass.vue'

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Title, ChartDataLabels)

const chartCanvas = ref(null)
const chartInstance = ref(null)

const tasks = [
  { label: 'Input Message API', start: 100, duration: 300 },
  { label: 'Build Notification Messages', start: 200, duration: 500 },
  { label: 'Send Push Notifications', start: 0, duration: 600 }
]

onMounted(() => {
  const ctx = chartCanvas.value.getContext('2d')

  chartInstance.value = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: tasks.map(t => t.label),
      datasets: [
        {
          label: 'Offset',
          data: tasks.map(t => t.start),
          backgroundColor: 'transparent',
          stack: 'gantt',
          datalabels: { display: false },
        },
        {
          label: 'Duration',
          data: tasks.map(t => t.duration),
          backgroundColor: 'rgba(17, 19, 68, 1)',
          stack: 'gantt',
        }
      ]
    },
    options: {
      indexAxis: 'y',
      responsive: true,
      maintainAspectRatio: false,
      barThickness: 50,
      categoryPercentage: 1,
      scales: {
        x: {
          beginAtZero: true,
          ticks: { stepSize: 100 },
          title: { display: true, text: 'Time (ms)' }
        },
        y: {
          ticks: { display: false },
          grid: {
            drawTicks: false,
            drawOnChartArea: true
          }
        }
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          filter: (tooltipItem) => tooltipItem.datasetIndex === 1,
        },
        datalabels: {
          display: () => false
        }
      }
    }
  })
})
</script>

<template>
  <div id="operations-execution-time-section">
    <div class="card shadow-md p-4">
      <div class="mb-4 d-flex align-items-center">
        <h5 class="m-0 me-4 fs-6 d-flex align-items-center">
          <IconHourGlass class="me-2" width="24" fill="black" /> Service Operations Execution Time
        </h5>
        <div class="form-group d-flex align-items-center" style="width: 200px;">
          <label class="form-label m-0 me-2"><span style="font-size: 14px;">Aggregation:</span></label>
          <select name="" id="" class="form-select form-select-sm">
            <option value="">minimum</option>
            <option value="">maximum</option>
            <option value="">average</option>
            <option value="">median</option>
            <option value="">percentile</option>
          </select>
        </div>
      </div>
      <div id="gantt-wrapper" class="d-flex">
        <div id="chart-labels">
          <div class="label-box label-0"><span>Input Message API</span></div>
          <div class="label-box label-1"><span>Build Notification Messages</span></div>
          <div class="label-box label-2"><span>Send Push Notifications</span></div>
        </div>
        <div id="canvas-container">
          <canvas ref="chartCanvas"></canvas>
        </div>
      </div>
    </div>
  </div>
</template>
