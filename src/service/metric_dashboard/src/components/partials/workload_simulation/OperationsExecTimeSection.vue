<script setup>
import { ref, onMounted, reactive, onBeforeMount, onUpdated } from 'vue'
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
import axios from 'axios'

const props = defineProps(['traceId'])
const reportData = ref(null)
const chartInstance = ref(null)

const labels = {
  'input_message_request': 'Input Message API',
  'relay_input_message': 'Relay Input Message',
  'build_push_notification_message': 'Build Notification Messages',
  'send_push_notification': 'Send Push Notifications',
}

const queryExecTimeSpansReport = async function() {
  try {
    const report = {}
    var listRequest = await axios.get(`http://localhost:8003/metric/workload/${props.traceId}/operations`)
    var timeSpansRequest = await axios.get(`http://localhost:8003/metric/workload/${props.traceId}/operations/report-execution-time-spans`)
    listRequest.data.data.forEach(opt => {
      if (report[opt.service_name] == undefined) {
        report[opt.service_name] = {}
      }
      report[opt.service_name][opt.service_operation] = {
        start_latency_ms: 0,
        duration_ms: 1,
      }
    })
    timeSpansRequest.data.data.forEach(opt => {
      var start = opt.operation_start_latency_ms
      var end = opt.operation_end_latency_ms
      var duration = end - start
      report[opt.service_name][opt.service_operation].start_latency_ms = start
      report[opt.service_name][opt.service_operation].duration_ms = duration
    })
    return report
  }
  catch (err) {
    console.log('Failed to query services operation execution time report.')
    console.log(err?.response?.data ?? err)
    return null
  }
}

const renderGanttChart = function() {
  const data = reportData.value
  const tasks = []
  const chartCanvas = document.getElementById('chart-canvas')

  if (data == null) {
    return
  }
  for (const svName in data) {
    for (const optName in data[svName]) {
      tasks.push({
        label: labels[optName],
        start: data[svName][optName]['start_latency_ms'],
        duration: data[svName][optName]['duration_ms'],
      })
    }
  }
  if (chartInstance.value != null) {
    chartInstance.value.destroy()
  }
  const barHeight = 30
  const tickHeight = 75
  const barMargin = 10
  chartCanvas.height = tickHeight + tasks.length * (barHeight + 2 * barMargin)

  chartInstance.value = new Chart(chartCanvas.getContext('2d'), {
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
          backgroundColor: 'rgba(0, 0, 0, 1)',
          stack: 'gantt',
        }
      ]
    },
    options: {
      indexAxis: 'y',
      responsive: true,
      categoryPercentage: 1,
      barPercentage: 0.8,
      scales: {
        x: {
          beginAtZero: true,
          ticks: {
            stepSize: 100,
          },
        },
        y: {
          ticks: {
            crossAlign: "far",
            // display: false
          },
          grid: {
            // drawTicks: false,
            // drawOnChartArea: true
          }
        }
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          filter: (tooltipItem) => tooltipItem.datasetIndex === 1,
          callbacks: {
            label: function (context) {
              const label = context.dataset.label || ''
              const value = context.parsed.x
              const valueAsMs = Intl.NumberFormat('en-US').format(value)
              const valueAsSeconds = Math.round(value/1000)
              return `${label}: ${valueAsMs}ms - ${valueAsSeconds}s`
            }
          }
        },
        datalabels: {
          display: () => false
        }
      }
    }
  })
}

onMounted(async () => {
  Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Title, ChartDataLabels)
  reportData.value = await queryExecTimeSpansReport()
  renderGanttChart()
})

onUpdated(async () => {
  reportData.value = await queryExecTimeSpansReport()
  renderGanttChart()
})

</script>

<template>
  <div id="operations-execution-time-section">
    <div class="card shadow-md p-4">
      <h5 class="mb-4 fs-6 d-flex align-items-center">
        <IconHourGlass class="me-2" width="24" fill="black" /> Service Operations Execution Time
      </h5>
      <!-- <p v-if="operations.length == 0" class="text-center text-secondary">No service operations found</p> -->
      <div id="gantt-wrapper" class="d-flex">
        <div id="canvas-container">
          <canvas id="chart-canvas"></canvas>
        </div>
      </div>
    </div>
  </div>
</template>
