import { createRouter, createWebHistory } from 'vue-router'
import WorkloadSimulation from '../views/WorkloadSimulation.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: WorkloadSimulation,
    },
  ],
})

export default router
