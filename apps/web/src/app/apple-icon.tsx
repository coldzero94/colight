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
          background: "#00D191",
          borderRadius: 40,
        }}
      >
        <svg
          width="110"
          height="110"
          viewBox="0 0 110 110"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M75 30a30 30 0 1 0 0 50"
            stroke="white"
            strokeWidth="12"
            strokeLinecap="round"
          />
        </svg>
      </div>
    ),
    { ...size },
  );
}
