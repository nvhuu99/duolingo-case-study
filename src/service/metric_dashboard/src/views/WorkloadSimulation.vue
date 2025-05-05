<script setup>
import '../assets/work_simulation.min.css'
import { ref } from 'vue'
import OperationsExecTimeSection from '../components/partials/workload_simulation/OperationsExecTimeSection.vue'
import WorkloadGenerateFormSection from '../components/partials/workload_simulation/WorkloadGenerateFormSection.vue'
import ServicesMetricSection from '../components/partials/workload_simulation/ServicesMetricSection.vue'
import InfraMetricSection from '../components/partials/workload_simulation/InfraMetricSection.vue'

const workloadId = ref(null)

</script>

<template>
  <section class="container pb-4">
    <WorkloadGenerateFormSection />
  </section>

  <div id="workload-not-loaded-message" class="container pb-4 d-none">
    <p class="text-center m-0">
      No data. Please generate a new workload or select one from the History panel.
    </p>
  </div>

  <section class="container pb-4">
    <OperationsExecTimeSection />
  </section>

  <section class="container pb-4">
    <ServicesMetricSection />
  </section>

  <section class="container">
    <InfraMetricSection />
  </section>
</template>
