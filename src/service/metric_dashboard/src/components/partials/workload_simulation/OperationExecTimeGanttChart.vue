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

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Title, ChartDataLabels)

const chartCanvas = ref(null)
const chartInstance = ref(null)

const tasks = [
  { label: 'Input Message API', start: 100, duration: 300 },
  { label: 'Build Notification Messages', start: 200, duration: 500 },
  { label: 'Send Push Notifications', start: 0, duration: 600 }
]

const colors = ['#7A99FF', '#7A99FF', '#7A99FF']

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
          backgroundColor: colors,
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
          filter: (tooltipItem) => tooltipItem.datasetIndex === 1
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
  <div id="gantt-wrapper">
    <div id="labels">
      <div class="label-box label-0"><span>Input Message API</span></div>
      <div class="label-box label-1"><span>Build Notification Messages</span></div>
      <div class="label-box label-2"><span>Send Push Notifications</span></div>
    </div>
    <canvas ref="chartCanvas"></canvas>
  </div>
</template>
