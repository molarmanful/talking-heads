import { TinyColor, mostReadable } from '@ctrl/tinycolor'

export { default as Splash } from './Splash.svelte'
export { default as Header } from './Header.svelte'
export { default as Msg } from './Msg.svelte'
export { default as Users } from './Users.svelte'
export { default as In } from './In.svelte'
export { default as Out } from './Out.svelte'

export const goodBG = h => {
  let tc = new TinyColor(h)
  return mostReadable(tc, [
    '#000',
    (({ h, s }) => new TinyColor({ h, s: s * 0.2, l: 1 - tc.getLuminance() }))(
      tc.complement().toHsl()
    ),
  ])
}
