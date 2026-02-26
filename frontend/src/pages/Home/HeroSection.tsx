import { Link } from 'react-router-dom';

interface HeroSectionProps {
  scrollY: number;
  prefersReducedMotion: boolean;
  revealClass: string;
}

export function HeroSection({ scrollY, prefersReducedMotion, revealClass }: HeroSectionProps) {
  return (
    <div className="relative overflow-hidden">
      <div
        className="pointer-events-none absolute -top-40 left-1/2 h-[680px] w-[680px] -translate-x-1/2 rounded-full blur-3xl"
        style={{
          background: 'radial-gradient(circle at 30% 30%, rgba(59,130,246,0.55), rgba(59,130,246,0) 60%)',
          transform: prefersReducedMotion ? undefined : `translate3d(-50%, ${scrollY * 0.08}px, 0)`,
        }}
      />
      <div
        className="pointer-events-none absolute -top-24 right-[-180px] h-[520px] w-[520px] rounded-full blur-3xl"
        style={{
          background: 'radial-gradient(circle at 50% 50%, rgba(124,58,237,0.55), rgba(124,58,237,0) 60%)',
          transform: prefersReducedMotion ? undefined : `translate3d(0, ${scrollY * 0.06}px, 0)`,
        }}
      />

      <div className="container mx-auto px-4 pt-10 pb-8">
        <div className="max-w-5xl mx-auto">
          <div className="flex justify-end">
            <Link
              to="/login"
              className="inline-flex items-center justify-center rounded-full bg-white/10 px-5 py-2 text-sm font-semibold text-white ring-1 ring-white/20 backdrop-blur transition-colors hover:bg-white/15"
            >
              Already in a pool? Log in
            </Link>
          </div>
          <div className={`transition-all duration-700 ease-out ${revealClass}`}>
            <div className="text-xs font-semibold tracking-wider text-blue-200">MARCH MARKETS</div>
            <h1 className="mt-4 text-5xl sm:text-6xl font-bold tracking-tight">
              Invest.
              <span className="text-blue-300"> Own.</span>
              <span className="text-purple-200"> Earn.</span>
            </h1>
            <p className="mt-6 text-lg sm:text-xl text-white/80 max-w-2xl">
              A bracket alternative where you build a portfolio of teams and score based on what you own.
            </p>

            <div className="mt-9 flex flex-col sm:flex-row gap-3">
              <Link
                to="/pools/create"
                className="inline-flex items-center justify-center rounded-full bg-white px-7 py-3 font-semibold text-gray-900 transition-colors hover:bg-white/90"
              >
                Create a pool
              </Link>
              <Link
                to="/pools"
                className="inline-flex items-center justify-center rounded-full bg-white/10 px-7 py-3 font-semibold text-white ring-1 ring-white/20 backdrop-blur transition-colors hover:bg-white/15"
              >
                Join a pool
              </Link>
            </div>

            <div className="mt-6 flex items-center gap-4 text-sm text-white/70">
              <a href="#invest" className="hover:text-white">
                Invest
              </a>
              <a href="#own" className="hover:text-white">
                Own
              </a>
              <a href="#earn" className="hover:text-white">
                Earn
              </a>
              <Link to="/rules" className="text-white/70 hover:text-white">
                How it works →
              </Link>
            </div>
          </div>

          <div className="mt-14 h-px w-full bg-gradient-to-r from-transparent via-white/20 to-transparent" />
        </div>
      </div>
    </div>
  );
}
