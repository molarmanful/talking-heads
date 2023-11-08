<script>
  let clazz
  export { clazz as class }
</script>

<div class="flex mt-3 gap-8 {clazz}">
  <div class="w-32 text-right"><slot name="h" /></div>
  <div class="flex-1"><slot name="b" /></div>
</div>
