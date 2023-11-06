<script>
  export let ws
  export let value
  export let el
  export let pre = true
  $: if (!pre && el) el.focus()
</script>

<form
  class="flex-(~ row)"
  on:submit|preventDefault={() => {
    ws.send('m ' + value.slice(0, 250))
    value = ''
  }}
>
  <input
    bind:this={el}
    class="flex-1 p-2 b-(1 r-0 black)"
    disabled={pre}
    placeholder="{pre ? 'loading' : 'chat'}..."
    type="text"
    bind:value
  />
  <button class="p-2 bg-transparent b-(1 black)" disabled={pre}>send</button>
</form>
