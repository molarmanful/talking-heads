<script>
  import { onMount } from 'svelte'
  import { writable } from 'svelte/store'

  import { Header, In, Out, Splash } from '$lib'

  let out = writable([])
  let typ = writable([])
  let splash = true
  let value = ''
  let idc = {}
  let ws
  let fav = 'NONE'

  let D = {}
  D.msg = (idc, m) => out.update(o => o.concat({ type: 'msg', idc, m }))
  D.err = m => out.update(o => o.concat({ type: 'err', m }))
  D.info = m => out.update(o => o.concat({ type: 'info', m }))
  D.succ = m => out.update(o => o.concat({ type: 'succ', m }))
  D.tpush = idc => typ.update(o => o.concat(idc))
  D.tpop = id => typ.update(o => o.filter(a => a.id != id))

  let parse = data => {
    let [h, id, c, ...b] = data.split` `

    let f = {
      ['+']() {
        idc = { id, c }
      },

      ['w']() {
        ;[fav, ...b] = b.join` `.split`\n`.filter(x => x.trim())
        for (let x of b) {
          let [id, c, ...s] = x.split` `
          D.msg({ id, c }, s.join` `)
        }
      },

      ['+t']() {
        D.tpush({ id, c })
      },

      ['-t']() {
        D.tpop(id)
      },

      m() {
        D.msg({ id, c }, b.join` `)
      },

      r() {
        D[+b[0] > 0 ? 'succ' : 'err'](
          `You ${+b[0] > 0 ? 'gained' : 'lost'} favor with ${id}.`
        )
      },

      e() {
        D.err(b.join` `)
      },

      ['-']() {},
    }

    if (f[h]) f[h]()
  }

  onMount(() => {
    ws = new WebSocket(
      `${location.protocol == 'https:' ? 'wss' : 'ws'}://${location.host}/ws`
    )

    ws.addEventListener('open', () => {})

    ws.addEventListener('message', ({ data }) => {
      parse(data)
    })

    ws.addEventListener('close', () => {
      D.err('disconnected')
    })
  })
</script>

<main class="flex-(~ col) screen max-screen p-(8 t-3) gap-5 overflow-hidden">
  <Splash bind:splash />
  <Header {splash} />
  <Out {out} {typ} />
  <In {D} {fav} {idc} {splash} {value} {ws} />
</main>
