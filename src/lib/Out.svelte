<script>
  import { goodBG } from '.'

  export let el
  export let out
  export let typ
  export let value

  let scrollB = () => {
    if (el)
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight
      })
  }
  $: value, scrollB()
</script>

<div bind:this={el} class="flex-1 p-(8 r-[31%]) bord overflow-auto">
  {#each $out as { type, idc, m }}
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
  {/each}

  {#each $typ as { id, c }}
    <p class="text-gray">
      &gt; <span style:color={c} style:background-color={goodBG(c)}>{id}</span> is
      thinking...
    </p>
  {/each}
</div>
