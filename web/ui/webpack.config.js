const path = require('path');

module.exports = {
  entry: './src/main.jsx',
  output: {
    path: path.resolve(__dirname, '../dist'),
    filename: 'bundle.js',
    clean: false, // Don't wipe the Go backend or assets if any
  },
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: [
              '@babel/preset-env',
              ['@babel/preset-react', { runtime: 'automatic' }]
            ]
          }
        }
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader'] // Loads the CSS dynamically
      }
    ]
  },
  resolve: {
    extensions: ['.js', '.jsx']
  },
  devServer: {
    port: 5173,
    proxy: [
        {
          context: ['/api'],
          target: 'http://127.0.0.1:3000',
        },
      ]
  }
};
