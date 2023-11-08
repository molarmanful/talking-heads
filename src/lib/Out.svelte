<script>
  import { createResizeObserver } from '@grail-ui/svelte'
  import { fly } from 'svelte/transition'

  import { goodBG } from '.'

  let el
  export let out
  export let typ
  export let ws
  export let wslock
  let scrh = 0

  const { useResizeObserver, entries } = createResizeObserver()

  export const scrollB = () => {
    if (el)
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight
      })
  }
  export const keepScroll = () => {
    if (el)
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight - scrh
      })
  }
  $: $entries[0], out, typ, scrollB()
</script>

<div
  bind:this={el}
  class="flex-1 p-(8 r-[31%]) bord overflow-auto"
  ovr-parent
  on:scroll={() => {
    if (el.scrollTop <= 0 && !wslock) {
      scrh = el.scrollHeight
      ws.send('g')
      wslock = true
    }
  }}
  use:useResizeObserver
>
  {#each $out as { type, idc, m }}
    <div transition:fly={{ duration: 300, x: 40 }}>
      {#if type == 'msg'}
        <p class="text-gray-300">
          <strong style:color={idc.c} style:background-color={goodBG(idc.c)}
            >{idc.id}</strong
          >: {m}
        </p>
      {:else}
        <p
          class={[
            'text-sm',
            type == 'succ' ? 'text-green' : type == 'err' ? 'text-red' : '',
          ].join` `}
        >
          &gt; {m}
        </p>
      {/if}
    </div>
  {/each}

  {#each $typ as { id, c }}
    <p class="text-sm" transition:fly={{ duration: 300, y: 40 }}>
      &gt; <span style:color={c} style:background-color={goodBG(c)}>{id}</span
      >...
    </p>
  {/each}

  <div style:overflow-anchor="auto" class="h-[1px]" />
</div>

<style>
  p {
    --at-apply: 'mt-3';
  }

  [ovr-parent] * {
    overflow-anchor: none;
  }
</style>
