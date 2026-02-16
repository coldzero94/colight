interface ColightLogoProps {
  size?: number;
  className?: string;
}

export function ColightLogo({ size = 28, className }: ColightLogoProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="colight-bg" x1="6" y1="4" x2="35" y2="36" gradientUnits="userSpaceOnUse">
          <stop stopColor="#7C8CFF" />
          <stop offset="0.52" stopColor="#4B8DFF" />
          <stop offset="1" stopColor="#23CFE8" />
        </linearGradient>
        <linearGradient id="colight-flow" x1="10" y1="30" x2="20" y2="9" gradientUnits="userSpaceOnUse">
          <stop stopColor="#E7F5FF" stopOpacity="0.9" />
          <stop offset="1" stopColor="#FFFFFF" />
        </linearGradient>
      </defs>

      <rect x="2" y="2" width="36" height="36" rx="11" fill="url(#colight-bg)" />
      <rect x="2" y="2" width="36" height="36" rx="11" fill="black" fillOpacity="0.14" />

      <path
        d="M10 28.5C12 22 15.2 17.8 20 14.2"
        stroke="url(#colight-flow)"
        strokeWidth="2.8"
        strokeLinecap="round"
      />
      <path
        d="M30 28.5C28 22 24.8 17.8 20 14.2"
        stroke="url(#colight-flow)"
        strokeWidth="2.8"
        strokeLinecap="round"
      />
      <path
        d="M20 14.2V9.2"
        stroke="white"
        strokeWidth="2.8"
        strokeLinecap="round"
      />

      <circle cx="10" cy="28.5" r="2.2" fill="#DDF2FF" />
      <circle cx="30" cy="28.5" r="2.2" fill="#DDF2FF" />
      <circle cx="20" cy="14.2" r="2.4" fill="white" />

      <path d="M20 6.2V4.8" stroke="white" strokeWidth="1.6" strokeLinecap="round" />
      <path d="M23 8.2L24 7.3" stroke="white" strokeWidth="1.6" strokeLinecap="round" />
      <path d="M17 8.2L16 7.3" stroke="white" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  );
}
