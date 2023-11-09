<script>
  import { fade, fly } from 'svelte/transition'

  import { goodBG } from '.'

  export let users, splash
</script>

{#if !splash}
  <div
    class="w-44 min-h-0 ml-5 p-3 overflow-(x-hidden y-auto) bord text-sm lt-lg-hidden"
    transition:fade={{ duration: 300, x: 40 }}
  >
    <p class="mb-3">Online - {$users.size}</p>
    {#each [...$users].sort() as [id, c]}
      <div
        class="flex-(~ col) pl-2 mb-1"
        transition:fly={{ duration: 300, x: 40 }}
      >
        <div>
          <span style:color={c} style:background-color={goodBG(c)}>{id}</span>
        </div>
      </div>
    {/each}
  </div>
{/if}
