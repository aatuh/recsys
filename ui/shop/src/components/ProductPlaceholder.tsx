type Props = {
  seed: string;
  label?: string;
  className?: string;
};

function hashCode(input: string): number {
  let h = 0;
  for (let i = 0; i < input.length; i++) {
    h = (h << 5) - h + input.charCodeAt(i);
    h |= 0;
  }
  return Math.abs(h);
}

export default function ProductPlaceholder({ seed, label, className }: Props) {
  const h = hashCode(seed);
  const hue = h % 360;
  const hue2 = (h * 7) % 360;
  const shape = h % 3; // 0 circle, 1 rect, 2 triangle
  const bg = `hsl(${hue}, 45%, 18%)`;
  const fg = `hsl(${hue2}, 70%, 55%)`;
  const fg2 = `hsl(${(hue2 + 60) % 360}, 70%, 50%)`;

  const text = (label || seed).slice(0, 12);

  return (
    <svg
      viewBox="0 0 400 300"
      className={className}
      role="img"
      aria-label={label || "placeholder"}
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect width="400" height="300" fill={bg} />
      {shape === 0 ? (
        <circle cx="200" cy="150" r="90" fill={fg} />
      ) : shape === 1 ? (
        <rect x="110" y="80" width="180" height="140" rx="16" fill={fg} />
      ) : (
        <polygon points="200,60 320,240 80,240" fill={fg} />
      )}
      <circle cx="290" cy="70" r="26" fill={fg2} opacity="0.9" />
      <text
        x="200"
        y="270"
        textAnchor="middle"
        fill="white"
        fontSize="24"
        fontWeight="600"
        style={{ opacity: 0.9 }}
      >
        {text}
      </text>
    </svg>
  );
}
