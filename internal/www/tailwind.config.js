const colors = require('tailwindcss/colors')

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        './internal/www/assets/static/*.{html,js}',
        './internal/www/templates/*.html',
    ],
    theme: {
        colors: {
            transparent: 'transparent',
            white: "white",
            main: {  
                '50': '#ebf0ff',
                '100': '#dbe3ff',
                '200': '#becbff',
                '300': '#97a8ff',
                '400': '#6d78ff',
                '500': '#4c4bff',
                '600': '#3e2cff',
                '700': '#3320e2',
                '800': '#2a1db6',
                '900': '#27208e',
                '950': '#191353',
            },
            main_darker: "#1b2065",
            second: "#aea8ba",
            third: "#7a7485",
            fourth: "#800000",
            fifth: "#be2325"
        }

    }
    // ...
}