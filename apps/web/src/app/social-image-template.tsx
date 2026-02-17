function ColightMark() {
  return (
    <svg
      width="192"
      height="192"
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id="og-bg" x1="6" y1="4" x2="35" y2="36" gradientUnits="userSpaceOnUse">
          <stop stopColor="#7C8CFF" />
          <stop offset="0.52" stopColor="#4B8DFF" />
          <stop offset="1" stopColor="#23CFE8" />
        </linearGradient>
        <linearGradient id="og-flow" x1="10" y1="30" x2="20" y2="9" gradientUnits="userSpaceOnUse">
          <stop stopColor="#E7F5FF" stopOpacity="0.9" />
          <stop offset="1" stopColor="#FFFFFF" />
        </linearGradient>
      </defs>

      <rect x="2" y="2" width="36" height="36" rx="11" fill="url(#og-bg)" />
      <rect x="2" y="2" width="36" height="36" rx="11" fill="black" fillOpacity="0.14" />

      <path
        d="M10 28.5C12 22 15.2 17.8 20 14.2"
        stroke="url(#og-flow)"
        strokeWidth="2.8"
        strokeLinecap="round"
      />
      <path
        d="M30 28.5C28 22 24.8 17.8 20 14.2"
        stroke="url(#og-flow)"
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

export function SocialImageTemplate() {
  return (
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        position: "relative",
        overflow: "hidden",
        background:
          "radial-gradient(circle at 14% 20%, rgba(76,141,255,0.35), rgba(76,141,255,0) 38%), radial-gradient(circle at 86% 24%, rgba(35,207,232,0.28), rgba(35,207,232,0) 35%), linear-gradient(145deg, #070a13 0%, #090f1f 55%, #06142a 100%)",
        color: "#F4F8FF",
        fontFamily: "Inter, system-ui, sans-serif",
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: "36px",
          borderRadius: "28px",
          border: "1px solid rgba(255,255,255,0.14)",
          background:
            "linear-gradient(160deg, rgba(255,255,255,0.08), rgba(255,255,255,0.02))",
        }}
      />

      <div
        style={{
          position: "absolute",
          right: "-70px",
          bottom: "-70px",
          width: "360px",
          height: "360px",
          borderRadius: "999px",
          background: "radial-gradient(circle, rgba(76,141,255,0.32), rgba(76,141,255,0) 68%)",
        }}
      />

      <div
        style={{
          position: "relative",
          width: "100%",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "84px 90px",
          gap: "48px",
        }}
      >
        <div style={{ display: "flex", flexDirection: "column", flex: 1 }}>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              borderRadius: "999px",
              border: "1px solid rgba(123,164,255,0.5)",
              background:
                "linear-gradient(130deg, rgba(124,140,255,0.26), rgba(35,207,232,0.14))",
              padding: "10px 18px",
              fontSize: "22px",
              fontWeight: 700,
              letterSpacing: "0.18em",
            }}
          >
            COLIGHT
          </div>

          <div
            style={{
              marginTop: "22px",
              fontSize: "64px",
              lineHeight: 1.04,
              letterSpacing: "-0.03em",
              fontWeight: 800,
            }}
          >
            Together We Light
          </div>

          <div
            style={{
              marginTop: "18px",
              maxWidth: "640px",
              fontSize: "30px",
              lineHeight: 1.35,
              color: "rgba(236,243,255,0.9)",
            }}
          >
            AI-powered resume coaching for stronger stories, better interviews,
            and career growth.
          </div>

          <div
            style={{
              marginTop: "34px",
              display: "flex",
              alignItems: "center",
              fontSize: "25px",
              color: "rgba(195,222,255,0.9)",
              letterSpacing: "0.03em",
            }}
          >
            colight.vercel.app
          </div>
        </div>

        <div
          style={{
            display: "flex",
            width: "250px",
            height: "250px",
            borderRadius: "52px",
            justifyContent: "center",
            alignItems: "center",
            border: "1px solid rgba(255,255,255,0.18)",
            background:
              "linear-gradient(155deg, rgba(255,255,255,0.16), rgba(255,255,255,0.04))",
            boxShadow:
              "0 24px 54px rgba(0,0,0,0.32), inset 0 1px 0 rgba(255,255,255,0.24)",
          }}
        >
          <ColightMark />
        </div>
      </div>
    </div>
  );
}
