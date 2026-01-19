/** @type {import('next').NextConfig} */
const internalApiUrl = (process.env.INTERNAL_API_URL || 'http://backend:12322').replace(/\/$/, '');

const nextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${internalApiUrl}/api/:path*` },
      { source: '/v1/:path*', destination: `${internalApiUrl}/v1/:path*` },
      { source: '/health', destination: `${internalApiUrl}/health` },
    ];
  },
};

module.exports = nextConfig;
