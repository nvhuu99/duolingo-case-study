<script setup>
import '../assets/work_simulation.min.css'
import { computed, onBeforeMount, ref } from 'vue'
import OperationsExecTimeSection from '../components/partials/workload_simulation/OperationsExecTimeSection.vue'
import WorkloadGenerateFormSection from '../components/partials/workload_simulation/WorkloadGenerateFormSection.vue'
import ServicesMetricSection from '../components/partials/workload_simulation/ServicesMetricSection.vue'
import InfraMetricSection from '../components/partials/workload_simulation/InfraMetricSection.vue'
import axios from 'axios'

const traceId = ref('bfcd13ec-f6e1-420c-868d-44733545ad78')

const workload = ref(null)

onBeforeMount(async () => {
  var response = await fetchWorkloadMetadata()
  workload.value = response
})

async function fetchWorkloadMetadata() {
  try {
    var res = await axios.get(`http://localhost:8003/metric/workload/${traceId.value}/workload-metadata`)
    return res.data.data
  }
  catch (err) {
    console.log('Failed to fetch workload metadata.')
    console.log(err?.response?.data ?? err)
    return null
  }
}

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
    <OperationsExecTimeSection :trace-id="traceId"/>
  </section>

  <section class="container pb-4">
    <ServicesMetricSection v-if="workload" :trace-id="traceId" :workload="workload"/>
  </section>

  <section class="container">
    <InfraMetricSection v-if="workload" :trace-id="traceId" :workload="workload"/>
  </section>
</template>
