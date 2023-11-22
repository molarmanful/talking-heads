<script>
  import { fly } from 'svelte/transition'

  let el
  export let D, ws, value, idc, fav, scrollB
  export let splash = true
  export let censor = true

  $: if (!splash && el)
    requestAnimationFrame(() => {
      el.focus()
    })
</script>

<form
  class="flex-(~ row) gap-2 lg:gap-3"
  on:submit|preventDefault={() => {
    let v = value.trim()

    let m = /^\/\S+/.exec(v)
    if (m) {
      D.send(m)
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
      scrollB()
      return
    }

    if (v && v.length <= 250) {
      ws.send('m ' + value.slice(0, 250))
      value = ''
      scrollB()
    }
  }}
>
  {#if !splash}
    <button
      class="bord"
      type="button"
      on:click={() => {
        censor = !censor
      }}
      transition:fly={{ duration: 300, x: -40 }}
      >{censor ? '$%!@' : 'SHIT'}</button
    >

    <div class="relative flex-1" transition:fly={{ duration: 300, y: 40 }}>
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
        class="absolute flex right-1 top-1 flex-items-center flex-justify-items-center text-sm"
      >
        <span style:color={value.length > 240 ? 'red' : 'inherit'}>
          {value.length}/250
        </span>
      </div>
    </div>

    <button
      class="bord"
      disabled={splash || !value.length}
      transition:fly={{ duration: 300, x: 40 }}
    >
      send
    </button>
  {/if}
</form>
