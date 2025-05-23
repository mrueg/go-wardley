module.exports = {
	output: "export",
	distDir: 'dist',
	  webpack: (
    config,
    { buildId, dev, isServer, defaultLoaders, nextRuntime, webpack }
  ) => {
    config.plugins.push(
    new webpack.optimize.LimitChunkCountPlugin({
      maxChunks: 1,
    }),
    )
    return config
  },
}
