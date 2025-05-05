<script setup>
import '../assets/work_simulation.min.css'
import OperationExecTimeGanttChart from '../components/partials/workload_simulation/OperationExecTimeGanttChart.vue'

const workloadId = ref(null)

</script>

<template>
  <section id="operations-execution-time-gantt-chart">
    <OperationExecTimeGanttChart />
  </section>
</template>
