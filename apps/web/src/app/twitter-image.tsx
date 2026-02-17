import { ImageResponse } from "next/og";
import { SocialImageTemplate } from "./social-image-template";

export const alt = "Colight - AI resume coaching platform";
export const size = {
  width: 1200,
  height: 630,
};
export const contentType = "image/png";

export default function TwitterImage() {
  return new ImageResponse(<SocialImageTemplate />, {
    ...size,
  });
}
