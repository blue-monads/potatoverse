

const KEY = "_tnx_atoken_";


export const saveAccessToken = (accessToken: string) => {
    localStorage.setItem(KEY, accessToken);
}

export const getAccessToken = () => {
    return localStorage.getItem(KEY);
}

export const removeAccessToken = () => {
    localStorage.removeItem(KEY);
}


