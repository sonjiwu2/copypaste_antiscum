const OUTLINE = '#111827'
const SKIN = '#ffc29b'
const SKIN_SHADE = '#f28b63'
const HAIR = '#21191a'

export function BuyerCharacter() {
  return (
    <svg
      className='pixel-rig pixel-rig--buyer'
      viewBox='0 12 180 188'
      shapeRendering='crispEdges'
      aria-hidden='true'
    >
      <g className='rig__bob'>
        <path d='M36 105h18V94h69v10h17v17h12v64H29v-64h7z' fill={OUTLINE} />
        <path d='M42 109h17V101h58v8h16v19h10v49H37v-50h5z' fill='#14885e' />
        <path d='M48 115h25v57H39v-42h9zM105 110h21v62h-28v-49h7z' fill='#39af7d' />
        <path d='M69 105h34v17H69z' fill='#53c893' />
        <path d='M70 123h5v29h-5zM95 123h5v29h-5z' fill='#f4fff9' />
        <path d='M41 154h17v9H47v12H36v-15h5z' fill='#087452' />

        <g className='rig__head rig__head--buyer'>
          <path
            d='M43 32h11V20h66v6h15v11h10v56h-9v15h-18v11H67v-7H49v-12H38V49h5z'
            fill={OUTLINE}
          />
          <path d='M51 47h77v8h8v37h-8v11h-17v8H68v-7H54V94h-8V56h5z' fill={SKIN} />
          <path d='M51 70h8v17h-8zM126 70h8v17h-8z' fill={SKIN_SHADE} />
          <path
            d='M57 50h16v7h18v-8h31v8h10v13h-15V60h-15v16H91V62H77v13H63v-9H51V53h6z'
            fill={HAIR}
          />
          <g className='rig__eyes'>
            <path d='M69 73h9v16h-9zM106 73h9v16h-9z' fill={OUTLINE} />
          </g>
          <path d='M87 94h21v5h-6v5H91v-5h-4z' fill='#d95645' />
          <path d='M46 35h9V22h62v5h15v9h8v21H37V44h9z' fill={OUTLINE} />
          <path d='M53 35V27h59v5h14v8h7v9H43v-8h10z' fill='#1c9a6c' />
          <path d='M62 27h47v6H62zM48 41h75v6H48z' fill='#55c998' />
        </g>

        <g className='rig__phone-arm'>
          <path d='M116 118h15v-8h13v10h10v42h-9v13h-18v-10h-14v-31h3z' fill={OUTLINE} />
          <path d='M121 123h12v-7h7v8h7v33h-7v10h-10v-8h-10v-25h1z' fill={SKIN} />
          <path d='M130 86h31v5h6v62h-5v6h-36v-6h-5V97h5V91h4z' fill={OUTLINE} />
          <path d='M130 94h28v56h-28z' fill='#596170' />
          <path d='M134 98h20v7h-20z' fill='#aeb8c6' />
          <path className='rig__screen-glow' d='M136 112h16v25h-16z' fill='#81dff2' opacity='.12' />
          <path d='M139 143h10v3h-10z' fill='#dce4ec' />
          <path className='rig__tap' d='M118 133h18v7h-12v8h-8v-13h2z' fill={SKIN} />
        </g>
      </g>
    </svg>
  )
}
