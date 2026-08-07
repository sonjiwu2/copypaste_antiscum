const OUTLINE = '#111827'
const SKIN = '#ffc29b'
const SKIN_SHADE = '#f28b63'
const HAIR = '#21191a'

export function SellerCharacter() {
  return (
    <svg
      className='pixel-rig pixel-rig--seller'
      viewBox='0 12 190 188'
      shapeRendering='crispEdges'
      aria-hidden='true'
    >
      <g className='rig__bob'>
        <path d='M29 107h20V95h68v10h17v16h11v65H23v-64h6z' fill={OUTLINE} />
        <path d='M36 110h17v-8h58v9h16v17h10v50H31v-50h5z' fill='#d83d31' />
        <path d='M41 117h29v55H31v-42h10zM101 111h19v63H94v-49h7z' fill='#ef5544' />
        <path d='M66 105h34v17H66z' fill='#ff6b55' />
        <path d='M68 123h5v29h-5zM93 123h5v29h-5z' fill='#fff7f3' />

        <g className='rig__head rig__head--seller'>
          <path
            d='M38 33h12V21h65v6h15v10h10v55h-9v15h-18v11H62v-7H45V99H33V49h5z'
            fill={OUTLINE}
          />
          <path d='M46 48h77v8h8v36h-8v10h-17v8H64v-6H51V94h-9V57h4z' fill={SKIN} />
          <path d='M47 70h8v17h-8zM121 70h8v17h-8z' fill={SKIN_SHADE} />
          <path
            d='M51 51h16v7h18v-8h31v8h10v13h-15V61H96v16H85V63H72v13H58v-9H46V54h5z'
            fill={HAIR}
          />
          <g className='rig__eyes'>
            <path d='M65 74h9v16h-9zM102 74h9v16h-9z' fill={OUTLINE} />
          </g>
          <path d='M83 95h21v5h-6v5H87v-5h-4z' fill='#c63e35' />
          <path d='M41 35h8V23h62v5h14v8h9v22H32V45h9z' fill={OUTLINE} />
          <path d='M49 35v-7h57v5h14v8h7v9H39v-8h10z' fill='#e94336' />
          <path d='M58 28h45v6H58zM43 42h72v6H43z' fill='#ff6b55' />
          <path d='M79 28h16v4h4v13h-5v4H79v-4h-5V33h5z' fill='#fff' />
          <path d='M82 32h9v4h4v5h-4v4h-9v-4h-4v-5h4z' fill='#ef5544' />
        </g>

        <g className='rig__parcel'>
          <path d='M88 111h84v7h7v62h-7v7H88v-7H76v-56h12z' fill={OUTLINE} />
          <path d='M93 118h72v7h7v48h-7v7H93z' fill='#c87927' />
          <path d='M95 121h33v52H95z' fill='#b66520' />
          <path d='M130 121h35v52h-35z' fill='#e59a3e' />
          <path d='M126 119h7v59h-7z' fill='#714018' />
          <path d='M143 137h18v22h-18z' fill='#f8f5eb' />
          <path d='M147 142h10v4h-10zM147 150h8v4h-8z' fill='#8b94a2' />
          <path className='rig__label-glow' d='M157 122h10v12h-10z' fill='#2792ef' />
          <path d='M157 124h6v6h-6z' fill='#9bd8ff' />
          <path d='M78 142h20v8H86v16H75v-20h3z' fill={SKIN} />
        </g>
      </g>
    </svg>
  )
}
