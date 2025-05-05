<template>
  <div id="workload-generate-form-section" class="d-flex column-gap-3">
    <div class="form col card shadow-sm p-4">
      <h5 class="mb-4" style="font-size: 16px"> Workload Generate </h5>
      <div class="px-4">
        <div class="form-group mb-3">
          <label class="form-label">Select workload</label>
          <select name="" id="" class="form-select">
            <option value="">10.000 devices</option>
            <option value="">50.000 devices</option>
            <option value="">100.000 devices</option>
          </select>
        </div>
        <div class="form-group mb-3">
          <label class="form-label">Message title</label>
          <input type="text" name="" class="form-control">
        </div>
        <div class="form-group mb-3">
          <label class="form-label">Select workload</label>
          <textarea name="" class="form-control" ></textarea>
        </div>
        <div class="text-center">
          <button class="btn btn-md btn-outline-dark"> Generate </button>
        </div>
      </div>
    </div>
    <div class="col card shadow-sm p-4">
      <h5 class="mb-4" style="font-size: 16px"> History </h5>
      <div class="ps-3">
        <table class="table table-borderless m-0" style="background: transparent;">
          <tbody>
            <tr>
              <td>#3</td>
              <td>2025/05/05 16:26:00 - Workload: 50.000 devices - Title: "workload_#1 - 3 workers cluster size"</td>
            </tr>
            <tr>
              <td>#2</td>
              <td>2025/05/05 16:26:00 - Workload: 50.000 devices - Title: "workload_#1 - 3 workers cluster size"</td>
            </tr>
            <tr>
              <td>#1</td>
              <td>2025/05/05 16:26:00 - Workload: 50.000 devices - Title: "workload_#1 - 3 workers cluster size"</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
