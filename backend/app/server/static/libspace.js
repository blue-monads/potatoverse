console.log("libspace.js/start");

const spaceRedirrectToAuth = (redirectBackUrl, actualPage) => {
    const prePageUrl = new URL('/zz/pages/auth/space/in_space/pre_page', window.location.origin);
    prePageUrl.searchParams.set('redirect_back_url', redirectBackUrl);
    prePageUrl.searchParams.set('actual_page', actualPage);
    window.location.href = prePageUrl.toString();
}

const spaceGetToken = (key) => {
    return localStorage.getItem(`${key}_space_token`);
}

window.spaceRedirrectToAuth = spaceRedirrectToAuth;
window.spaceGetToken = spaceGetToken;

console.log("libspace.js/end");
