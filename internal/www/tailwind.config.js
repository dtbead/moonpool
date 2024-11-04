/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        './internal/www/assets/static/*.{html,js}',
        './internal/www/templates/*.html',
    ],
    theme: {
        colors: {
            transparent: 'transparent',
            primary: "#27208e"
        }

    }
    // ...
}