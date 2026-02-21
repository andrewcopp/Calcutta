import { useMemo } from 'react';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { useScrollY } from '../hooks/useScrollY';
import { useScrollReveal } from '../hooks/useScrollReveal';
import { HeroSection } from './Home/HeroSection';
import { InvestSection } from './Home/InvestSection';
import { OwnSection } from './Home/OwnSection';
import { EarnSection } from './Home/EarnSection';
import { CtaSection } from './Home/CtaSection';

export function HomePage() {
  const prefersReducedMotion = useReducedMotion();
  const scrollY = useScrollY(!prefersReducedMotion);

  const sectionIds = useMemo(() => ['hero', 'invest', 'own', 'earn', 'cta'] as const, []);
  const { revealClass } = useScrollReveal(sectionIds, {
    enabled: !prefersReducedMotion,
    initiallyVisible: ['hero'],
  });

  return (
    <div
      className="min-h-screen text-white"
      style={{
        background:
          'linear-gradient(180deg, #070A12 0%, #091B3A 38%, #070A12 100%)',
      }}
    >
      <HeroSection
        scrollY={scrollY}
        prefersReducedMotion={prefersReducedMotion}
        revealClass={revealClass('hero')}
      />

      <div className="container mx-auto px-4">
        <div className="max-w-5xl mx-auto">
          <InvestSection revealClass={revealClass('invest')} />
          <OwnSection prefersReducedMotion={prefersReducedMotion} revealClass={revealClass('own')} />
          <EarnSection revealClass={revealClass('earn')} />
          <CtaSection revealClass={revealClass('cta')} />
        </div>
      </div>
    </div>
  );
}
