

const KEY = "_tnx_login_info_";


export const saveLoginData = (accessToken: string, userInfo: any) => {
     localStorage.setItem(KEY, JSON.stringify({ accessToken, userInfo }));
}

export const getLoginData = () => {
    const item = localStorage.getItem(KEY);
    if (!item) return null;
    return JSON.parse(item);
}

export const removeLoginData = () => {
    localStorage.removeItem(KEY);
}


