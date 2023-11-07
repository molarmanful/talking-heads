<script>
  import { goodBG } from '$lib'

  export let D
  export let ws
  export let value
  export let el
  export let idc
  export let splash = true
  export let fav
  $: if (!splash && el)
    requestAnimationFrame(() => {
      el.focus()
    })
</script>

<form
  class="flex-(~ row) gap-3"
  on:submit|preventDefault={() => {
    let v = value.trim()

    let m = /^\/\w+/.exec(v)
    if (m) {
      m = m[0].slice(1)

      let f = {
        help() {
          D.info(
            'Available commands: ' + Object.keys(f).map(x => '/' + x).join`, `
          )
        },
        whoami() {
          D.info(`You are ${idc.id} of ${fav}.`)
        },
      }
      f['?'] = f.h = f.help
      f['i'] = f.whoami
      if (f[m]) f[m]()
      else D.err('unknown command')

      value = ''
      return
    }

    if (v && v.length <= 250) {
      ws.send('m ' + value.slice(0, 250))
      value = ''
    }
  }}
>
  <label
    style:color={idc.c}
    style:background-color={goodBG(idc.c)}
    class="h-full p-2 bord"
    for="in"
  >
    {idc.id}
  </label>
  <div class="relative flex-1">
    <input
      bind:this={el}
      id="in"
      class="full p-2 bg-transparent bord"
      disabled={splash}
      maxlength="250"
      placeholder="{splash ? 'loading' : 'chat'}..."
      type="text"
      bind:value
    />
    <div
      class="absolute flex inset-(r-1 t-1) flex-items-center flex-justify-items-center text-sm"
    >
      <span style:color={value.length > 240 ? 'red' : 'inherit'}>
        {value.length}/250
      </span>
    </div>
  </div>
  <button class="p-2 bg-transparent bord" disabled={splash}>send</button>
</form>
