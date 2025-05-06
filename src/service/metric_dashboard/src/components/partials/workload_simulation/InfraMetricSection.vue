<script setup>
import { ref, onMounted } from 'vue'
import {
  Chart,
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  TimeScale,
  Tooltip,
  Filler,
  CategoryScale
} from 'chart.js'
import 'chartjs-adapter-date-fns' // for time scale (X axis)
import IconStack from '@/components/icons/IconStack.vue'
import IconLock from '@/components/icons/IconLock.vue'
import IconCommand from '@/components/icons/IconCommand.vue'

Chart.register(LineController, LineElement, PointElement, LinearScale, TimeScale, Tooltip, Filler, CategoryScale)

const canvasRef = ref(null)

const dataPoints = [
  { x: 0, y: 20 },
  { x: 500, y: 35 },
  { x: 1000, y: 55 },
  { x: 1500, y: 40 },
  { x: 2000, y: 70 },
  { x: 2500, y: 60 }
]

onMounted(() => {
  new Chart(canvasRef.value.getContext('2d'), {
    type: 'line',
    data: {
      datasets: [
        {
          label: 'CPU Utilization',
          data: dataPoints,
          borderColor: 'rgb(17, 19, 68)',
          backgroundColor: 'rgba(17, 19, 68, 0.2)',
          tension: 0.4,
          fill: true,
          pointRadius: 3
        }
      ]
    },
    options: {
      responsive: true,
      plugins: {
        legend: { display: false },
        datalabels: {
          display: () => false
        },
        tooltip: {
          callbacks: {
            label: function (context) {
              const x = context.parsed.x;
              const y = context.parsed.y;
              return `Time: ${x} ms, CPU: ${y}%`;
            }
          }
        }
      },
      scales: {
        x: {
          type: 'linear',
          position: 'bottom',
          title: { display: false }, // Hides "Time (ms)"
          ticks: { beginAtZero: true }
        },
        y: {
          title: { display: false }, // Hides "CPU Utilization (%)"
          beginAtZero: true,
          max: 100
        }
      }
    }
  })
})
</script>

<template>
  <div id="infra-metric-section">
    <div class="card shadow-md p-4">
      <div class="mb-4 d-flex align-items-center">
        <div class="form-group d-flex align-items-center">
          <IconStack class="d-inline-block me-2" width="24" />
          <h5 class="m-0 me-3" style="white-space: nowrap;">Infrastructure Metric</h5>
          <select name="" id="" class="form-select form-select-sm d-inline-block" style="width: 220px;">
            <option value="">RabbitMQ</option>
            <option value="">Redis</option>
            <option value="">MongoDB</option>
          </select>
        </div>
      </div>

      <div class="infra-selection input-group mb-4">
        <button class="col btn btn-lg btn-dark rounded-0 fs-6">
          <span class="d-flex align-items-center justify-content-center">
            <IconLock class="me-2" width="20" /> Locks Stats
          </span>
        </button>
        <button class="col btn btn-lg btn-outline-dark rounded-0 fs-6">
          <span class="d-flex align-items-center justify-content-center">
            <IconCommand class="me-2" width="20" /> Command Stats
          </span>
        </button>
      </div>

      <div class="chart-container d-flex column-gap-3">
        <div class="col p-0">
          <div class="form-group d-flex align-items-center mb-3" style="width: 260px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Data type:</label>
            <select name="" id="" class="form-select form-select-sm">
              <option value="">Lock wait time (ms)</option>
              <option value="">Lock time to live (ms)</option>
            </select>
          </div>
          <canvas ref="canvasRef"></canvas>
        </div>
        <div id="infra-metric-aggregations" class="col-auto p-0">
          <div class="rounded-3 bg-light p-3 h-100" style="font-size: 15px;">
            <table class="table table-borderless m-0" style="background: transparent;">
              <tbody>
                <tr>
                  <th colspan="2" style="font-size: 14px;">Reservation (3 instances)</th>
                </tr>
                <tr>
                  <td class="ps-4" style="width: 130px;">CPU</td>
                  <td>1 vCPU/instance</td>
                </tr>
                <tr>
                  <td class="ps-4">Memory</td>
                  <td>1 GB/instance</td>
                </tr>
                <tr>
                  <th colspan="2" style="font-size: 14px;">Summarization</th>
                </tr>
                <tr>
                  <td class="ps-4">Duration</td>
                  <td>500 ms</td>
                </tr>
                <tr>
                  <td class="ps-4">Average</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-4">Median</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-4">Percentile</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-4">Minimum</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-4">Maximum</td>
                  <td>35%</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
