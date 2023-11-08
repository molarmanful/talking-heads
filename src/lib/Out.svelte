<script>
  import { createResizeObserver } from '@grail-ui/svelte'
  import { fly } from 'svelte/transition'

  import { goodBG } from '.'

  let el
  export let out
  export let typ

  const { useResizeObserver, entries } = createResizeObserver()

  let scrollB = () => {
    if (el)
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight
      })
  }
  $: $entries[0], out, typ, scrollB()
</script>

<div
  bind:this={el}
  class="flex-1 p-(8 r-[31%]) bord overflow-auto"
  ovf-parent
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
    <p class="text-sm" transition:fly={{ duration: 300, x: -40 }}>
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

  [ovf-parent] * {
    overflow-anchor: none;
  }
</style>
