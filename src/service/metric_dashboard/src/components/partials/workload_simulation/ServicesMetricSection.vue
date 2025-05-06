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
import IconInstace from '@/components/icons/IconInstace.vue'
import IconCPU from '@/components/icons/IconCPU.vue'
import IconRAM from '@/components/icons/IconRAM.vue'
import IconDisk from '@/components/icons/IconDisk.vue'

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
  <div id="services-metric-section">
    <div class="card shadow-md p-4">
      <div class="mb-4 d-flex align-items-center">
        <div class="form-group d-flex align-items-center">
          <IconInstace class="d-inline-block me-2" width="24" />
          <h5 class="m-0 me-3" style="white-space: nowrap;">Services Metric</h5>
          <select name="" id="" class="form-select form-select-sm d-inline-block" style="width: 220px;">
            <option value="">Input Messages API</option>
            <option value="">Notification Builders</option>
            <option value="">Push Notification Senders</option>
          </select>
        </div>
      </div>

      <div class="services-selection input-group mb-4">
        <button class="col-4 btn btn-lg btn-dark rounded-0 fs-6">
          <span class="d-flex align-items-center justify-content-center">
            <IconCPU class="me-2" width="20" /> CPU stats
          </span>
        </button>
        <button class="col-4 btn btn-lg btn-outline-dark rounded-0 fs-6">
          <span class="d-flex align-items-center justify-content-center">
            <IconRAM class="me-2" width="20" /> Memory stats
          </span>
        </button>
        <button class="col-4 btn btn-lg btn-outline-dark rounded-0 fs-6">
          <span class="d-flex align-items-center justify-content-center">
            <IconDisk class="me-2" width="20" /> Disk IO stats
          </span>
        </button>
      </div>

      <div class="chart-container d-flex column-gap-3">
        <div class="col p-0">
          <div class="form-group d-flex align-items-center mb-3" style="width: 240px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Data:</label>
            <select name="" id="" class="form-select form-select-sm">
              <option value="">CPU Utilization (%)</option>
              <option value="">IO Wait (ms)</option>
            </select>
          </div>
          <canvas ref="canvasRef"></canvas>
        </div>
        <div id="service-metric-aggregations" class="col-auto p-0">
          <div class="aggregations rounded-3 bg-light p-3 h-100" style="font-size: 15px;">
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
