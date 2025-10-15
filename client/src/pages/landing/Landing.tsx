import { FaqWithImage } from './FaqWithImage'
import { Hero } from './Hero'

function Landing() {
  return (
    <div className="flex flex-col items-center justify-center min-w-screen min-h-screen max-w-screen w-screen bg-gray-900">
      <div className="min-w-screen px-[5vw]! min-h-screen">
        <Hero />
      </div>

      <div className="min-w-screen px-[5vw]! my-30!">
        <FaqWithImage />
      </div>
    </div>
  )
}

export default Landing
