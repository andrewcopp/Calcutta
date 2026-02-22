import { Link } from 'react-router-dom';

interface CtaSectionProps {
  revealClass: string;
}

export function CtaSection({ revealClass }: CtaSectionProps) {
  return (
    <section id="cta" className="pt-10 pb-20 sm:pb-28">
      <div className={`transition-all duration-700 ease-out ${revealClass}`}>
        <div className="flex flex-col items-center text-center">
          <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Start a pool with friends</h2>
          <div className="mt-7 flex flex-col sm:flex-row gap-3">
            <Link
              to="/calcuttas/create"
              className="inline-flex items-center justify-center rounded-full bg-white px-8 py-3 font-semibold text-gray-900 transition-colors hover:bg-white/90"
            >
              Create a pool
            </Link>
            <Link
              to="/calcuttas"
              className="inline-flex items-center justify-center rounded-full bg-white/10 px-8 py-3 font-semibold text-white ring-1 ring-white/20 backdrop-blur transition-colors hover:bg-white/15"
            >
              Join a pool
            </Link>
          </div>

          <div className="mt-8 max-w-2xl text-xs sm:text-sm text-white/60">
            Free to play. The only currency here is credits and the satisfaction of being right.
          </div>
        </div>
      </div>
    </section>
  );
}
