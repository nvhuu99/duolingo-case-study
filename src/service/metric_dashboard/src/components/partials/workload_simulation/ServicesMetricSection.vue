<script setup>
import { ref, onMounted, toRef, watch, computed, reactive } from 'vue'
import {
  Chart,
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  Title,
  CategoryScale,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import 'chartjs-adapter-date-fns'
import axios from 'axios'
import IconInstance from '@/components/icons/IconInstance.vue'
import IconCPU from '@/components/icons/IconCPU.vue'
import IconRAM from '@/components/icons/IconRAM.vue'

const SERVICE_NAMES = {
  noti_builder: { key: 'noti_builder', label: 'Notification Builder'},
  input_message_api: { key: 'input_message_api', label: 'Input Message API'},
  push_noti_sender: { key: 'push_noti_sender', label: 'Push Notification Sender'},
}
const METRIC_TYPES = {
  all: {key: 'all', label: 'Show all'},
  median: {key: 'median', label: 'Median'},
  lttb: {key: 'lttb', label: 'Largest Triangle (LTTB)'},
  percentiles: {key: 'percentiles', label: 'Percentiles'},
}
const METRIC_NAMES = {
  cpu_util: {key: 'cpu_util', label: 'CPU Utilization (%)'},
  memory_used_mb: {key: 'memory_used_mb', label: 'Memory used (MB)'},
}

const REDUCTION_STEPS = [100, 200, 500, 1000, 1500, 2000, 2500, 5000]

const props = defineProps([
  'traceId',
  'workload',
])
const traceId = toRef(props, 'traceId')
const chartInstance = ref(null)
const selection = reactive({
  service_name: SERVICE_NAMES.noti_builder.key,
  reduction_step: 1000,
  metric_name: METRIC_NAMES.cpu_util.key,
  metric_type: METRIC_TYPES.all.key,
  instance_id: '',
})
const metricSummary = ref(null)

const isChartReady = computed(() => chartInstance.value != null)
const serviceInstances = computed(() => {
  var idx = props.workload.service_instances.findIndex(sv => sv.service_name == selection.service_name)
  return (idx != -1) ? props.workload.service_instances[idx].instance_ids : []
})
const metricUnit = computed(() => {
  switch (selection.metric_name) {
    case METRIC_NAMES.memory_used_mb.key:
      return ' mb';
    case METRIC_NAMES.cpu_util.key:
    default:
      return '%';
  }
})

onMounted(() => {
  Chart.register(
    LineController,
    LineElement,
    PointElement,
    LinearScale,
    CategoryScale,
    Title,
    Tooltip,
    Legend,
    Filler,
  );
  renderChart()
  updateSummary()
})

watch(() => selection.service_name, function() {
  selection.instance_id = ""
})
watch([traceId, selection], async () => {
  await destroyChart()
  await renderChart()
  await updateSummary()
});


function toggleWarning() {
  document.getElementById('data-warning')?.classList?.toggle('d-none')
}

function shouldShowMetricType(type) {
  return selection.metric_type == METRIC_TYPES.all.key || selection.metric_type == type.key
}

function getSummary(type) {
  var val = metricSummary.value?.[selection.service_name]?.[selection.metric_name]?.reduced_snapshots?.[type]?.[0].value
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
  if (!metrics?.[selection.service_name]?.[selection.metric_name]) return

  const snapshots = metrics[selection.service_name][selection.metric_name].reduced_snapshots
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
    // {
    //   type: 'line',
    //   label: '1th percentile',
    //   data: makeChartDataSet('p1'),
    //   backgroundColor: 'rgba(0, 0, 0, 0.1)',
    //   borderWidth: 0,
    //   pointRadius: 0,
    //   fill: '+1',
    //   hidden: !shouldShowMetricType(METRIC_TYPES.percentiles),
    // },
    // {
    //   type: 'line',
    //   label: '99th percentile',
    //   data: makeChartDataSet('p99'),
    //   backgroundColor: 'rgba(0, 0, 0, 0)',
    //   borderWidth: 0,
    //   pointRadius: 0,
    //   fill: false,
    //   hidden: !shouldShowMetricType(METRIC_TYPES.percentiles),
    // },
  ]

  const canvas = document.getElementById('service-stats-chart-canvas')
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
        legend: { display: false },
        tooltip: {
          callbacks: {
            label(context) {
              const asMs = Intl.NumberFormat('en-US').format(context.parsed.x);
              const asSeconds = (context.parsed.x / 1000).toFixed(1);
              const formatedVal = Intl.NumberFormat('en-US').format(Math.round(context.parsed.y).toFixed(0))
              switch (selection.metric_name) {
                case METRIC_NAMES.memory_used_mb.key:
                  return `After ${asMs}ms - ${asSeconds}seconds, Memory used ${formatedVal} mb`;
                case METRIC_NAMES.cpu_util.key:
                default:
                  return `After ${asMs}ms - ${asSeconds}seconds, CPUs utilized ${formatedVal}%`;
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
    var res = await axios.post(`http://localhost:8003/metric/workload/${traceId.value}/service-metrics`, {
        service_name: selection.service_name,
        service_operation: '',
        instance_ids: [selection.instance_id],
        reduction_step: selection.reduction_step,
        metric_names: [selection.metric_name],
        strategies: ['median', 'lttb', 'p5', 'p25', 'p75', 'p95']
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
    var res = await axios.post(`http://localhost:8003/metric/workload/${traceId.value}/service-metrics`, {
        service_name: selection.service_name,
        service_operation: '',
        instance_ids: [selection.instance_id],
        reduction_step: props.workload.duration_ms,
        metric_names: [selection.metric_name],
        strategies: ['median', 'min', 'max', 'p5', 'p25', 'p75', 'p95']
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
  <div id="services-metric-section">
    <div class="card shadow-md p-4">
      <div class="mb-4 d-flex align-items-center">
        <div class="form-group d-flex align-items-center">
          <IconInstance class="d-inline-block me-2" width="24" />
          <h5 class="m-0 me-3" style="white-space: nowrap;">Services Metric</h5>
          <select v-model="selection.service_name" class="form-select form-select-sm d-inline-block" style="width: 220px;">
            <option v-for="svName in SERVICE_NAMES" :key="svName.key" :value="svName.key">
              {{ svName.label }}
            </option>
          </select>
        </div>
      </div>

      <p v-show="!isChartReady" class="text-center text-secondary">Failed to fetch services system statistic.</p>

      <div v-show="isChartReady" class="services-selection input-group mb-4">
        <button class="col btn btn-lg rounded-0 fs-6" :class="selection.metric_name == METRIC_NAMES.cpu_util.key ? 'btn-dark' : 'btn-outline-dark'" @click="() => selection.metric_name = METRIC_NAMES.cpu_util.key">
          <span class="d-flex align-items-center justify-content-center">
            <IconCPU class="me-2" width="20" /> CPU stats
          </span>
        </button>
        <button class="col btn btn-lg rounded-0 fs-6" :class="selection.metric_name == METRIC_NAMES.memory_used_mb.key ? 'btn-dark' : 'btn-outline-dark'" @click="() => selection.metric_name = METRIC_NAMES.memory_used_mb.key">
          <span class="d-flex align-items-center justify-content-center">
            <IconRAM class="me-2" width="20" /> Memory stats
          </span>
        </button>
        <!-- <button class="col-4 btn btn-lg btn-outline-dark rounded-0 fs-6">
          <span class="d-flex align-items-center justify-content-center">
            <IconDisk class="me-2" width="20" /> Disk IO stats
          </span>
        </button> -->
      </div>

      <div v-show="isChartReady" id="data-warning" class="alert alert-warning mb-4" style="font-size: 14px;">
        <span class="me-1">The data is aggregated for all service operations of all instances</span>
        <span class="text-primary" role="button" @click="toggleWarning">(close warning)</span>
      </div>

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

          <div class="form-group d-inline-flex align-items-center ms-4 mb-3" style="width: 260px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Instance:</label>
            <select v-model="selection.instance_id" class="form-select form-select-sm">
              <option value="">All instances</option>
              <option v-for="(instanceId, idx) in serviceInstances" :key="idx" :value="instanceId">
                {{ instanceId }}
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
          <canvas id="service-stats-chart-canvas"></canvas>
        </div>
        <div id="service-metric-aggregations" class="col-auto p-0" >
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
