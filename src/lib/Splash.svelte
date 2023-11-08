<script>
  import { fade } from 'svelte/transition'

  export let splash = true
</script>

{#if splash}
  <div
    class="fixed screen inset-0 flex-(~ justify-center items-center) bg-black cursor-pointer z-36"
    role="presentation"
    on:click={() => {
      splash = false
    }}
    transition:fade={{ duration: 200 }}
  >
    <div class="text-center">
      <h1 class="text-white">talking-heads</h1>
      <p class="mt-2">by ben pang</p>
      <p class="text-normal text-white mt-6">[press to start]</p>
    </div>
  </div>
{/if}
