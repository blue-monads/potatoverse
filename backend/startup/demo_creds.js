(function () {
    var USERNAME = '{username}';
    var PASSWORD = '{password}';

    // React tracks input values internally, so assigning `el.value` is ignored by
    // onChange and gets reverted on the next render. Go through the native setter
    // and fire the event React actually subscribes to.
    function setReactValue(el, value) {
        var setter = Object.getOwnPropertyDescriptor(
            window.HTMLInputElement.prototype,
            'value'
        ).set;
        setter.call(el, value);
        el.dispatchEvent(new Event('input', { bubbles: true }));
    }

    // Pre-hydration the plain DOM value sticks too, so read back what React
    // rendered with instead: that only matches once the state update landed.
    function reactAccepted(el, value) {
        for (var key in el) {
            if (key.indexOf('__reactProps$') === 0) {
                return el[key].value === value;
            }
        }
        return false;
    }

    function fill() {
        if (window.location.pathname !== '/zz/pages/auth/login') {
            return false;
        }

        var usernameInput = document.querySelector('input[name="username"]');
        var passwordInput = document.querySelector('input[name="password"]');
        if (!usernameInput || !passwordInput) {
            return false;
        }

        setReactValue(usernameInput, USERNAME);
        setReactValue(passwordInput, PASSWORD);

        return reactAccepted(usernameInput, USERNAME) && reactAccepted(passwordInput, PASSWORD);
    }

    var attempts = 0;
    var timer = setInterval(function () {
        if (fill() || ++attempts > 50) {
            clearInterval(timer);
        }
    }, 100);

    fill();
})();
