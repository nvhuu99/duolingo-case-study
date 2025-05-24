<script setup>
import { ref, onMounted, watch, toRef, computed } from 'vue'
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
const traceId = toRef(props.traceId);
const chartInstance = ref(null)

const isChartReady = computed(() => chartInstance.value != null)

const destroyChart = async function() {
  if (chartInstance.value != null) await chartInstance.value.destroy()
  chartInstance.value = null
}

const renderChart = async function() {
  const report = await buildOperationExecTimeSpansReport()
  if (report == null) return

  // Build chart data
  const OPERATION_LABELS = {
    'input_message_request': 'Input Message API',
    'relay_input_message': 'Relay Input Message',
    'build_push_notification_message': 'Build Notification Messages',
    'send_push_notification': 'Send Push Notifications',
  }
  const chartData = []
  for (const svName in report) {
    for (const optName in report[svName]) {
      chartData.push({
        label: OPERATION_LABELS[optName] ?? "Unknown Operation",
        start: report[svName][optName]['start_latency_ms'],
        duration: report[svName][optName]['duration_ms'],
      })
    }
  }

  // Config and render chart
  const chartCanvas = document.getElementById('chart-canvas')
  chartCanvas.width = document.getElementById('gantt-wrapper').clientWidth
  chartCanvas.height = 90 + chartData.length * 50 // tick-height + bar-heights (margin included)
  chartInstance.value = new Chart(chartCanvas.getContext('2d'), {
    type: 'bar',
    data: {
      labels: chartData.map(t => t.label),
      datasets: [
        {
          label: 'Offset',
          data: chartData.map(t => t.start),
          backgroundColor: 'transparent',
          stack: 'gantt',
          datalabels: { display: false },
        },
        {
          label: 'Duration',
          data: chartData.map(t => t.duration),
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
          // ticks: {
          //   stepSize: 500,
          // },
          beginAtZero: true,
          title: { display: true, text: 'Captured time in millisecond (relative to the workload start time)' }
        },
        y: {
          ticks: {
            crossAlign: "far",
          },
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

onMounted(() => {
  Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Title, ChartDataLabels)
  renderChart()
})

watch(traceId, async () => {
  await destroyChart()
  renderChart()
});

const buildOperationExecTimeSpansReport = async function() {
  try {
    // Fetch the list of services operation, and set default time span values
    const report = {}
    var listRequest = await axios.get(`http://localhost:8003/metric/workload/${traceId.value}/list-operations`)
    listRequest.data.data.forEach(opt => {
      if (!report[opt.service_name]) report[opt.service_name] = {}
      report[opt.service_name][opt.service_operation] = { start_latency_ms: 0, duration_ms: 1 }
    })
    // Fetch and merge the time spans into the report
    var timeSpansRequest = await axios.get(`http://localhost:8003/metric/workload/${traceId.value}/service-execution-time-spans`)
    timeSpansRequest.data.data.forEach(opt => report[opt.service_name][opt.service_operation] = {
      start_latency_ms: opt.operation_start_latency_ms,
      duration_ms: opt.operation_end_latency_ms - opt.operation_start_latency_ms
    })
    return report
  }
  catch (err) {
    console.log('Failed to query services operation execution time report.')
    console.log(err?.response?.data ?? err)
    return null
  }
}
</script>

<template>
  <div id="operations-execution-time-section">
    <div class="card shadow-md p-4">
      <h5 class="mb-4 fs-6 d-flex align-items-center">
        <IconHourGlass class="me-2" width="24" fill="black" /> Service Operations Execution Time
      </h5>
      <p v-show="!isChartReady" class="text-center text-secondary">Failed to query services operation execution time report.</p>
      <div id="gantt-wrapper">
        <div id="chart-container" :class="isChartReady ? '' : 'd-none'">
          <canvas id="chart-canvas"></canvas>
        </div>
      </div>
    </div>
  </div>
</template>
