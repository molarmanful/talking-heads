<script>
  let clazz
  export { clazz as class }
</script>

<div class="text-gray-300 flex {clazz}">
  <div class="w-50 text-right"><slot name="h" /></div>
  <div class="flex-1"><slot name="b" /></div>
</div>
