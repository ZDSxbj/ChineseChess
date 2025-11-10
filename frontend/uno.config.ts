import {
  defineConfig,
  presetAttributify,
  presetUno,
  presetWind3,
} from 'unocss'

export default defineConfig({
  presets: [
    presetAttributify(),
    presetWind3(),
    presetUno(),
  ],
  rules: [
    [/^scrollbar-thin$/, () => ({
      'scrollbar-width': 'thin',
      '&::-webkit-scrollbar': {
        width: '6px',
      },
    })],
    [/^scrollbar-track-(.+)$/, ([, c]) => ({
      '&::-webkit-scrollbar-track': {
        'background-color': `rgb(var(--un-${c}))`,
      },
    })],
    [/^scrollbar-thumb-(.+)$/, ([, c]) => ({
      '&::-webkit-scrollbar-thumb': {
        'background-color': `rgb(var(--un-${c}))`,
        'border-radius': '3px',
      },
    })],
  ],
})
