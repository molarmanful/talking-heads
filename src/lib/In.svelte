<script>
  export let ws
  export let value
  export let el
  export let pre = true
  $: if (!pre && el)
    requestAnimationFrame(() => {
      el.focus()
    })
</script>

<form
  class="flex-(~ row)"
  on:submit|preventDefault={() => {
    if (value.trim()) {
      ws.send('m ' + value.slice(0, 250))
      value = ''
    }
  }}
>
  <input
    bind:this={el}
    class="flex-1 p-2 bg-transparent bord outline-none"
    disabled={pre}
    placeholder="{pre ? 'loading' : 'chat'}..."
    type="text"
    bind:value
  />
  <button class="p-2 bg-transparent bord" disabled={pre}>send</button>
</form>
