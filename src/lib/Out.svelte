<script>
  import { createResizeObserver } from '@grail-ui/svelte'
  import { fly } from 'svelte/transition'

  import { Msg, goodBG } from '.'

  let el
  export let out, typ, ws, wslock
  let scrh = 0

  const { useResizeObserver, entries } = createResizeObserver()

  export const scrollB = () => {
    if (el)
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight + 1
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
  class="flex-1 p-8 bord overflow-auto"
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
    <div transition:fly={{ duration: 300, y: 40 }}>
      {#if type == 'msg'}
        <Msg class="text-gray-300">
          <strong
            style:color={idc.c}
            style:background-color={goodBG(idc.c)}
            slot="h">{idc.id}</strong
          >
          <svelte:fragment slot="b">{m}</svelte:fragment>
        </Msg>
      {:else}
        <Msg
          class={[
            'text-sm',
            type == 'succ'
              ? 'text-green'
              : type == 'err'
              ? 'text-red'
              : type == 'warn'
              ? 'text-yellow'
              : '',
          ].join` `}
        >
          <svelte:fragment slot="h">&gt</svelte:fragment>
          <svelte:fragment slot="b">{m}</svelte:fragment>
        </Msg>
      {/if}
    </div>
  {/each}

  {#each $typ as { id, c }}
    <div transition:fly={{ duration: 300, y: 40 }}>
      <Msg class="text-sm">
        <svelte:fragment slot="h">&gt</svelte:fragment>
        <svelte:fragment slot="b"
          ><span style:color={c} style:background-color={goodBG(c)}>{id}</span
          >...</svelte:fragment
        >
      </Msg>
    </div>
  {/each}

  <div style:overflow-anchor="auto" class="h-[1px]" />
</div>

<style>
  [ovr-parent] * {
    overflow-anchor: none;
  }
</style>
