const nextConfig = {
  images: {
    dangerouslyAllowSVG: true,
    remotePatterns: [
      {
        protocol: "https",
        hostname: "static.wikia.nocookie.net",
        pathname: "/little-alchemy/**",
      },
    ],
  },
  eslint: {
    // heheheheheheh
    ignoreDuringBuilds: true,
  },
};

export default nextConfig;
