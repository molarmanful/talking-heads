<script>
  import { onMount } from 'svelte'
  import { writable } from 'svelte/store'

  import { In, Out } from '$lib'

  let out = writable([])
  let span = (b, c) => `<span style="color:${c}">${b}</span>`
  let cout = (s, c = 'inherit') => out.update(o => o.concat(span(s, c)))
  let value = ''
  let ws
  let pre = true
  let inp
  let outp

  let parse = data => {
    let [h, id, c, ...b] = data.split` `

    let f = {
      ['+']() {
        cout(`${span(id, c)} joined`, 'gray')
      },

      m() {
        cout(`${span(id, c)}: ${b.join` `}`)
      },

      e() {
        out += cout(`(${b})`, 'red')
      },

      ['-']() {
        out += cout(`${span(id, c)} left`, 'gray')
      },
    }

    if (f[h]) f[h]()
  }

  onMount(() => {
    ws = new WebSocket(
      `${location.protocol == 'https' ? 'wss' : 'ws'}://${location.host}/ws`
    )

    ws.addEventListener('open', () => {
      console.log('open')
      pre = false
    })

    ws.addEventListener('message', ({ data }) => {
      console.log(data)
      parse(data)
    })

    ws.addEventListener('close', () => {
      console.log('close')
      cout('(disconnected)')
    })
  })
</script>

<main class="flex-(~ col) screen p-8 b-(1 black)">
  <Out el={outp} {out} {value} />
  <!-- TODO: char indicator -->
  <In el={inp} {pre} {value} {ws} />
</main>
