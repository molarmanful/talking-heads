import extractorSvelte from '@unocss/extractor-svelte'
import { colorResolver } from '@unocss/preset-mini/utils'
import {
  presetUno,
  presetWebFonts,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss'

export default {
  presets: [
    presetUno(),
    presetWebFonts({
      provider: 'google',
      fonts: {
        mono: 'Fira Code',
      },
    }),
  ],
  transformers: [transformerDirectives(), transformerVariantGroup()],
  safelist: ['aspect-video', 'aspect-portrait', 'aspect-square'],
  rules: [],
  shortcuts: [
    {
      screen: 'w-screen h-screen',
      full: 'w-full h-full',
      'max-full': 'max-w-full max-h-full',
      'max-screen': 'max-w-screen max-h-screen',
    },
    [/^ofade-([\d]*)$/, ([, c]) => `transition-opacity duration-${c}`],
  ],
  extractors: [extractorSvelte],
}
