<script setup>
import { ref, onMounted, toRef, watch, computed, reactive } from 'vue'
import {
  Chart,
  ScatterController,
  PointElement,
  LinearScale,
  TimeScale,
  Tooltip,
  Legend
} from 'chart.js'
import 'chartjs-adapter-date-fns'
import axios from 'axios'
import IconStack from '@/components/icons/IconStack.vue'

const METRIC_TARGETS = {
  redis: { key: "redis", label: "Redis" },
  rabbitmq: { key: "rabbitmq", label: "RabbitMQ" },
}
const METRIC_TYPES = {
  all: { key: "all", label: "Show all" },
  median: { key: "median", label: "Moving Median" },
  lttb: { key: "lttb", label: "Largest Triangle (LTTB)" },
  percentiles: { key: "percentiles", label: "Percentiles" },

}
const METRIC_NAMES = {
  redis: {
    command_rate: { key: "command_rate", label: "Command execution rate" },
    lock_waited_ms: { key: "lock_waited_ms", label: "Lock waited time (ms)" },
    lock_held_ms: { key: "lock_held_ms", label: "Lock held time (ms)" },
  },
  rabbitmq: {
    published_rate: { key: "published_rate", label: "Message publishing rate" },
    delivered_rate: { key: "delivered_rate", label: "Message delivery rate" },
  },
}
const METRIC_NAME_DEFAULT = {
  redis: METRIC_NAMES.redis.command_rate.key,
  rabbitmq: METRIC_NAMES.rabbitmq.published_rate.key,
}
const REDUCTION_STEPS = [100, 200, 500, 1000, 1500, 2000, 2500, 5000]

const props = defineProps([
  'traceId',
  'workload',
])
const traceId = toRef(props, 'traceId')
const chartInstance = ref(null)
const selection = reactive({
  reduction_step: 1000,
  metric_name: METRIC_NAME_DEFAULT.redis,
  metric_type: METRIC_TYPES.all.key,
  metric_target: METRIC_TARGETS.redis.key,
})
const metricSummary = ref(null)

const isChartReady = computed(() => chartInstance.value != null)
const metricUnit = computed(() => {
  switch (selection.metric_name) {
    case METRIC_NAMES.rabbitmq.delivered_rate.key:
    case METRIC_NAMES.rabbitmq.published_rate.key:
      return ' messages';
    case METRIC_NAMES.redis.lock_waited_ms.key:
    case METRIC_NAMES.redis.lock_held_ms.key:
      return ' ms'
    case METRIC_NAMES.redis.command_rate.key:
    default:
      return ' commands';
  }
})

onMounted(() => {
  Chart.register(
    ScatterController,
    PointElement,
    LinearScale,
    TimeScale,
    Tooltip,
    Legend,
  )
  renderChart()
  updateSummary()
})

watch(() => selection.metric_target, function() {
  selection.metric_name = METRIC_NAME_DEFAULT[selection.metric_target]
})
watch([traceId, selection], async () => {
  await destroyChart()
  await renderChart()
  await updateSummary()
});


function shouldShowMetricType(type) {
  return selection.metric_type == METRIC_TYPES.all.key || selection.metric_type == type.key
}

function getSummary(type) {
  var val = metricSummary.value?.[selection.metric_target]?.[selection.metric_name]?.reduced_snapshots?.[type]?.[0].value
  if (isNaN(val)) {
    return '(unknown)'
  }
  var formated = Intl.NumberFormat('en-US').format(Math.round(val).toFixed(0))
  return formated + metricUnit.value
}

async function destroyChart() {
  if (chartInstance.value != null) await chartInstance.value.destroy()
  chartInstance.value = null
}

async function renderChart() {
  const metrics = await fetchMetrics()
  if (!metrics?.[selection.metric_target]?.[selection.metric_name]) return

  const snapshots = metrics[selection.metric_target][selection.metric_name].reduced_snapshots
  const makeChartDataSet = function(type) {
    return snapshots[type]?.map((item) => { return {
      x: new Date(item.timestamp) - new Date(props.workload.start_time),
      y: item.value,
    }})
  }

  const datasets = [
    // Median
    {
      type: 'line',
      label: 'Median',
      data: makeChartDataSet('median'),
      borderColor: '#ff69b4',
      borderWidth: 1,
      pointRadius: 1,
      fill: false,
      hidden: !shouldShowMetricType(METRIC_TYPES.median),
    },
    // Largest Triangle (LTTB)
    {
      type: 'line',
      label: 'Largest Triangle (LTTB)',
      data: makeChartDataSet('lttb'),
      borderColor: 'blue',
      borderWidth: 1,
      pointRadius: 1,
      fill: false,
      hidden: !shouldShowMetricType(METRIC_TYPES.lttb),
    },
    // 95% confidence band (fill between 5th and 95th)
    {
      type: 'line',
      label: '5th percentile',
      data: makeChartDataSet('p5'),
      backgroundColor: 'rgba(255, 133, 175, 0.2)',
      borderWidth: 0,
      pointRadius: 0,
      fill: '+1', // fill to next dataset
      hidden: !shouldShowMetricType(METRIC_TYPES.percentiles),
    },
    {
      type: 'line',
      label: '95th percentile',
      data: makeChartDataSet('p95'),
      backgroundColor: 'rgba(0, 0, 0, 0)', // no color
      borderWidth: 0,
      pointRadius: 0,
      fill: false,
      hidden: !shouldShowMetricType(METRIC_TYPES.percentiles),
    },
    // 50% confidence band (fill between 25th and 75th)
    {
      type: 'line',
      label: '25th percentile',
      data: makeChartDataSet('p25'),
      backgroundColor: 'rgba(112, 87, 255, 0.2)',
      borderWidth: 0,
      pointRadius: 0,
      fill: '+1',
      hidden: !shouldShowMetricType(METRIC_TYPES.percentiles),
    },
    {
      type: 'line',
      label: '75th percentile',
      data: makeChartDataSet('p75'),
      backgroundColor: 'rgba(0, 0, 0, 0)',
      borderWidth: 0,
      pointRadius: 0,
      fill: false,
      hidden: !shouldShowMetricType(METRIC_TYPES.percentiles),
    },
  ]

  const canvas = document.getElementById('infra-stats-chart-canvas')
  chartInstance.value = new Chart(canvas.getContext('2d'), {
    type: 'scatter',
    data: { datasets },
    options: {
      responsive: true,
      scales: {
        x: {
          title: {
            display: true,
            text: 'Captured time in millisecond (relative to the workload start time)',

          },
          ticks: {
            stepSize: selection.reduction_step,
          },
        },
        y: {
          title: { display: true, text: 'Metric value' }
        }
      },
      plugins: {
        legend: { display: true },
        tooltip: {
          callbacks: {
            label(context) {
              const asMs = Intl.NumberFormat('en-US').format(context.parsed.x);
              const asSeconds = (context.parsed.x / 1000).toFixed(1);
              const formatedVal = Intl.NumberFormat('en-US').format(Math.round(context.parsed.y).toFixed(0))
              switch (selection.metric_name) {
                case METRIC_NAMES.redis.command_rate.key:
                  return `After ${asMs}ms - ${asSeconds}seconds, the command rate is ${formatedVal} commands per ${selection.reduction_step} ms`;
                case METRIC_NAMES.redis.lock_waited_ms.key:
                  return `After ${asMs}ms - ${asSeconds}seconds, lock waited upto ${formatedVal} ms`;
                case METRIC_NAMES.redis.lock_held_ms.key:
                  return `After ${asMs}ms - ${asSeconds}seconds, lock held upto ${formatedVal} ms`;
                case METRIC_NAMES.rabbitmq.published_rate.key:
                  return `After ${asMs}ms - ${asSeconds}seconds, the message publish rate is ${formatedVal} messages for ${selection.reduction_step} ms`;
                case METRIC_NAMES.rabbitmq.delivered_rate.key:
                  return `After ${asMs}ms - ${asSeconds}seconds, the message delivery rate is ${formatedVal} messages for ${selection.reduction_step} ms`;
                default:
                  return `After ${asMs}ms - ${asSeconds} is the value of an unknown metric`;
              }
            }
          }
        },
        datalabels: {
          display: () => false
        },
      }
    }
  })
}

async function updateSummary() {
  var result = await fetchSummary()
  metricSummary.value = result
}

async function fetchMetrics() {
  try {
    var endpoint = (selection.metric_target == METRIC_TARGETS.redis.key)
                    ? `http://localhost:8003/metric/workload/${traceId.value}/redis-metrics`
                    : `http://localhost:8003/metric/workload/${traceId.value}/rabbitmq-metrics`
    var res = await axios.post(endpoint, {
        reduction_step: selection.reduction_step,
        metric_names: [selection.metric_name],
        strategies: ['median', 'lttb', 'p5', 'p25', 'p75', 'p95'],
    })
    // map the items by the metric_target, metric_name for easier access
    var result = {}
    res.data.data.forEach(sts => { result[sts.metric_target] = {
      [sts.metric_name]: sts,
    }})
    return result
  }
  catch (err) {
    console.log('Failed to fetch services system statistic snapshots.')
    console.log(err?.response?.data ?? err)
    return null
  }
}

async function fetchSummary() {
  try {
    var endpoint = (selection.metric_target == METRIC_TARGETS.redis.key)
                    ? `http://localhost:8003/metric/workload/${traceId.value}/redis-metrics`
                    : `http://localhost:8003/metric/workload/${traceId.value}/rabbitmq-metrics`
    var res = await axios.post(endpoint, {
        reduction_step: props.workload.duration_ms,
        metric_names: [selection.metric_name],
        strategies: ['median', 'min', 'max', 'p5', 'p25', 'p75', 'p95'],
    })
    // map the items by the metric_target, metric_name for easier access
    var result = {}
    res.data.data.forEach(sts => { result[sts.metric_target] = {
      [sts.metric_name]: sts,
    }})
    return result
  }
  catch (err) {
    console.log('Failed to fetch services system statistic snapshots.')
    console.log(err?.response?.data ?? err)
    return null
  }
}
</script>

<template>
  <div id="infra-metric-section">
    <div class="card shadow-md p-4">
      <div class="mb-4 d-flex align-items-center">
        <div class="form-group d-flex align-items-center">
          <IconStack class="d-inline-block me-2" width="24" />
          <h5 class="m-0 me-3" style="white-space: nowrap;">Infrastructure Metric</h5>
          <select v-model="selection.metric_target" class="form-select form-select-sm d-inline-block" style="width: 220px;">
            <option v-for="target in METRIC_TARGETS" :key="target.key" :value="target.key">
              {{ target.label }}
            </option>
          </select>
        </div>
      </div>

      <p v-show="!isChartReady" class="text-center text-secondary">Failed to fetch infra service statistic.</p>

      <div class="chart-container d-flex column-gap-3" :class="isChartReady ? 'd-flex' : 'd-none'">
        <div class="col p-0">

          <div class="form-group d-inline-flex align-items-center mb-3" style="width: 150px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Reduction:</label>
            <select v-model="selection.reduction_step" class="form-select form-select-sm">
              <option v-for="(rd, idx) in REDUCTION_STEPS" :key="idx" :value="rd">
                {{ rd }}
              </option>
            </select>
          </div>

          <div class="form-group d-inline-flex align-items-center ms-4 mb-3" style="width: 270px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Metric type:</label>
            <select v-model="selection.metric_type" class="form-select form-select-sm">
              <option v-for="type in METRIC_TYPES" :key="type.key" :value="type.key">
                {{ type.label }}
              </option>
            </select>
          </div>

          <div class="form-group d-inline-flex align-items-center ms-4 mb-3" style="width: 250px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Data:</label>
            <select v-model="selection.metric_name" class="form-select form-select-sm">
              <option v-for="data in METRIC_NAMES[selection.metric_target]" :key="data.key" :value="data.key">
                {{ data.label }}
              </option>
            </select>
          </div>

          <canvas id="infra-stats-chart-canvas"></canvas>
        </div>
        <div id="infra-metric-aggregations" class="col-auto p-0" >
          <div class="aggregations rounded-3 bg-light p-3 h-100" style="font-size: 15px;">
            <table class="table table-borderless m-0" style="background: transparent;min-width: 250px;">
              <tbody>
                <tr>
                  <th colspan="2" style="font-size: 14px;">Summarization</th>
                </tr>
                <tr>
                  <td class="ps-4">Median</td>
                  <td>{{ getSummary('median') }}</td>
                </tr>
                <tr>
                  <td class="ps-4">Minimum</td>
                  <td>{{ getSummary('min') }}</td>
                </tr>
                <tr>
                  <td class="ps-4">Maximum</td>
                  <td>{{ getSummary('max') }}</td>
                </tr>
                <tr>
                  <td class="ps-4">Percentiles</td>
                </tr>
                <tr>
                  <td colspan="2" class="ps-4">
                    <div class="ps-4x">
                      <table class="table w-100">
                        <tbody>
                          <tr>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">p5</td>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">{{ getSummary('p5') }}</td>
                          </tr>
                          <tr>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">p25</td>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">{{ getSummary('p25') }}</td>
                          </tr>
                          <tr>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">p75</td>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">{{ getSummary('p75') }}</td>
                          </tr>
                          <tr>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">p95</td>
                            <td class="px-2 py-1 border-1" style="font-size: 14px;">{{ getSummary('p95') }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
