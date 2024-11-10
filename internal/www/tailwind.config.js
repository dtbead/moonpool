/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        './internal/www/assets/static/*.{html,js}',
        './internal/www/templates/*.html',
    ],
    theme: {
        colors: {
            transparent: 'transparent',
            main: "#27208e",
            second: "#aea8ba",
            third: "#7a7485",
            fourth: "#800000",
            fifth: "#be2325"
        }

    }
    // ...
}