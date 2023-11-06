<script>
  export let el
  export let out
  export let value

  let scrollB = () => {
    if (el)
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight
      })
  }
  $: value, scrollB()
</script>

<div bind:this={el} class="flex-1 mb-4 b-(1 black)">
  {#each $out as m}
    <p>{@html m}</p>
  {/each}
</div>
