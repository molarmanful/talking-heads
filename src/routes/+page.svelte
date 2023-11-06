<script>
  import { onMount } from 'svelte'
  import { writable } from 'svelte/store'
  import tinycolor from 'tinycolor2'

  import { In, Out } from '$lib'

  let out = writable([])
  let value = ''
  let ws
  let pre = true
  let inp
  let outp

  let span = (b, c, inv = false) => {
    let tc = tinycolor(c)
    return `<span style="color:${c};${
      inv
        ? 'background-color:' +
          tinycolor.mostReadable(tc, [
            '#000',
            tc
              .complement()
              .lighten((1 - tc.getLuminance()) * 100)
              .desaturate(69),
          ])
        : ''
    }">${b}</span>`
  }
  let cout = (s, c = 'inherit') => out.update(o => o.concat(span(s, c)))

  let parse = data => {
    let [h, id, c, ...b] = data.split` `

    let f = {
      ['+']() {
        cout(`> ${span(id, c, true)} joined`, 'gray')
      },

      // TODO
      t() {},

      m() {
        cout(`${span(id, c, true)}: ${b.join` `}`)
      },

      e() {
        out += cout(`> ${b}`, 'red')
      },

      ['-']() {
        out += cout(`> ${span(id, c, true)} left`, 'gray')
      },
    }

    if (f[h]) f[h]()
  }

  onMount(() => {
    ws = new WebSocket(
      `${location.protocol == 'https' ? 'wss' : 'ws'}://${location.host}/ws`
    )

    ws.addEventListener('open', () => {
      pre = false
    })

    ws.addEventListener('message', ({ data }) => {
      parse(data)
    })

    ws.addEventListener('close', () => {
      cout('(disconnected)', 'red')
    })
  })
</script>

<main class="flex-(~ col) screen max-screen p-8">
  <Out el={outp} {out} {value} />
  <!-- TODO: char indicator -->
  <In el={inp} {pre} {value} {ws} />
</main>
