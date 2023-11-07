<script>
  import { fly } from 'svelte/transition'

  import { goodBG } from '.'

  let el
  export let out
  export let typ

  let scrollB = () => {
    if (el)
      requestAnimationFrame(() => {
        setTimeout(() => {
          el.scrollTop = el.scrollHeight
        })
      })
  }
  $: $out, $typ, scrollB()
</script>

<div bind:this={el} class="flex-1 p-(8 r-[31%]) bord overflow-auto">
  {#each $out as { type, idc, m }}
    <div transition:fly={{ duration: 200, x: -40 }}>
      {#if type == 'msg'}
        <p>
          <strong style:color={idc.c} style:background-color={goodBG(idc.c)}
            >{idc.id}</strong
          >: {m}
        </p>
      {:else if type == 'err'}
        <p class="text-red">&gt; {m}</p>
      {:else}
        <p class="text-gray">&gt; {m}</p>
      {/if}
    </div>
  {/each}

  {#each $typ as { id, c }}
    <p class="text-gray" transition:fly={{ duration: 200, y: 40 }}>
      &gt; <span style:color={c} style:background-color={goodBG(c)}>{id}</span> is
      thinking...
    </p>
  {/each}
</div>

<style>
  p {
    --at-apply: 'mt-3';
  }
</style>
