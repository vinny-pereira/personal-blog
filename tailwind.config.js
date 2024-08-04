/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.html"],
  darkMode: 'selector',
  theme: {
    fontFamily: {
        body: ['Roboto', 'Montserrat', 'Open Sans']
    },
    extend: {
     transitionProperty: {
        'visibility': 'height, width, opacity'
      }
    },
  },
  plugins: [],
}

