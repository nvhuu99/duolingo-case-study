<script setup>
import IconGenerate from '@/components/icons/IconGenerate.vue';
import IconHistory from '@/components/icons/IconHistory.vue';
import IconLightbulb from '@/components/icons/IconLightbulb.vue';
</script>

<template>
  <div id="workload-generate-form-section" class="d-flex column-gap-3">
    <div class="form col card shadow-md p-4">
      <h5 class="mb-4 fs-6 d-flex align-items-center">
        <IconGenerate class="me-2" fill="black" height="24" /> Workload Generate
      </h5>
      <div class="px-4">
        <div class="form-group mb-3">
          <label class="form-label" style="font-size: 15px;">Select workload</label>
          <select name="" id="" class="form-select">
            <option value="">10.000 devices</option>
            <option value="">50.000 devices</option>
            <option value="">100.000 devices</option>
          </select>
        </div>
        <div class="form-group mb-3">
          <label class="form-label" style="font-size: 15px;">Message title</label>
          <input type="text" name="" class="form-control">
        </div>
        <div class="form-group mb-3">
          <label class="form-label" style="font-size: 15px;">Message body</label>
          <textarea name="" class="form-control"></textarea>
        </div>
        <div class="text-center">
          <button class="btn btn-md btn-outline-dark d-inline-flex align-items-center">
            <IconLightbulb class="me-2" width="20"/>Generate
          </button>
        </div>
      </div>
    </div>
    <div class="col card shadow-md p-4">
      <h5 class="mb-4 fs-6 d-flex align-items-center">
        <IconHistory fill="black" class="me-2" height="24" /> History
      </h5>
      <div class="ps-3">
        <table class="table table-borderless m-0" style="background: transparent;">
          <tbody>
            <tr class="link link-active">
              <td>#3</td>
              <td>🕛 2025/05/05 16:26:00 - ⚙️ 50.000 devices<br>🖋️ 3 workers cluster size</td>
            </tr>
            <tr class="link">
              <td>#2</td>
              <td>🕛 2025/05/05 16:26:00 - ⚙️ 10.000 devices<br>🖋️ test RabbitMQ latency</td>
            </tr>
            <tr class="link">
              <td>#1</td>
              <td>🕛 2025/05/05 16:26:00 - ⚙️ 100.000 devices<br>🖋️ execution time benchmark</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
