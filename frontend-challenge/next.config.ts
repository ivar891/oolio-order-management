import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  devIndicators: false,
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'orderfoodonline.deno.dev',
        pathname: '/public/images/**',
      },
    ],
  },
};

export default nextConfig;
