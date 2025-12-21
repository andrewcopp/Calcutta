export function mixHex(hexA: string, hexB: string, amountB: number) {
  const clamp = (n: number) => Math.max(0, Math.min(255, Math.round(n)));
  const norm = (hex: string) => hex.replace('#', '');
  const a = norm(hexA);
  const b = norm(hexB);
  const ar = parseInt(a.substring(0, 2), 16);
  const ag = parseInt(a.substring(2, 4), 16);
  const ab = parseInt(a.substring(4, 6), 16);
  const br = parseInt(b.substring(0, 2), 16);
  const bg = parseInt(b.substring(2, 4), 16);
  const bb = parseInt(b.substring(4, 6), 16);
  const r = clamp(ar * (1 - amountB) + br * amountB);
  const g = clamp(ag * (1 - amountB) + bg * amountB);
  const bl = clamp(ab * (1 - amountB) + bb * amountB);
  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${bl.toString(16).padStart(2, '0')}`;
}

export function desaturateHex(hex: string, amountWhite = 0.55) {
  return mixHex(hex, '#FFFFFF', amountWhite);
}
