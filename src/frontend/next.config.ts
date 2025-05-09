const nextConfig = {
    images: {
      dangerouslyAllowSVG: true,
      remotePatterns: [
        {
          protocol: 'https',
          hostname: 'static.wikia.nocookie.net',
          pathname: '/little-alchemy/**',
        },
      ],
    },
  };
  
  export default nextConfig;