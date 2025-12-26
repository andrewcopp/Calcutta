import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

export function HomePage() {
  const prefersReducedMotion = useMemo(() => {
    return (
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    );
  }, []);

  const [scrollY, setScrollY] = useState(0);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    if (prefersReducedMotion) return;

    const onScroll = () => {
      if (rafRef.current != null) return;
      rafRef.current = window.requestAnimationFrame(() => {
        rafRef.current = null;
        setScrollY(window.scrollY || 0);
      });
    };

    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', onScroll);
      if (rafRef.current != null) {
        window.cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
    };
  }, [prefersReducedMotion]);

  const [visibleSections, setVisibleSections] = useState<Record<string, boolean>>({
    hero: true,
    invest: false,
    own: false,
    earn: false,
    cta: false,
  });

  useEffect(() => {
    if (prefersReducedMotion) {
      setVisibleSections({ hero: true, invest: true, own: true, earn: true, cta: true });
      return;
    }

    const ids = ['invest', 'own', 'earn', 'cta'];
    const els = ids
      .map((id) => ({ id, el: document.getElementById(id) }))
      .filter((x): x is { id: string; el: HTMLElement } => Boolean(x.el));

    if (els.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        setVisibleSections((prev) => {
          const next = { ...prev };
          for (const entry of entries) {
            const id = (entry.target as HTMLElement).dataset.sectionId;
            if (!id) continue;
            if (entry.isIntersecting) next[id] = true;
          }
          return next;
        });
      },
      { threshold: 0.18 },
    );

    for (const { id, el } of els) {
      el.dataset.sectionId = id;
      observer.observe(el);
    }

    return () => observer.disconnect();
  }, [prefersReducedMotion]);

  const revealClass = (id: string) => {
    const isVisible = visibleSections[id];
    if (prefersReducedMotion) return '';
    return isVisible
      ? 'opacity-100 translate-y-0'
      : 'opacity-0 translate-y-6';
  };

  const darkTooltipContentStyle = useMemo(() => {
    return {
      background: 'rgba(17,24,39,0.92)',
      border: '1px solid rgba(255,255,255,0.12)',
      borderRadius: 12,
      color: 'rgba(255,255,255,0.92)',
      boxShadow: '0 14px 40px rgba(0,0,0,0.45)',
    } as React.CSSProperties;
  }, []);

  const darkTooltipLabelStyle = useMemo(() => {
    return { color: 'rgba(255,255,255,0.92)' } as React.CSSProperties;
  }, []);

  const darkTooltipItemStyle = useMemo(() => {
    return { color: 'rgba(255,255,255,0.92)' } as React.CSSProperties;
  }, []);

  const darkBarCursor = useMemo(() => {
    return { fill: 'rgba(255,255,255,0.06)' };
  }, []);

  const investData = [
    { label: 'Favorite', credits: 40, fill: '#2563EB' },
    { label: 'Contender', credits: 25, fill: '#16A34A' },
    { label: 'Value', credits: 20, fill: '#F59E0B' },
    { label: 'Dark Horse', credits: 15, fill: '#7C3AED' },
  ];

  const ownershipScenarios = useMemo(
    () => [
      {
        key: 'favorite-small',
        archetype: 'Favorite',
        subtitle: 'Top seed (1–3), heavily owned',
        seed: 1,
        credits: 5,
        totalCredits: 250,
        fill: '#2563EB',
      },
      {
        key: 'contender-balanced',
        archetype: 'Contender',
        subtitle: 'Strong seed, moderately owned',
        seed: 6,
        credits: 15,
        totalCredits: 120,
        fill: '#16A34A',
      },
      {
        key: 'value-medium',
        archetype: 'Value',
        subtitle: 'Middle seed, lightly owned',
        seed: 9,
        credits: 20,
        totalCredits: 80,
        fill: '#F59E0B',
      },
      {
        key: 'darkhorse-large',
        archetype: 'Dark Horse',
        subtitle: 'Lower seed (10–16), lightly owned',
        seed: 12,
        credits: 25,
        totalCredits: 55,
        fill: '#7C3AED',
      },
    ],
    [],
  );

  const earnBars = useMemo(
    () => [
      { key: 'favorite', label: 'Favorite', yourPoints: 21, fill: '#2563EB' },
      { key: 'contender', label: 'Contender', yourPoints: 62.5, fill: '#16A34A' },
      { key: 'value', label: 'Value', yourPoints: 38, fill: '#F59E0B' },
      { key: 'darkhorse', label: 'Dark Horse', yourPoints: 17.5, fill: '#7C3AED' },
    ],
    [],
  );

  return (
    <div
      className="min-h-screen text-white"
      style={{
        background:
          'linear-gradient(180deg, #070A12 0%, #091B3A 38%, #070A12 100%)',
      }}
    >
      <div className="relative overflow-hidden">
        <div
          className="pointer-events-none absolute -top-40 left-1/2 h-[680px] w-[680px] -translate-x-1/2 rounded-full blur-3xl"
          style={{
            background: 'radial-gradient(circle at 30% 30%, rgba(59,130,246,0.55), rgba(59,130,246,0) 60%)',
            transform: prefersReducedMotion
              ? undefined
              : `translate3d(-50%, ${scrollY * 0.08}px, 0)`,
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
            <div className={`transition-all duration-700 ease-out ${revealClass('hero')}`}>
              <div className="text-xs font-semibold tracking-wider text-blue-200">
                CALCUTTA INVESTMENT POOL
              </div>
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
                  to="/calcuttas/create"
                  className="inline-flex items-center justify-center rounded-full bg-white px-7 py-3 font-semibold text-gray-900 transition-colors hover:bg-white/90"
                >
                  Create a pool
                </Link>
                <Link
                  to="/calcuttas"
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

      <div className="container mx-auto px-4">
        <div className="max-w-5xl mx-auto">
          <section id="invest" className="py-20 sm:py-28">
            <div className={`transition-all duration-700 ease-out ${revealClass('invest')}`}>
              <div className="text-xs font-semibold tracking-wider text-blue-200">INVEST</div>
              <div className="mt-3 flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
                <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Allocate 100 credits</h2>
                <div className="text-sm text-white/70 max-w-md">
                  Favorite, Contender, Value, Dark Horse — your portfolio is your strategy.
                </div>
              </div>

              <div className="mt-10">
                <div className="h-72 sm:h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={investData} margin={{ top: 8, right: 20, left: 0, bottom: 10 }}>
                      <CartesianGrid stroke="rgba(255,255,255,0.12)" strokeDasharray="3 3" />
                      <XAxis dataKey="label" tick={{ fontSize: 12, fill: 'rgba(255,255,255,0.75)' }} axisLine={{ stroke: 'rgba(255,255,255,0.25)' }} />
                      <YAxis tick={{ fontSize: 12, fill: 'rgba(255,255,255,0.75)' }} axisLine={{ stroke: 'rgba(255,255,255,0.25)' }} />
                      <Tooltip
                        formatter={(value: number) => [`${value} credits`, 'Investment']}
                        contentStyle={darkTooltipContentStyle}
                        labelStyle={darkTooltipLabelStyle}
                        itemStyle={darkTooltipItemStyle}
                        cursor={darkBarCursor}
                      />
                      <Bar dataKey="credits" radius={[12, 12, 0, 0]}>
                        {investData.map((entry) => (
                          <Cell key={entry.label} fill={entry.fill} />
                        ))}
                      </Bar>
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </div>
          </section>

          <section id="own" className="py-20 sm:py-28">
            <div className={`transition-all duration-700 ease-out ${revealClass('own')}`}>
              <div className="text-xs font-semibold tracking-wider text-blue-200">OWN</div>
              <div className="mt-3 flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
                <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Your slice of each team</h2>
                <div className="text-sm text-white/70 max-w-md">
                  Four different paths: favorites dilute, value and dark horses can be bigger shares.
                </div>
              </div>

              <div className="mt-10">
                <div className="home-own-marquee">
                  <div
                    className={`home-own-track ${prefersReducedMotion ? 'home-own-track--no-anim' : ''}`}
                    aria-label="Ownership examples"
                  >
                    {[...ownershipScenarios, ...ownershipScenarios].map((s, idx) => {
                      const yourPct = (s.credits / s.totalCredits) * 100;
                      const restPct = 100 - yourPct;

                      return (
                        <div
                          key={`${s.key}-${idx}`}
                          className="home-own-card rounded-3xl bg-white/5 ring-1 ring-white/10 backdrop-blur px-6 py-6"
                        >
                          <div className="flex items-start justify-between gap-3">
                            <div className="h-[56px] flex flex-col justify-center overflow-hidden">
                              <div className="text-lg font-semibold leading-snug truncate">{s.archetype}</div>
                              <div className="text-sm text-white/70 leading-snug truncate">{s.subtitle}</div>
                            </div>
                            <div className="text-xs font-semibold text-white/80 rounded-full bg-white/10 ring-1 ring-white/10 px-3 py-1">
                              Seed {s.seed}
                            </div>
                          </div>

                          <div className="mt-5 h-56">
                            <ResponsiveContainer width="100%" height="100%">
                              <PieChart>
                                <Pie
                                  data={[
                                    { name: 'Your share', value: yourPct, fill: s.fill },
                                    { name: 'Everyone else', value: restPct, fill: 'rgba(255,255,255,0.18)' },
                                  ]}
                                  cx="50%"
                                  cy="50%"
                                  innerRadius={54}
                                  outerRadius={92}
                                  paddingAngle={2}
                                  dataKey="value"
                                  isAnimationActive={false}
                                >
                                  <Cell fill={s.fill} />
                                  <Cell fill="rgba(255,255,255,0.18)" />
                                </Pie>
                                <Tooltip
                                  formatter={(value: number) => [`${value.toFixed(2)}%`, 'Ownership']}
                                  contentStyle={darkTooltipContentStyle}
                                  labelStyle={darkTooltipLabelStyle}
                                  itemStyle={darkTooltipItemStyle}
                                />
                              </PieChart>
                            </ResponsiveContainer>
                          </div>

                          <div className="mt-5 grid grid-cols-3 gap-3">
                            <div className="rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3">
                              <div className="text-xs text-white/60">You</div>
                              <div className="mt-1 text-lg font-bold">{s.credits}</div>
                            </div>
                            <div className="rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3">
                              <div className="text-xs text-white/60">Total</div>
                              <div className="mt-1 text-lg font-bold">{s.totalCredits}</div>
                            </div>
                            <div className="rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3">
                              <div className="text-xs text-white/60">Your %</div>
                              <div className="mt-1 text-lg font-bold" style={{ color: s.fill }}>
                                {yourPct.toFixed(2)}%
                              </div>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section id="earn" className="py-20 sm:py-28">
            <div className={`transition-all duration-700 ease-out ${revealClass('earn')}`}>
              <div className="text-xs font-semibold tracking-wider text-blue-200">EARN</div>
              <div className="mt-3 flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
                <h2 className="text-4xl sm:text-5xl font-bold tracking-tight">Multiple paths to points</h2>
                <div className="text-sm text-white/70 max-w-md">
                  Four archetypes can contribute in different ways — a small slice of a favorite, or a bigger slice of the right dark horse.
                </div>
              </div>

              <div className="mt-10">
                <div className="h-72 sm:h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={earnBars} margin={{ top: 8, right: 20, left: 0, bottom: 12 }}>
                      <CartesianGrid stroke="rgba(255,255,255,0.12)" strokeDasharray="3 3" />
                      <XAxis dataKey="label" tick={{ fontSize: 12, fill: 'rgba(255,255,255,0.75)' }} axisLine={{ stroke: 'rgba(255,255,255,0.25)' }} />
                      <YAxis tick={{ fontSize: 12, fill: 'rgba(255,255,255,0.75)' }} axisLine={{ stroke: 'rgba(255,255,255,0.25)' }} />
                      <Tooltip
                        formatter={(value: number) => [`${value.toFixed(1)} pts`, 'Contribution']}
                        contentStyle={darkTooltipContentStyle}
                        labelStyle={darkTooltipLabelStyle}
                        itemStyle={darkTooltipItemStyle}
                        cursor={darkBarCursor}
                      />
                      <Bar dataKey="yourPoints" radius={[12, 12, 0, 0]}>
                        {earnBars.map((b) => (
                          <Cell key={b.key} fill={b.fill} />
                        ))}
                      </Bar>
                    </BarChart>
                  </ResponsiveContainer>
                </div>

                <div className="mt-6 flex justify-center">
                  <Link to="/rules" className="text-white/80 hover:text-white underline underline-offset-4">
                    See the full scoring + strategy walkthrough
                  </Link>
                </div>
              </div>
            </div>
          </section>

          <section id="cta" className="pt-10 pb-20 sm:pb-28">
            <div className={`transition-all duration-700 ease-out ${revealClass('cta')}`}>
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
                  This is a friendly game. Calcutta tracks points and ownership for your group — it does not facilitate gambling or real-money winnings.
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}