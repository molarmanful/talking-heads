<script>
  import { onMount } from 'svelte'

  let out = ''
  let value = ''
  let ws
  let pre = true

  let parse = data => {
    let [h, id, c, ...b] = data.split` `
    let f = {
      m() {
        out += `${id}: ${b.join` `}\n`
      },
    }
    if (f[h]) f[h]()
  }

  onMount(() => {
    ws = new WebSocket(
      `${location.protocol == 'https' ? 'wss' : 'ws'}://${location.host}/ws`
    )

    ws.addEventListener('open', e => {
      console.log('open')
      pre = false
    })

    ws.addEventListener('message', ({ data }) => {
      console.log(data)
      parse(data)
    })

    ws.addEventListener('close', e => {
      console.log('close')
    })
  })
</script>

<pre>{out}</pre>

<form
  on:submit|preventDefault={() => {
    ws.send(value)
    value = ''
  }}
>
  <input
    type="text"
    bind:value
    placeholder="{pre ? 'loading' : 'chat'}..."
    disabled={pre}
  />
  <button>send</button>
</form>
