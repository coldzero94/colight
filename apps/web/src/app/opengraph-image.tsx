import { ImageResponse } from "next/og";
import { SocialImageTemplate } from "./social-image-template";

export const alt = "Colight - Together We Light";
export const size = {
  width: 1200,
  height: 630,
};
export const contentType = "image/png";

export default function OpenGraphImage() {
  return new ImageResponse(<SocialImageTemplate />, {
    ...size,
  });
}
