import { defineComponent, h } from 'vue'

type PathDef = { d: string; fillRule?: string; clipRule?: string }

function createSvgIcon(name: string, paths: PathDef[]) {
  return defineComponent({
    name,
    setup() {
      return () =>
        h(
          'svg',
          {
            width: '1em',
            height: '1em',
            viewBox: '0 0 24 24',
            fill: 'currentColor',
            xmlns: 'http://www.w3.org/2000/svg',
          },
          paths.map((p, idx) =>
            h('path', {
              key: idx,
              d: p.d,
              fill: 'currentColor',
              fillRule: p.fillRule,
              clipRule: p.clipRule,
            }),
          ),
        )
    },
  })
}

export const Menu = createSvgIcon('MenuIcon', [
  { d: 'M3 6h18v2H3V6z' },
  { d: 'M3 11h18v2H3v-2z' },
  { d: 'M3 16h18v2H3v-2z' },
])

export const Collection = createSvgIcon('CollectionIcon', [
  { d: 'M6 4h12a2 2 0 0 1 2 2v2H4V6a2 2 0 0 1 2-2z' },
  { d: 'M4 9h16v9a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9z' },
])

export const Document = createSvgIcon('DocumentIcon', [
  { d: 'M6 2h8l4 4v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2z' },
  { d: 'M14 2v6h6' },
  { d: 'M8 12h8v2H8v-2z' },
  { d: 'M8 16h6v2H8v-2z' },
])

export const Expand = createSvgIcon('ExpandIcon', [
  { d: 'M5 5h6v2H7v4H5V5z' },
  { d: 'M19 5v6h-2V7h-4V5h6z' },
  { d: 'M19 19h-6v-2h4v-4h2v6z' },
  { d: 'M5 19v-6h2v4h4v2H5z' },
])

export const Fold = createSvgIcon('FoldIcon', [
  { d: 'M5 5h14v2H5V5z' },
  { d: 'M7 9h10v2H7V9z' },
  { d: 'M11 13h2v6h-2v-6z' },
])

export default { Menu, Collection, Document, Expand, Fold }