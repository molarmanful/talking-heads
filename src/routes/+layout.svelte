<script>
  import '@unocss/reset/tailwind-compat.css'
  import 'uno.css'
  import '../app.postcss'

  import { onMount } from 'svelte'

  let loaded = false

  onMount(() => {
    loaded = true
  })
</script>

<svelte:head>
  <title>talking-heads</title>
</svelte:head>

<div class="{loaded ? 'opacity-100' : 'opacity-0'} transition-opacity-400">
  <slot />
</div>
