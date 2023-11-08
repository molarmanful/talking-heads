<script>
  let clazz
  export { clazz as class }
</script>

<div class="flex mt-3 {clazz}">
  <div class="w-50 text-right pr-5"><slot name="h" /></div>
  <div class="flex-1"><slot name="b" /></div>
</div>
