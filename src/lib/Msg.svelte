<script>
  let clazz
  export { clazz as class }
</script>

<div class="flex {clazz}">
  <div class="w-36 text-right"><slot name="h" /></div>
  <div class="flex-1 pl-5"><slot name="b" /></div>
</div>
