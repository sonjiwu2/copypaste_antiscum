const OUTLINE = '#111827'
const SKIN = '#ffc29b'
const HAIR = '#21191a'

export function HeroCharacter() {
  return (
    <svg
      className='pixel-rig pixel-rig--hero'
      viewBox='0 0 360 230'
      shapeRendering='crispEdges'
      aria-hidden='true'
    >
      <g className='rig__hero-bob'>
        <g className='rig__shield'>
          <path
            d='M59 66h20V57h42v8h20v10h12v70h-8v24h-12v17h-17v14H94v-9H77v-15H65v-24H56V76h3z'
            fill={OUTLINE}
          />
          <path
            d='M66 73h18v-8h32v8h18v8h10v62h-8v23h-10v13h-18v10H96v-8H83v-13H73v-22h-7z'
            fill='#1d6bc2'
          />
          <path d='M76 84h57v57h-7v23h-12v12H96v-9H84v-22h-8z' fill='#2e8be3' />
          <path d='M85 112h14v14h10v-10h19v11h-9v10h-10v9H98v-8H90v-9h-5z' fill='#76f0bd' />
          <path className='rig__shield-glow' d='M75 76h61v101H75z' fill='#8ff7ff' opacity='0' />
        </g>

        <g className='rig__hero-person'>
          <path
            d='M162 105h18V95h67v10h16v18h9v63h-30v39h-25v-38h-21v38h-26v-39h-21v-63h13z'
            fill={OUTLINE}
          />
          <path d='M169 109h16v-8h56v9h15v18h8v50h-29v-8h-57v8h-22v-50h13z' fill='#f4f6f8' />
          <path d='M170 124h22v44h-28v-38h6zM229 112h21v56h-28v-43h7z' fill='#dce3e9' />
          <path d='M190 119h5v31h-5zM216 119h5v31h-5z' fill='#237cd7' />
          <path d='M177 178h25v39h-25zM224 178h25v39h-25z' fill='#1d293b' />
          <path d='M172 215h36v10h-36zM220 215h36v10h-36z' fill='#fff' />
          <path d='M176 219h29v6h-29zM225 219h28v6h-28z' fill='#2182df' />

          <g className='rig__hero-head'>
            <path
              d='M164 31h13V19h66v7h15v10h10v58h-10v14h-18v10h-51v-7h-17V99h-12V48h4z'
              fill={OUTLINE}
            />
            <path d='M173 48h77v8h9v37h-9v10h-17v8h-43v-6h-14V94h-8V57h5z' fill={SKIN} />
            <path
              d='M178 51h17v8h17v-9h31v8h10v13h-15V61h-15v17h-11V63h-14v13h-14v-9h-12V54h6z'
              fill={HAIR}
            />
            <g className='rig__eyes'>
              <path d='M191 74h9v17h-9zM228 74h9v17h-9z' fill={OUTLINE} />
            </g>
            <path d='M209 96h21v5h-6v5h-11v-5h-4z' fill='#cf4b3d' />
            <path d='M167 34h9V21h62v5h15v9h9v23H158V44h9z' fill={OUTLINE} />
            <path d='M175 34v-7h58v5h14v8h8v10h-90v-9h10z' fill='#1975d5' />
            <path d='M184 27h45v6h-45zM169 42h75v6h-75z' fill='#49a6ff' />
            <path d='M199 27h17v4h4v14h-5v4h-16v-4h-5V33h5z' fill='#fff' />
            <path d='M203 32h9v4h4v5h-4v4h-9v-4h-4v-5h4z' fill='#2a82de' />
          </g>
        </g>

        <g className='rig__dog'>
          <g className='rig__dog-tail'>
            <path d='M327 165h13v-11h11v7h7v28h-9v10h-21z' fill={OUTLINE} />
            <path d='M333 170h10v-9h5v6h5v17h-8v8h-12z' fill='#c97931' />
            <path d='M346 184h7v8h-7z' fill='#fff' />
          </g>
          <path d='M257 148h65v9h14v48h-10v16h-23v-10h-28v10h-23v-17h-8v-38h13z' fill={OUTLINE} />
          <path d='M262 155h55v8h12v36h-9v13h-14v-10h-35v10h-14v-13h-7v-28h12z' fill='#c87932' />
          <path d='M274 160h35v42h-35z' fill='#f4f2e6' />
          <path
            d='M260 102h16V85h14v13h25V85h14v18h10v57h-10v13h-17v9h-35v-8h-17v-14h-10v-48h10z'
            fill={OUTLINE}
          />
          <g className='rig__dog-ear rig__dog-ear--left'>
            <path d='M265 103V91h9v-13h12v33z' fill='#c87932' />
            <path d='M273 96V87h6v14z' fill='#ef9a7d' />
          </g>
          <g className='rig__dog-ear rig__dog-ear--right'>
            <path d='M311 109V78h12v13h9v15z' fill='#c87932' />
            <path d='M317 101V87h6v9z' fill='#ef9a7d' />
          </g>
          <path d='M260 110h69v48h-9v12h-17v8h-26v-8h-17v-13h-7v-39h7z' fill='#c87932' />
          <path d='M273 119h42v39h-7v13h-28v-8h-7z' fill='#fff8ea' />
          <g className='rig__dog-eyes'>
            <path d='M270 126h8v12h-8zM310 126h8v12h-8z' fill={OUTLINE} />
          </g>
          <path d='M288 139h13v9h-13z' fill={OUTLINE} />
          <path d='M282 151h26v6h-5v8h-15v-8h-6z' fill='#e24d46' />
          <path d='M272 175h45v8h-45z' fill='#148bd2' />
          <path d='M291 176h9v9h-9z' fill='#63e7ca' />
        </g>
      </g>
    </svg>
  )
}
