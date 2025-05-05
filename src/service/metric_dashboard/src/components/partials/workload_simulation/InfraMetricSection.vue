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
          borderColor: '#4caf50',
          backgroundColor: 'rgba(76, 175, 80, 0.2)',
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
    <div class="card shadow-sm p-4">
      <div class="mb-4 d-flex align-items-center">
        <h5 class="m-0 me-4" style="font-size: 16px"> Infrastructure Metric </h5>
        <div class="form-group d-flex align-items-center" style="width: 260px;">
          <label class="form-label m-0 me-2"><span style="font-size: 14px;">Service:</span></label>
          <select name="" id="" class="form-select form-select-sm">
            <option value="">RabbitMQ</option>
            <option value="">Redis</option>
          </select>
        </div>
      </div>

      <div class="services-selection input-group mb-4">
        <button class="col btn btn-lg btn-dark rounded-0 fs-6">Locks Stats</button>
        <button class="col btn btn-lg btn-outline-dark rounded-0 fs-6">Commands Stats</button>
      </div>

      <div class="chart-container d-flex column-gap-3">
        <div class="col-8 p-0">
          <div class="form-group d-flex align-items-center mb-3" style="width: 260px;">
            <label class="form-label m-0 me-2" style="font-size: 14px;white-space: nowrap;">Data type:</label>
            <select name="" id="" class="form-select form-select-sm">
              <option value="">Lock wait time (ms)</option>
              <option value="">Lock time to live (ms)</option>
            </select>
          </div>
          <canvas ref="canvasRef"></canvas>
        </div>
        <div id="service-metric-aggregations" class="col p-0">
          <div class="rounded-3 bg-light p-3 h-100">
            <table class="table table-borderless m-0" style="background: transparent;">
              <tbody>
                <tr>
                  <th>Duration</th>
                  <td>500ms</td>
                </tr>
                <tr>
                  <th>Snapshots</th>
                  <td>3</td>
                </tr>
                <tr>
                  <th>Summarization</th>
                </tr>
                <tr>
                  <td class="ps-5">Average</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-5">Median</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-5">Percentile</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-5">Minimum</td>
                  <td>35%</td>
                </tr>
                <tr>
                  <td class="ps-5">Maximum</td>
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
