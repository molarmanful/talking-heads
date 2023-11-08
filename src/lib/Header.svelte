<script>
  import { fly } from 'svelte/transition'

  export let splash = true
</script>

{#if !splash}
  <h3
    class="text-center font-bold underline"
    transition:fly={{ duration: 300, y: -40 }}
  >
    <a
      class="hover:text-white transition-colors"
      href="https://github.com/molarmanful/talking-heads"
      target="_blank">talking-heads</a
    >
  </h3>
{/if}
