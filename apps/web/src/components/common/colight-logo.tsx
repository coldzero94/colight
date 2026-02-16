interface ColightLogoProps {
  size?: number;
  className?: string;
}

export function ColightLogo({ size = 28, className }: ColightLogoProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      <rect width="32" height="32" rx="8" fill="currentColor" className="text-primary" />
      <path
        d="M21 11a7.5 7.5 0 1 0 0 10"
        stroke="white"
        strokeWidth="3"
        strokeLinecap="round"
      />
    </svg>
  );
}
