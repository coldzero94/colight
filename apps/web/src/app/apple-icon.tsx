import { ImageResponse } from "next/og";

export const size = { width: 180, height: 180 };
export const contentType = "image/png";

export default function AppleIcon() {
  return new ImageResponse(
    (
      <div
        style={{
          width: 180,
          height: 180,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background:
            "linear-gradient(145deg, rgb(124,140,255), rgb(75,141,255) 55%, rgb(35,207,232))",
          borderRadius: 40,
          position: "relative",
        }}
      >
        <div
          style={{
            position: "absolute",
            inset: 0,
            borderRadius: 40,
            background: "rgba(0,0,0,0.14)",
          }}
        />
        <svg
          width="110"
          height="110"
          viewBox="0 0 40 40"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          style={{ position: "relative" }}
        >
          <path
            d="M10 28.5C12 22 15.2 17.8 20 14.2"
            stroke="rgba(231,245,255,0.95)"
            strokeWidth="2.8"
            strokeLinecap="round"
          />
          <path
            d="M30 28.5C28 22 24.8 17.8 20 14.2"
            stroke="rgba(231,245,255,0.95)"
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
      </div>
    ),
    { ...size },
  );
}
