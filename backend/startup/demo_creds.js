window.onload = function() {
    console.log('demo creds');
    // if page is /zz/pages/auth/login
    if (window.location.pathname === '/zz/pages/auth/login') {
        // find the input field with name 'username'
        const usernameInput = document.querySelector('input[name="username"]');
        if (usernameInput) {
            usernameInput.value = '{username}';
        }
        // find the input field with name 'password'
        const passwordInput = document.querySelector('input[name="password"]');
        if (passwordInput) {
            passwordInput.value = '{password}';
        }
    }
}