export default {
  plugins: {
    autoprefixer: {
      overrideBrowserslist: [
        '> 1%',
        'last 2 versions',
        'not dead',
        'not IE 11',
        'Chrome >= 90',
        'Firefox >= 88',
        'Edge >= 90',
        'Safari >= 14',
        'iOS >= 14',
        'Android >= 90'
      ]
    }
  }
}
