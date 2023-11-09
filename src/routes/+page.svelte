<script>
  import { onMount } from 'svelte'
  import { writable } from 'svelte/store'

  import { Header, In, Out, Splash, Users } from '$lib'

  let out = writable([])
  let typ = writable([])
  let users = writable(new Map())
  let splash = true
  let value = ''
  let idc = {}
  let ws
  let fav = '[REDACTED]'
  let wslock = false
  let scrollB
  let keepScroll

  let D = {}
  D.msg = (idc, m) => out.update(o => [...o, { type: 'msg', idc, m }])
  D.gsm = (idc, m) => out.update(o => [{ type: 'msg', idc, m }, ...o])
  D.err = m => out.update(o => [...o, { type: 'err', m }])
  D.warn = m => out.update(o => [...o, { type: 'warn', m }])
  D.info = m => out.update(o => [...o, { type: 'info', m }])
  D.succ = m => out.update(o => [...o, { type: 'succ', m }])
  D.tpush = idc => typ.update(o => [...o, idc])
  D.tpop = id => typ.update(o => o.filter(a => a.id != id))

  let parse = data => {
    let [h, id, c, ...b] = data.split` `

    let f = {
      ['+']() {
        idc = { id, c }
        users.update(o => o.set(id, c))
      },

      w() {
        ;[fav, ...b] = b.join` `.split`\n`.filter(x => x.trim())
        for (let x of b) {
          let [id, c, ...s] = x.split` `
          D.msg({ id, c }, s.join` `)
        }
        D.info(
          `Welcome, ${id} of ${fav}. Type /help for a list of available commands.`
        )
      },

      u() {
        for (let x of b.join` `.split`\n`) {
          let [id, c] = x.split` `
          users.update(o => o.set(id, c))
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
        D[+b[0] > 0 ? 'succ' : 'warn'](
          `You ${+b[0] > 0 ? 'gained' : 'lost'} favor with ${id}.`
        )
      },

      g() {
        b = b.join` `.split`\n`.reverse()
        for (let x of b) {
          let [id, c, ...s] = x.split` `
          D.gsm({ id, c }, s.join` `)
        }
        console.log(b)
        wslock = false
        keepScroll()
      },

      e() {
        D.err(b.join` `)
      },

      ['-']() {
        users.update(o => (o.delete(id), o))
      },
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
      D.err('Disconnected. Please reload the page to rejoin.')
    })
  })
</script>

<main
  class="flex-(~ col) screen max-screen p-(5 t-3) lg:p-(8 t-3) overflow-hidden"
>
  <Splash bind:splash />
  <Header {splash} />
  <div class="flex-(~ 1) py-3 lg:pb-5 overflow-hidden">
    <Out {out} {typ} {ws} bind:wslock bind:scrollB bind:keepScroll />
    <Users {splash} {users} />
  </div>
  <In {D} {fav} {idc} {scrollB} {splash} {value} {ws} />
</main>
