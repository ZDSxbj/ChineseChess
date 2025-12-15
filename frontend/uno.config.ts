import {
  defineConfig,
  presetAttributify,
  presetUno,
  presetWind3,
  presetIcons,
} from 'unocss'

export default defineConfig({
  presets: [
    presetAttributify(),
    presetWind3(),
    presetUno(),
    // 允许使用 i-carbon-* 等图标类
    presetIcons({
      scale: 1.1,
      extraProperties: {
        display: 'inline-block',
        'vertical-align': 'middle',
      },
    }),
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
