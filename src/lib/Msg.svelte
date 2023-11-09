<script>
  let clazz
  export { clazz as class }
  export let col = false
</script>

<div class="flex {col ? 'lt-lg:flex-col' : ''} mt-3 {clazz}">
  <div class="lg:w-36 lg:text-right"><slot name="h" /></div>
  <div class="flex-1 pl-5"><slot name="b" /></div>
</div>
