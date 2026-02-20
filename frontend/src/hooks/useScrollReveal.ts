import { useEffect, useState } from 'react';

/**
 * Uses an IntersectionObserver to reveal sections as they scroll into view.
 *
 * Each section element must have an `id` matching one of the provided
 * `sectionIds`. Once a section becomes visible it stays visible
 * (one-way reveal).
 *
 * @param sectionIds  - DOM element ids to observe.
 * @param options     - Configuration options.
 * @param options.threshold     - IntersectionObserver threshold (default 0.18).
 * @param options.enabled       - When false, all sections are immediately
 *   marked visible (used for prefers-reduced-motion).
 * @param options.initiallyVisible - Section ids that start visible even
 *   before they intersect (e.g. the hero section).
 * @returns A record mapping each section id to a boolean visibility flag
 *   and a helper that returns Tailwind transition classes.
 */
export function useScrollReveal(
  sectionIds: readonly string[],
  options: {
    threshold?: number;
    enabled?: boolean;
    initiallyVisible?: readonly string[];
  } = {},
): {
  visibleSections: Record<string, boolean>;
  revealClass: (id: string) => string;
} {
  const { threshold = 0.18, enabled = true, initiallyVisible = [] } = options;

  const [visibleSections, setVisibleSections] = useState<Record<string, boolean>>(() => {
    const init: Record<string, boolean> = {};
    for (const id of sectionIds) {
      init[id] = initiallyVisible.includes(id);
    }
    return init;
  });

  useEffect(() => {
    if (!enabled) {
      const all: Record<string, boolean> = {};
      for (const id of sectionIds) {
        all[id] = true;
      }
      setVisibleSections(all);
      return;
    }

    const els = sectionIds
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
      { threshold },
    );

    for (const { id, el } of els) {
      el.dataset.sectionId = id;
      observer.observe(el);
    }

    return () => observer.disconnect();
  }, [enabled, sectionIds, threshold]);

  const revealClass = (id: string): string => {
    const isVisible = visibleSections[id];
    if (!enabled) return '';
    return isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-6';
  };

  return { visibleSections, revealClass };
}
